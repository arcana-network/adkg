package dacss

import (
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

type PssTestNode2 struct {
	testutils.PssTestNode
	newTransport *MockTransport
	messageCount int
}

func (n *PssTestNode2) Transport() *MockTransport {
	return n.newTransport
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

func (n *PssTestNode2) GetPublicKeyFor(idx int, fromNewCommittee bool) curves.Point {
	nodes := n.Nodes(fromNewCommittee)
	for _, n := range nodes {
		if n.Index == idx {
			pk, err := testutils.TestCurve().NewIdentityPoint().Set(&n.PubKey.X, &n.PubKey.Y)
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
		selectedNodes = n.Transport().nodesNew
	} else {
		selectedNodes = n.Transport().nodesOld
	}

	nodes := make(map[common.NodeDetailsID]common.NodeDetails, len(selectedNodes))
	for _, node := range selectedNodes {
		nodes[node.Details().GetNodeDetailsID()] = node.Details()
	}

	return nodes
}

func NewEmptyNode(index int, keypair common.KeyPair, Transport *MockTransport, isFaulty, isNewCommittee bool) *PssTestNode2 {
	pssTestNode := testutils.NewEmptyNode(index, keypair, nil, isFaulty, isNewCommittee)
	node := PssTestNode2{
		PssTestNode:  *pssTestNode,
		newTransport: Transport,
		messageCount: 0,
	}

	return &node
}

func (node *PssTestNode2) Broadcast(toNewCommittee bool, msg common.PSSMessage) {
	node.newTransport.Broadcast(toNewCommittee, node.Details(), msg)
}

func (node *PssTestNode2) Send(receiver common.NodeDetails, msg common.PSSMessage) error {
	node.newTransport.Send(node.Details(), receiver, msg)
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
	node.newTransport.receivedMessages = append(node.newTransport.receivedMessages, PssMessage) // Save the message
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
	case dacss.DacssCommitmentMessageType:
		processDACSSMessage[*dacss.DacssCommitmentMessage](PssMessage.Data, sender, node, dacss.DacssCommitmentMessageType)

	default:
		log.Infof("No handler found. MsgType=%s", PssMessage.Type)
	}

}

func (t *MockTransport) GetSentMessages() []common.PSSMessage {
	return t.sentMessages
}

func (t *MockTransport) GetReceivedMessages() []common.PSSMessage {
	return t.receivedMessages
}

func (t *MockTransport) GetBroadcastedMessages() []common.PSSMessage {
	return t.broadcastedMessages
}

func (node *PssTestNode2) CountReceivedMessages(msgType string) int {
	receivedMessages := node.Transport().receivedMessages
	filteredMessages := make([]common.PSSMessage, 0)

	for _, msg := range receivedMessages {
		if msg.Type == msgType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return len(filteredMessages)
}

func (node *PssTestNode2) GetReceivedMessages(msgType string) []common.PSSMessage {
	receivedMessages := node.Transport().receivedMessages
	filteredMessages := make([]common.PSSMessage, 0)

	for _, msg := range receivedMessages {
		if msg.Type == msgType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return filteredMessages
}
