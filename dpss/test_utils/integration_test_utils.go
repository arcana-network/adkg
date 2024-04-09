package testutils

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

type IntegrationTestNode struct {
	PssTestNode
	NewTransport             *IntegrationMockTransport
	MessageCount             int
	ProcessMessagesInterface ProcessMessagesInterface
}

type ProcessMessagesInterface interface {
	ProcessMessages(sender common.NodeDetails, PssMessage common.PSSMessage, node *IntegrationTestNode)
}

func (n *IntegrationTestNode) Transport() *IntegrationMockTransport {
	return n.NewTransport
}

type IntegrationMockTransport struct {
	nodesOld            []*IntegrationTestNode
	nodesNew            []*IntegrationTestNode
	output              chan string
	broadcastedMessages []common.PSSMessage // Store messages that are broadcasted
	sentMessages        []common.PSSMessage
	ReceivedMessages    []common.PSSMessage
}

func NewMockTransport(nodesOld, nodesNew []*IntegrationTestNode) *IntegrationMockTransport {
	return &IntegrationMockTransport{nodesOld: nodesOld, nodesNew: nodesNew, output: make(chan string, 100)}
}

func (transport *IntegrationMockTransport) Init(nodesOld, nodesNew []*IntegrationTestNode) {
	transport.nodesOld = nodesOld
	transport.nodesNew = nodesNew
}

func (n *IntegrationTestNode) GetPublicKeyFor(idx int, fromNewCommittee bool) curves.Point {
	nodes := n.Nodes(fromNewCommittee)
	for _, n := range nodes {
		if n.Index == idx {
			pk, err := TestCurve().NewIdentityPoint().Set(&n.PubKey.X, &n.PubKey.Y)
			if err != nil {
				return nil
			}
			return pk
		}
	}
	return nil
}

func (n *IntegrationTestNode) Nodes(fromNewCommittee bool) map[common.NodeDetailsID]common.NodeDetails {
	var selectedNodes []*IntegrationTestNode
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

func NewIntegrationTestNode(index int, keypair common.KeyPair, transport *IntegrationMockTransport, isFaulty, isNewCommittee bool, processMessagesInterface ProcessMessagesInterface) *IntegrationTestNode {
	pssTestNode := NewEmptyNode(index, keypair, nil, isFaulty, isNewCommittee)
	node := IntegrationTestNode{
		PssTestNode:              *pssTestNode,
		NewTransport:             transport,
		MessageCount:             0,
		ProcessMessagesInterface: processMessagesInterface,
	}

	return &node
}

func (node *IntegrationTestNode) ReceiveMessage(sender common.NodeDetails, PssMessage common.PSSMessage) {
	node.NewTransport.ReceivedMessages = append(node.NewTransport.ReceivedMessages, PssMessage) // Save the message
	node.MessageCount = node.MessageCount + 1
	node.ProcessMessagesInterface.ProcessMessages(sender, PssMessage, node)
}

func (node *IntegrationTestNode) Broadcast(toNewCommittee bool, msg common.PSSMessage) {
	node.NewTransport.Broadcast(toNewCommittee, node.Details(), msg)
}

func (node *IntegrationTestNode) Send(receiver common.NodeDetails, msg common.PSSMessage) error {
	node.NewTransport.Send(node.Details(), receiver, msg)
	return nil
}

func (t *IntegrationMockTransport) Send(sender, receiver common.NodeDetails, msg common.PSSMessage) {

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

func (t *IntegrationMockTransport) Broadcast(toNewCommittee bool, sender common.NodeDetails, m common.PSSMessage) {
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

type MessageProcessor interface {
	Process(sender common.NodeDetails, node common.PSSParticipant)
}

func ProcessDACSSMessage[T MessageProcessor](data []byte, sender common.NodeDetails, node common.PSSParticipant, messageType string) {
	log.Debugf("Got %s", messageType)
	var msg T
	err := bijson.Unmarshal(data, &msg)
	if err != nil {
		log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", messageType)
		return
	}
	msg.Process(sender, node)
}

func (t *IntegrationMockTransport) GetSentMessages() []common.PSSMessage {
	return t.sentMessages
}

func (t *IntegrationMockTransport) GetReceivedMessages() []common.PSSMessage {
	return t.ReceivedMessages
}

func (t *IntegrationMockTransport) GetBroadcastedMessages() []common.PSSMessage {
	return t.broadcastedMessages
}

func (node *IntegrationTestNode) CountReceivedMessages(msgType string) int {
	receivedMessages := node.Transport().ReceivedMessages
	filteredMessages := make([]common.PSSMessage, 0)

	for _, msg := range receivedMessages {
		if msg.Type == msgType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return len(filteredMessages)
}

func (node *IntegrationTestNode) GetReceivedMessages(msgType string) []common.PSSMessage {
	receivedMessages := node.Transport().ReceivedMessages
	filteredMessages := make([]common.PSSMessage, 0)

	for _, msg := range receivedMessages {
		if msg.Type == msgType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return filteredMessages
}
