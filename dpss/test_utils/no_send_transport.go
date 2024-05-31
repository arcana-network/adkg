package testutils

import (
	"sync"
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
	ReceivedMessages    []common.PSSMessage
	MsgReceivedSignal   chan struct{}
	MsgSentSignal       chan struct{}
	MsgBroadcastSignal  chan struct{}
	sync.Mutex
}

func NewNoSendMockTransport(nodesOld, nodesNew []*PssTestNode) *NoSendMockTransport {
	return &NoSendMockTransport{nodesNew: nodesNew, nodesOld: nodesOld,
		output:             make(chan string, 100),
		MsgReceivedSignal:  make(chan struct{}, 1001),
		MsgSentSignal:      make(chan struct{}, 1001),
		MsgBroadcastSignal: make(chan struct{}, 1001),
	}
}

func (transport *NoSendMockTransport) Init(nodesOld, nodesNew []*PssTestNode) {
	transport.nodesOld = nodesOld
	transport.nodesNew = nodesNew
}

// Registers a message was broadcast by a node
func (transport *NoSendMockTransport) Broadcast(sender common.NodeDetails, m common.PSSMessage) {
	transport.Lock()
	transport.BroadcastedMessages = append(transport.BroadcastedMessages, m)
	transport.Unlock()

}

// Registers a message was sent from one node to another
func (transport *NoSendMockTransport) Send(sender, receiver common.NodeDetails, msg common.PSSMessage) {
	transport.Lock()
	transport.sentMessages = append(transport.sentMessages, msg)
	transport.Unlock()
}

// returns the sent messages through this transport
func (transport *NoSendMockTransport) GetSentMessages() []common.PSSMessage {
	return transport.sentMessages
}

func (transport *NoSendMockTransport) CountSentMsg(msgType string) int {
	sentMessages := transport.GetSentMessages()
	filteredMessages := make([]common.PSSMessage, 0)

	for _, msg := range sentMessages {
		if msg.Type == msgType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	countSentMsg := len(filteredMessages)
	return countSentMsg
}

func (transport *NoSendMockTransport) GetBroadcastedMessages() []common.PSSMessage {
	return transport.BroadcastedMessages
}

func (transport *NoSendMockTransport) AssertNoMsgsBroadcast(t *testing.T) {
	broadcastedMsgs := transport.BroadcastedMessages

	assert.Equal(t, 0, len(broadcastedMsgs))
}

func (transport *NoSendMockTransport) WaitForMessagesReceived(count int) {
	for i := 0; i < count; i++ {
		<-transport.MsgReceivedSignal
	}
}

func (transport *NoSendMockTransport) WaitForMessagesSent(count int) {
	for i := 0; i < count; i++ {
		<-transport.MsgSentSignal
	}
}

func (transport *NoSendMockTransport) WaitForBroadcastSent(count int) {
	for i := 0; i < count; i++ {
		<-transport.MsgBroadcastSignal
	}
}
