package dacss

import (
	"math/big"
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

type PssTestNode2 struct {
	// Index & PubKey of this node
	details             common.NodeDetails
	isNewCommittee      bool
	committeeTestParams common.CommitteeParams

	state       *common.PSSNodeState
	LongtermKey common.KeyPair // NOTE this key must coincide with the pubkey in the details
	isFaulty    bool

	Transport    *MockTransport
	messageCount int
	//shares of old/new committee
	//false: old, true: new
	shares map[bool]map[int64]*big.Int
}

type MockTransport struct {
	nodesOld            []*PssTestNode2
	nodesNew            []*PssTestNode2
	output              chan string
	broadcastedMessages []common.PSSMessage // Store messages that are broadcasted
	sentMessages        []common.PSSMessage
	receivedMessages    []common.PSSMessage
}

func NewMockTransport(nodesOld, nodesNew []*PssTestNode2) *MockTransport {
	return &MockTransport{nodesOld: nodesOld, nodesNew: nodesNew, output: make(chan string, 100)}
}

func (transport *MockTransport) Init(nodesOld, nodesNew []*PssTestNode2) {
	transport.nodesOld = nodesOld
	transport.nodesNew = nodesNew
}

func (n *PssTestNode2) State() *common.PSSNodeState {
	return n.state
}

func (n *PssTestNode2) ID() int {
	return n.details.Index
}

func (n *PssTestNode2) IsNewNode() bool {
	return !n.isNewCommittee
}
func (n *PssTestNode2) Details() common.NodeDetails {
	return n.details
}

func (n *PssTestNode2) PrivateKey() curves.Scalar {
	return n.LongtermKey.PrivateKey
}

func (n *PssTestNode2) GetPublicKeyFor(idx int, fromNewCommittee bool) curves.Point {
	nodes := n.Nodes(fromNewCommittee)
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

func (n *PssTestNode2) Nodes(fromNewCommittee bool) map[common.NodeDetailsID]common.NodeDetails {
	var selectedNodes []*PssTestNode2
	if fromNewCommittee {
		selectedNodes = n.Transport.nodesNew
	} else {
		selectedNodes = n.Transport.nodesOld
	}

	nodes := make(map[common.NodeDetailsID]common.NodeDetails, len(selectedNodes))
	for _, node := range selectedNodes {
		nodes[node.Details().GetNodeDetailsID()] = node.details
	}

	return nodes
}

func GetSingleNode(isNewCommittee bool, isFaulty bool) (*PssTestNode2, *MockTransport) {
	nodesOld := []*PssTestNode2{}
	nodesNew := []*PssTestNode2{}
	keypair := common.GenerateKeyPair(curves.K256())
	transport := NewMockTransport(nodesOld, nodesNew)

	node := NewEmptyNode(1, keypair, transport, isFaulty, isNewCommittee)

	if isNewCommittee {
		transport.Init(nodesOld, []*PssTestNode2{node})
	} else {
		transport.Init([]*PssTestNode2{node}, nodesNew)
	}

	return node, transport
}

func NewEmptyNode(index int, keypair common.KeyPair, Transport *MockTransport, isFaulty, isNewCommittee bool) *PssTestNode2 {
	var params common.CommitteeParams
	if isNewCommittee {
		params = StandardNewCommitteeParams()
	} else {
		params = StandardOldCommitteeParams()
	}
	node := PssTestNode2{
		details:             common.NodeDetails{Index: index, PubKey: common.CurvePointToPoint(keypair.PublicKey, common.SECP256K1)},
		isNewCommittee:      isNewCommittee,
		committeeTestParams: params,
		state: &common.PSSNodeState{
			AcssStore:       &common.AcssStateMap{},
			DualAcssStarted: false,
		},
		Transport:   Transport,
		LongtermKey: keypair,
		isFaulty:    isFaulty,

		shares: make(map[bool]map[int64]*big.Int),
	}

	return &node
}

func (node *PssTestNode2) Broadcast(toNewCommittee bool, msg common.PSSMessage) {
	node.Transport.Broadcast(toNewCommittee, node.Details(), msg)
}
func (node *PssTestNode2) Params() (n int, k int, t int) {
	return node.committeeTestParams.N, node.committeeTestParams.K, node.committeeTestParams.T
}

func (node *PssTestNode2) Send(receiver common.NodeDetails, msg common.PSSMessage) error {
	node.Transport.Send(node.Details(), receiver, msg)
	return nil
}

func (t *MockTransport) Send(sender, receiver common.NodeDetails, msg common.PSSMessage) {

	t.sentMessages = append(t.sentMessages, msg) // Save the message
	flag := 0

	for _, n := range t.nodesOld {
		log.Debugf("msg=%s, sender=%d, receiver=%d, roundID=%s", msg.Type, n.ID(), receiver.Index, msg.PSSRoundDetails.PssID)
		if receiver.IsEqual(n.Details()) {
			flag = 1
			go n.ReceiveMessage(sender, msg)
			break
		}
	}

	for _, n := range t.nodesNew {
		if flag == 1 {
			break
		}

		log.Debugf("msg=%s, sender=%d, receiver=%d, round=%s", msg.Type, n.ID(), receiver.Index, msg.PSSRoundDetails.PssID)
		if receiver.IsEqual(n.Details()) {
			go n.ReceiveMessage(sender, msg)
			break
		}
	}
}

func (t *MockTransport) Broadcast(toNewCommittee bool, sender common.NodeDetails, m common.PSSMessage) {
	t.broadcastedMessages = append(t.broadcastedMessages, m) // Save the message

	if toNewCommittee {
		for _, p := range t.nodesNew {
			go func(node common.PSSParticipant) {
				node.ReceiveMessage(sender, m)
			}(p)
		}
	} else {
		for _, p := range t.nodesOld {
			go func(node common.PSSParticipant) {
				node.ReceiveMessage(sender, m)
			}(p)
		}
	}

}

func (node *PssTestNode2) ReceiveMessage(sender common.NodeDetails, PssMessage common.PSSMessage) {
	node.Transport.receivedMessages = append(node.Transport.receivedMessages, PssMessage) // Save the message
	node.messageCount = node.messageCount + 1
	switch {
	case strings.HasPrefix(PssMessage.Type, "dacss"):
		node.ProcessDACSSMessages(sender, PssMessage)

	default:
		log.Infof("No handler found. MsgType=%s", PssMessage.Type)
	}
}

type MessageProcessor interface {
	Process(sender common.NodeDetails, node common.PSSParticipant)
}

func processDACSSMessage[T MessageProcessor](data []byte, sender common.NodeDetails, node common.PSSParticipant, messageType string) {
	log.Debugf("Got %s", messageType)
	var msg T
	err := bijson.Unmarshal(data, &msg)
	if err != nil {
		log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", messageType)
		return
	}
	msg.Process(sender, node)
}

func (node *PssTestNode2) ProcessDACSSMessages(sender common.NodeDetails, PssMessage common.PSSMessage) {
	switch PssMessage.Type {
	case dacss.InitMessageType:
		processDACSSMessage[dacss.InitMessage](PssMessage.Data, sender, node, dacss.InitMessageType)
	case dacss.DacssEchoMessageType:
		processDACSSMessage[dacss.DacssEchoMessage](PssMessage.Data, sender, node, dacss.DacssEchoMessageType)
	case dacss.ShareMessageType:
		processDACSSMessage[dacss.DualCommitteeACSSShareMessage](PssMessage.Data, sender, node, dacss.ShareMessageType)
	case dacss.AcssProposeMessageType:
		processDACSSMessage[*dacss.AcssProposeMessage](PssMessage.Data, sender, node, dacss.AcssProposeMessageType)
	case dacss.AcssReadyMessageType:
		processDACSSMessage[*dacss.DacssReadyMessage](PssMessage.Data, sender, node, dacss.AcssReadyMessageType)
	case dacss.ImplicateExecuteMessageType:
		processDACSSMessage[*dacss.ImplicateExecuteMessage](PssMessage.Data, sender, node, dacss.ImplicateExecuteMessageType)
	case dacss.ImplicateReceiveMessageType:
		processDACSSMessage[*dacss.ImplicateReceiveMessage](PssMessage.Data, sender, node, dacss.ImplicateReceiveMessageType)
	case dacss.ShareRecoveryMessageType:
		processDACSSMessage[*dacss.ShareRecoveryMessage](PssMessage.Data, sender, node, dacss.ShareRecoveryMessageType)
	case dacss.ReceiveShareRecoveryMessageType:
		processDACSSMessage[*dacss.ReceiveShareRecoveryMessage](PssMessage.Data, sender, node, dacss.ReceiveShareRecoveryMessageType)
	case dacss.DacssOutputMessageType:
		processDACSSMessage[*dacss.DacssOutputMessage](PssMessage.Data, sender, node, dacss.DacssOutputMessageType)

	}

}

func (t *MockTransport) GetSentMessages() []common.PSSMessage {
	return t.sentMessages
}

func (t *MockTransport) GetReceivedMessages() []common.PSSMessage {
	return t.receivedMessages
}

func (node *PssTestNode2) CountReceivedMessages(msgType string) int {
	receivedMessages := node.Transport.receivedMessages
	filteredMessages := make([]common.PSSMessage, 0)

	for _, msg := range receivedMessages {
		if msg.Type == msgType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return len(filteredMessages)
}

func (node *PssTestNode2) GetReceivedMessages(msgType string) []common.PSSMessage {
	receivedMessages := node.Transport.receivedMessages
	filteredMessages := make([]common.PSSMessage, 0)

	for _, msg := range receivedMessages {
		if msg.Type == msgType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return filteredMessages
}
