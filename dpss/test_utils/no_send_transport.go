package testutils

import (
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/stretchr/testify/assert"
)

// NoSendMockTransport is a Mock Transport that only catches messages, doesn't send them
// It will only register what messages have been broadcasted, sent and received
// For UnitTest testpurposes this should be enough
// Note that if you're trying to test the actual sending of messages between nodes, you should use a different transport
type NoSendMockTransport struct {
	nodesOld            []*PssTestNode
	nodesNew            []*PssTestNode
	output              chan string
	BroadcastedMessages []common.PSSMessage
	sentMessages        []common.PSSMessage
	receivedMessages    []common.PSSMessage
}

func NewNoSendMockTransport(nodesOld, nodesNew []*PssTestNode) *NoSendMockTransport {
	return &NoSendMockTransport{nodesNew: nodesNew, nodesOld: nodesOld, output: make(chan string, 100)}
}

func (transport *NoSendMockTransport) Init(nodesOld, nodesNew []*PssTestNode) {
	transport.nodesNew = nodesOld
	transport.nodesOld = nodesNew
}

// Registers a message was broadcast by a node
func (transport *NoSendMockTransport) Broadcast(sender common.NodeDetails, m common.PSSMessage) {
	transport.BroadcastedMessages = append(transport.BroadcastedMessages, m)
}

// Registers a message was sent from one node to another
func (transport *NoSendMockTransport) Send(sender, receiver common.NodeDetails, msg common.PSSMessage) {
	transport.sentMessages = append(transport.sentMessages, msg)
}

// returns the sent messages through this transport
func (transport *NoSendMockTransport) GetSentMessages() []common.PSSMessage {
	return transport.sentMessages
}

func (transport *NoSendMockTransport) AssertNoMsgsBroadcast(t *testing.T) {
	broadcastedMsgs := transport.BroadcastedMessages

	assert.Equal(t, 0, len(broadcastedMsgs))
}
