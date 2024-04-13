package testutils

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

// Definition of a testNode meant for Integration testing
// In particular, this has a transport layer that is used to send messages between nodes
// additionally, it has a ProcessMessagesInterface to easily switch between definitions
// of how to process msgs
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

// Transport layer for Integration testing. Messages should be passed on, in addition to being stored
// what precise messages are passed on is defined in the functionality of ProcessMessagesInterface
type IntegrationMockTransport struct {
	nodesOld            []*IntegrationTestNode
	nodesNew            []*IntegrationTestNode
	output              chan string
	broadcastedMessages []common.PSSMessage // Store messages that are broadcasted
	sentMessages        []common.PSSMessage
	ReceivedMessages    []common.PSSMessage
}

func NewIntegrationMockTransport(nodesOld, nodesNew []*IntegrationTestNode) *IntegrationMockTransport {
	return &IntegrationMockTransport{nodesOld: nodesOld, nodesNew: nodesNew, output: make(chan string, 100)}
}

func (transport *IntegrationMockTransport) Init(nodesOld, nodesNew []*IntegrationTestNode) {
	transport.nodesOld = nodesOld
	transport.nodesNew = nodesNew
}

// Obtain public key for node with given index, from given committee
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

// Returns nodes from the given committee that are stored in the transport for this testNode
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

// Returns a new IntegrationTestNode
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

// Message is stored in ReceivedMessages and passed on to the ProcessMessagesInterface
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

func (t *IntegrationMockTransport) GetSentMessages() []common.PSSMessage {
	return t.sentMessages
}

func (t *IntegrationMockTransport) GetReceivedMessages() []common.PSSMessage {
	return t.ReceivedMessages
}

func (t *IntegrationMockTransport) GetBroadcastedMessages() []common.PSSMessage {
	return t.broadcastedMessages
}

func (t *IntegrationMockTransport) CountReceivedMessages(msgType string) int {
	filteredMessages := make([]common.PSSMessage, 0)

	for _, msg := range t.ReceivedMessages {
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

func (node *IntegrationTestNode) CountReceivedMessages(msgType string) int {
	return node.Transport().CountReceivedMessages(msgType)
}
