package dpss

import (
	"errors"
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/aba"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/keyset"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/new_committee"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/old_committee"
	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

// PSSNode represents a node participating in the DPSS protocol.
type PSSNode struct {
	PssNodeTransport *PssNodeTransport
	common.BaseNode
	state             *common.PSSNodeState
	OldCommitteeNodes common.NodeNetwork // Set of nodes belonging to the old committee.
	NewCommitteeNodes common.NodeNetwork // Set of nodes belonging to the new committee.
	NodeDetails       common.NodeDetails
}

// Creates a new PSSNode
func NewPSSNode(broker common.MessageBroker,
	nodeDetails common.NodeDetails,
	oldCommittee []common.NodeDetails,
	newCommittee []common.NodeDetails,
	bus eventbus.Bus,
	tOldCommittee int, kOldCommittee int,
	tNewCommittee int, kNewCommittee int,
	privateKey curves.Scalar,
	epoch int) (*PSSNode, error) {
	// TODO - check if it's ok to always set pss protocol prefix to dpss-1/
	// set pss protocol prefix to dpss-1/ since it will be used in both old & new committee
	transport := NewPssNodeTransport(bus, getPSSProtocolPrefix(), "dpss-transport")

	// Creates the committees
	oldCommitteeNetwork := common.NodeNetwork{
		N:     len(oldCommittee),
		T:     tOldCommittee,
		K:     kOldCommittee,
		Nodes: common.MapFromNodeList(oldCommittee),
	}
	newCommiteeNetwork := common.NodeNetwork{
		N:     len(newCommittee),
		T:     tNewCommittee,
		K:     kNewCommittee,
		Nodes: common.MapFromNodeList(newCommittee),
	}

	// Defines public key.
	g := curves.K256().NewGeneratorPoint()
	publicKey := g.Mul(privateKey)

	w := &common.Waiter{
		ThresholdCompletionWaiters: [](chan int){},
		TSetWaiters:                [](chan []int){},
	}
	// Creates the new node.
	newPSSNode := &PSSNode{
		PssNodeTransport: transport,
		state: &common.PSSNodeState{
			AcssStore:       &common.AcssStateMap{},
			ShareStore:      &common.PSSShareStore{},
			ABAStore:        &common.AbaStateMap{},
			PSSStore:        &common.PSSStateMap{Waiter: w},
			KeysetStore:     &common.KeysetStateMap{},
			BatchReconStore: &common.BatchRecStoreMap{},
			Waiter:          w,
		},
		BaseNode: common.NewBaseNode(
			&broker,
			nodeDetails,
			privateKey,
			publicKey,
		),
		OldCommitteeNodes: oldCommitteeNetwork,
		NewCommitteeNodes: newCommiteeNetwork,
		NodeDetails:       nodeDetails,
	}

	transport.Init()
	transport.SetPSSNode(newPSSNode)

	return newPSSNode, nil
}

// TODO check if and how this could be dynamic
// Returns the PSS protocol prefix, currently fixed to dpss-1
func getPSSProtocolPrefix() PSSProtocolPrefix {
	return PSSProtocolPrefix("dpss-1/")
}

// IsNewNode determines if the current node belongs to the new committee.
func (node *PSSNode) IsNewNode() bool {
	nodeDetails := node.NodeDetails
	_, found := node.NewCommitteeNodes.Nodes[nodeDetails.ToNodeDetailsID()]
	return found
}

func (node *PSSNode) OldNodeDetailsByID(idx int) (common.NodeDetails, error) {
	nodes := node.Nodes(false)
	for _, n := range nodes {
		if n.Index == idx {
			return n, nil
		}
	}
	return common.NodeDetails{}, errors.New("node not found in old committee")
}

// PublicKey returns the public key of the node with index idx that belongs to
// the old or new commitee according to the fromNewCommittee flag.
func (node *PSSNode) GetPublicKeyFor(idx int, fromNewCommittee bool) curves.Point {
	nodes := node.Nodes(fromNewCommittee)
	for _, n := range nodes {
		if n.Index == idx {
			pk, err := curves.K256().NewIdentityPoint().Set(&n.PubKey.X, &n.PubKey.Y)
			if err != nil {
				return nil
			}
			return pk
		}
	}
	return nil
}

// Returns the parameters for the network for the committee this node is part of.
// Parameters:
//   - n: number of nodes in the committee.
//   - k: number of corrupt nodes in the committee.
//   - t: the reconstruction threshold in that committee.
func (node *PSSNode) Params() (n, k, t int) {
	if !node.IsNewNode() {
		n = node.OldCommitteeNodes.N
		k = node.OldCommitteeNodes.K
		t = node.OldCommitteeNodes.T
	} else {
		n = node.NewCommitteeNodes.N
		k = node.NewCommitteeNodes.K
		t = node.NewCommitteeNodes.T
	}
	return
}

// Broadcast broadcasts a message to the given committee determined by the flag
// toNewCommittee.
func (node *PSSNode) Broadcast(toNewCommittee bool, msg common.PSSMessage) {
	nodesToBroadcast := node.Nodes(toNewCommittee)
	for _, n := range nodesToBroadcast {
		go func(receiver common.NodeDetails) {
			err := node.PssNodeTransport.Send(receiver, msg)
			if err != nil {
				log.WithField("Error", err).Error("Node.Broadcast()")
			}
		}(n)
	}
}
func (node *PSSNode) CurveParams(curveName string) (curves.Point, curves.Point) {
	return sharing.CurveParams(curveName)
}
func (n *PSSNode) StoreIndexToUser(index int, uid string, c common.CurveName) {
	// TODO: Impl?
	// Cant implement this, this needs to have uid, public key and appID also.
}
func (n *PSSNode) StoreShare(index int, share curves.Scalar, c common.CurveName) {
	si := share.BigInt()
	err := n.Broker().DBMethods().StoreCompletedPSSShare(*big.NewInt(int64(index)), *si, *si, c)
	if err != nil {
		log.WithError(err).Error("Node:StoreCompletedShare")
	}
}

func (n *PSSNode) DefaultBatchSize() int {
	return BATCHSIZE
}

// Nodes returns the set of nodes of the old or new committee according to the flag
// fromNewCommitte.
func (node *PSSNode) Nodes(fromNewCommittee bool) map[common.NodeDetailsID]common.NodeDetails {
	if fromNewCommittee {
		return node.NewCommitteeNodes.Nodes
	} else {
		return node.OldCommitteeNodes.Nodes
	}
}

type MessageProcessor interface {
	Process(sender common.NodeDetails, node common.PSSParticipant)
}

// General function to process messages of a given type:
// does the unmarshalling and calls the Process function of the message
func ProcessMessageForType[T MessageProcessor](data []byte, sender common.NodeDetails, node common.PSSParticipant, messageType string) {
	log.Debugf("Received %s", messageType)
	var msg T
	err := bijson.Unmarshal(data, &msg)
	if err != nil {
		log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", messageType)
		return
	}
	msg.Process(sender, node)
}

// ProcessMessage unmarshals the message and calls the appropriate handler for incoming message.
func (node *PSSNode) ProcessMessage(sender common.NodeDetails, message common.PSSMessage) error {
	switch message.Type {
	case dacss.InitMessageType:
		ProcessMessageForType[dacss.InitMessage](message.Data, sender, node, dacss.InitMessageType)
	case dacss.DacssEchoMessageType:
		ProcessMessageForType[dacss.DacssEchoMessage](message.Data, sender, node, dacss.DacssEchoMessageType)
	case dacss.ShareMessageType:
		ProcessMessageForType[dacss.DualCommitteeACSSShareMessage](message.Data, sender, node, dacss.ShareMessageType)
	case dacss.AcssProposeMessageType:
		ProcessMessageForType[*dacss.AcssProposeMessage](message.Data, sender, node, dacss.AcssProposeMessageType)
	case dacss.AcssReadyMessageType:
		ProcessMessageForType[*dacss.DacssReadyMessage](message.Data, sender, node, dacss.AcssReadyMessageType)
	case dacss.ImplicateExecuteMessageType:
		ProcessMessageForType[*dacss.ImplicateExecuteMessage](message.Data, sender, node, dacss.ImplicateExecuteMessageType)
	case dacss.ImplicateReceiveMessageType:
		ProcessMessageForType[*dacss.ImplicateReceiveMessage](message.Data, sender, node, dacss.ImplicateReceiveMessageType)
	case dacss.ShareRecoveryMessageType:
		ProcessMessageForType[*dacss.ShareRecoveryMessage](message.Data, sender, node, dacss.ShareRecoveryMessageType)
	case dacss.ReceiveShareRecoveryMessageType:
		ProcessMessageForType[*dacss.ReceiveShareRecoveryMessage](message.Data, sender, node, dacss.ReceiveShareRecoveryMessageType)
	case dacss.DacssOutputMessageType:
		ProcessMessageForType[*dacss.DacssOutputMessage](message.Data, sender, node, dacss.DacssOutputMessageType)
	case dacss.DacssCommitmentMessageType:
		ProcessMessageForType[*dacss.DacssCommitmentMessage](message.Data, sender, node, dacss.DacssCommitmentMessageType)
	case old_committee.DpssHimHandlerType:
		ProcessMessageForType[*old_committee.DpssHimMessage](message.Data, sender, node, old_committee.DpssHimHandlerType)
	case old_committee.InitRecHandlerType:
		ProcessMessageForType[*old_committee.InitRecMessage](message.Data, sender, node, old_committee.InitRecHandlerType)
	case old_committee.PreprocessBatchRecMessageType:
		ProcessMessageForType[*old_committee.PreprocessBatchRecMessage](message.Data, sender, node, old_committee.PreprocessBatchRecMessageType)
	case old_committee.PrivateRecMessageType:
		ProcessMessageForType[*old_committee.PrivateRecMsg](message.Data, sender, node, old_committee.PrivateRecMessageType)
	case old_committee.PublicRecMessageType:
		ProcessMessageForType[*old_committee.PublicRecMsg](message.Data, sender, node, old_committee.PublicRecMessageType)
	// mvba related messages
	case keyset.EchoMessageType:
		ProcessMessageForType[*keyset.EchoMessage](message.Data, sender, node, keyset.EchoMessageType)
	case keyset.ProposeMessageType:
		ProcessMessageForType[*keyset.ProposeMessage](message.Data, sender, node, keyset.ProposeMessageType)
	case keyset.ReadyMessageType:
		ProcessMessageForType[*keyset.ReadyMessage](message.Data, sender, node, keyset.ReadyMessageType)
	case keyset.OutputMessageType:
		ProcessMessageForType[*keyset.OutputMessage](message.Data, sender, node, keyset.OutputMessageType)
	case aba.Aux1MessageType:
		ProcessMessageForType[*aba.Aux1Message](message.Data, sender, node, aba.Aux1MessageType)
	case aba.Aux2MessageType:
		ProcessMessageForType[*aba.Aux2Message](message.Data, sender, node, aba.Aux2MessageType)
	case aba.AuxsetMessageType:
		ProcessMessageForType[*aba.AuxsetMessage](message.Data, sender, node, aba.AuxsetMessageType)
	case aba.CoinInitMessageType:
		ProcessMessageForType[*aba.CoinInitMessage](message.Data, sender, node, aba.CoinInitMessageType)
	case aba.CoinMessageType:
		ProcessMessageForType[*aba.CoinMessage](message.Data, sender, node, aba.CoinMessageType)
	case aba.Est1MessageType:
		ProcessMessageForType[*aba.Est1Message](message.Data, sender, node, aba.Est1MessageType)
	case aba.Est2MessageType:
		ProcessMessageForType[*aba.Est2Message](message.Data, sender, node, aba.Est2MessageType)
	case aba.InitMessageType:
		ProcessMessageForType[*aba.InitMessage](message.Data, sender, node, aba.InitMessageType)
	case new_committee.LocalComputationMessageType:
		ProcessMessageForType[*new_committee.LocalComputationMsg](message.Data, sender, node, new_committee.LocalComputationMessageType)
	default:
		log.Infof("No handler found. MsgType=%s", message.Type)
	}
	return nil
}

// ReceiveMessage passes on the message to the transport layer.
func (node *PSSNode) ReceiveMessage(sender common.NodeDetails, msg common.PSSMessage) {
	err := node.PssNodeTransport.Receive(sender, msg)
	if err != nil {
		log.WithError(err).Error("PSSNode:ReceiveMessage")
	}
}

// Details returns the details of the node, namely, its index and public key.
func (node *PSSNode) Details() common.NodeDetails {
	return node.NodeDetails
}

// Send sends a message to the node that has certain public key and index.
func (node *PSSNode) Send(n common.NodeDetails, msg common.PSSMessage) error {
	return node.PssNodeTransport.Send(n, msg)
}

// returns the state of the node.
func (node *PSSNode) State() *common.PSSNodeState {
	return node.state
}

// Returns the messageBroker
func (node *PSSNode) GetMessageBroker() *common.MessageBroker {
	return node.Broker()
}
