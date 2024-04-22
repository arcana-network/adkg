package dpss

import (
	"strconv"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dpss"
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
func NewPSSNode(broker common.MessageBroker, nodeDetails common.NodeDetails, oldCommittee []common.NodeDetails,
	newCommittee []common.NodeDetails, bus eventbus.Bus, tOldCommittee int, kOldCommittee int,
	tNewCommittee int, kNewCommittee int, privateKey curves.Scalar) (*PSSNode, error) {
	transport := NewPssNodeTransport(bus, getPSSProtocolPrefix(1), "dpss-transport")

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

	// Creates the new node.
	newPSSNode := &PSSNode{
		PssNodeTransport: transport,
		state: &common.PSSNodeState{
			AcssStore:       &common.AcssStateMap{},
			ShareStore:      &common.PSSShareStore{},
			BatchReconStore: &common.BatchRecStoreMap{},
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

// Returns the PSS protocol prefix in the form dpss-<epoch>
func getPSSProtocolPrefix(epoch int) PSSProtocolPrefix {
	return PSSProtocolPrefix("dpss" + "-" + strconv.Itoa(epoch) + "/")
}

// IsNewNode determines if the current node belongs to the new committee.
func (node *PSSNode) IsNewNode() bool {
	nodeDetails := node.NodeDetails
	_, found := node.NewCommitteeNodes.Nodes[nodeDetails.ToNodeDetailsID()]
	return found
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

// Params return the parameters for the network of the old or new committee
// according to the fromNewCommittee flag. It returns the parameters in a tuple
// with the following order:
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
	log.Debugf("Got %s", messageType)
	var msg T
	err := bijson.Unmarshal(data, &msg)
	if err != nil {
		log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", messageType)
		return
	}
	msg.Process(sender, node)
}

func (node *PSSNode) ProcessMessage(sender common.NodeDetails, message common.PSSMessage) error {

	switch message.Type {
	case dacss.InitMessageType:
		log.Info("AN INITMSG WAS RECEIVED")
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
	case dpss.DpssHimHandlerType:
		ProcessMessageForType[*dpss.DpssHimMessage](message.Data, sender, node, dpss.DpssHimHandlerType)
	case dpss.InitRecHandlerType:
		ProcessMessageForType[*dpss.InitRecMessage](message.Data, sender, node, dpss.InitRecHandlerType)
	case dpss.PreprocessBatchRecMessageType:
		ProcessMessageForType[*dpss.PreprocessBatchRecMessage](message.Data, sender, node, dpss.PreprocessBatchRecMessageType)
	case dpss.PrivateRecMessageType:
		ProcessMessageForType[*dpss.PrivateRecMsg](message.Data, sender, node, dpss.PrivateRecMessageType)
	case dpss.PublicRecMessageType:
		ProcessMessageForType[*dpss.PublicRecMsg](message.Data, sender, node, dpss.PublicRecMessageType)
	default:
		log.Infof("No handler found. MsgType=%s", message.Type)
	}
	return nil
}

func (node *PSSNode) ReceiveMessage(sender common.NodeDetails, msg common.PSSMessage) {
	err := node.PssNodeTransport.Receive(sender, msg)
	if err != nil {
		log.WithError(err).Error("PSSNode:ReceiveMessage")
	}
}

// FIXME: Implement this as long as we implement the DPSS protocol.
func (node *PSSNode) ProcessBroadcastMessage(message common.PSSMessage) error {
	return nil
}

// Details returns the details of the node, namely, its index and public key.
func (node *PSSNode) Details() common.NodeDetails {
	return node.NodeDetails
}

//TODO: Do we need this?
// func GenerateDPSSRoundID(rindex, noOfRandoms big.Int) common.PSSRoundID {
// 	index := strings.Join([]string{rindex.Text(16), noOfRandoms.Text(16)}, common.Delimiter2)
// 	return common.PSSRoundID(strings.Join([]string{"DPSS", index}, common.Delimiter3))
// }

// Send sends a message to the node that has certain public key and index.
func (node *PSSNode) Send(n common.NodeDetails, msg common.PSSMessage) error {
	return node.PssNodeTransport.Send(n, msg)
}

func (node *PSSNode) State() *common.PSSNodeState {
	return node.state
}

// TODO implement other functions of PSSParticipant interface
