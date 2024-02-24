package testutils

import (
	"github.com/arcana-network/dkgnode/common"
)

// NoSendMockTransport is a Mock Transport that only catches messages, doesn't send them
// It will only register what messages have been broadcasted, sent and received
// For UnitTest testpurposes this should be enough
// Note that if you're trying to test the actual sending of messages between nodes, you should use a different transport
type NoSendMockTransport struct {
	nodesOld            []*PssTestNode
	nodesNew            []*PssTestNode
	output              chan string
	broadcastedMessages []common.PSSMessage
	sentMessages        []common.PSSMessage
	receivedMessages    []common.PSSMessage
}

func NewNoSendMockTransport(nodesOld, nodesNew []*PssTestNode) *NoSendMockTransport {
	return &NoSendMockTransport{nodesNew: nodesNew, nodesOld: nodesOld, output: make(chan string, 100)}
}

func (t *NoSendMockTransport) Init(nodesOld, nodesNew []*PssTestNode) {
	t.nodesNew = nodesOld
	t.nodesOld = nodesNew
}

// Registers a message was broadcast by a node
func (t *NoSendMockTransport) Broadcast(sender common.NodeDetails, m common.PSSMessage) {
	t.broadcastedMessages = append(t.broadcastedMessages, m)
}

// Registers a message was sent from one node to another
func (t *NoSendMockTransport) Send(sender, receiver common.NodeDetails, msg common.PSSMessage) {
	t.sentMessages = append(t.sentMessages, msg)
}

// returns the sent messages through this transport
func (t *NoSendMockTransport) GetSentMessages() []common.PSSMessage {
	return t.sentMessages
}

// returns the broadcasted messages through this transport
func (t *NoSendMockTransport) GetBroadcastedMessages() []common.PSSMessage {
	return t.broadcastedMessages
}
