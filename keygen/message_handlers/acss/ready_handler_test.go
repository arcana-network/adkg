package acss

import (
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"

	"github.com/stretchr/testify/assert"
)

/*
Function: Process
Case: a node receives 1 `Ready` message
Expects:
- no messages are broadcast or sent
- keygen.State.ReceivedReady set to true for sender
- keygen.State.Phase not equal to ENDED
*/
func TestCorrectlyReceivedFirstReadyMsg(t *testing.T) {
	// This handler requires the same setup & data as the echo handler
	_, transport, node0, round, hash, encodedShares := setupEchoHandlerTest()

	msgToSend, _ := NewReadyMessage(round.ID(), encodedShares[node0.id-1], hash, common.CurveName(c.Name))
	
	defaultKeygen := NewDefaultKeygen(round)
	keygen, _ := node0.State().KeygenStore.GetOrSetIfNotComplete(round.ID(), defaultKeygen)	
	
	node0.ReceiveMessage(node0.Details(), *msgToSend)
	time.Sleep(1 * time.Second)

	// Checks
	// 1. no other message was sent or broadcast
	assert.True(t, node0.messageCount == 1) // (this is the initial message itself)
	assert.Equal(t, 0, countBroadcastedReadyMessages(transport), "No `Ready` messages should have been broadcasted")
	// 2. keygen.State.ReceivedReady set to true for sender
	assert.True(t, keygen.State.ReceivedReady[node0.id])
	// 3. keygen.State.Phase NOT equal to ENDED
	assert.NotEqual(t, common.Ended, keygen.State.Phase)
}

/*
Function: Process
Case: a node receives n `Ready` messages
Expects: 
- the node sends itself an `Output` message
- keygen.State.Phase has to be set to ENDED
- keygen.State.ReceivedReady set to true for sender
*/
func TestReadyMsgFromAllNodes(t *testing.T) {
	// (Node 3 is the dealer of this round and is the one that sends shares to all nodes)
	// Node 0 is the receiver of all the ready msg
	nodes, transport, node0, round, hash, encodedShares := setupEchoHandlerTest()
	defaultKeygen := NewDefaultKeygen(round)
	keygen, _ := node0.State().KeygenStore.GetOrSetIfNotComplete(round.ID(), defaultKeygen)	
	
	// Node0 receives n "ready" messages
	for _, nodeDetails := range node0.Nodes() {
		senderEcho := nodes[nodeDetails.Index-1]
		msgToSend, _ := NewReadyMessage(round.ID(), encodedShares[senderEcho.id-1], hash, common.CurveName(c.Name))
		node0.ReceiveMessage(senderEcho.Details(), *msgToSend)
	}
	time.Sleep(1 * time.Second)

	// Checks
	// 1. node0 should have been triggered to send itself an "output" message
	// (meaning total msg is n+1)
	assert.True(t, node0.messageCount == n+1)
	// 2. no `Ready` message should have been broadcasted
	assert.Equal(t, 0, countBroadcastedReadyMessages(transport), "No `Ready` messages should have been broadcasted")
	// 3. keygen.State.ReceivedReady set to true for sender
	assert.True(t, keygen.State.ReceivedReady[node0.id])
	// 4. keygen.State.Phase has to be set to ENDED
	assert.Equal(t, common.Ended, keygen.State.Phase)
}

/*
Function: Process
Case: a node receives 2*f+1 `Ready` messages
Expects: 
- the node sends itself an `Output` message
- keygen.State.Phase has to be set to ENDED
*/
func TestReceivedThresholdReadyMsgs(t *testing.T) {
	// (Node 3 is the dealer of this round and is the one that sends shares to all nodes)
	// Node 0 is the receiver of all the ready msg
	nodes, _, node0, round, hash, encodedShares := setupEchoHandlerTest()
	defaultKeygen := NewDefaultKeygen(round)
	keygen, _ := node0.State().KeygenStore.GetOrSetIfNotComplete(round.ID(), defaultKeygen)	
	
	// Node0 receives 2*f+1 "ready" messages
	for i :=0 ; i < (2*f+1) ; i++ {
		senderEcho := nodes[i]
		msgToSend, _ := NewReadyMessage(round.ID(), encodedShares[senderEcho.id-1], hash, common.CurveName(c.Name))
		node0.ReceiveMessage(senderEcho.Details(), *msgToSend)
	}
	time.Sleep(1 * time.Second)
	
	// Checks
	// 1. node0 should have been triggered to send itself an "output" message
	// (meaning total msg count is 2*f+1)
	assert.True(t, node0.messageCount == (2*f+2))
	// 2. keygen.State.Phase has to be set to ENDED
	assert.Equal(t, common.Ended, keygen.State.Phase)
}

/*
Function: Process
Case: 
- a node receives (f+1)-th `Ready` message - by setting RC to f manually
- has received f+1 echo messages - by setting c.EC manually
- hasn't broadcast `Ready` message yet
Expects: 
- ReadySent is set to true
- `Ready` message gets broadcasted
*/
func TestMustSendReady(t *testing.T) {
	// (Node 3 is the dealer of this round and is the one that sends shares to all nodes)
	// Node 0 is the receiver of all the ready msg
	// Node 1 is the sender of the (f+1)-th message
	nodes, transport, node0, round, hash, encodedShares := setupEchoHandlerTest()

	defaultKeygen := NewDefaultKeygen(round)
	keygen, _ := node0.State().KeygenStore.GetOrSetIfNotComplete(round.ID(), defaultKeygen)
	
	senderEcho := nodes[1]
	msgToSend, _ := NewReadyMessage(round.ID(), encodedShares[senderEcho.id-1], hash, common.CurveName(c.Name))
	msg := EchoMessage{
		round.ID(),
		EchoMessageType,
		common.CurveName(c.Name),
		encodedShares[senderEcho.id-1],
		hash,
	}
	cid := msg.Fingerprint()
	c := common.GetCStore(keygen, cid)
	// Set Ready Count for this message to f, so the next one passes the threshold
	c.RC = f
	// Make sure ReadySent is false
	c.ReadySent = false
	// Set Echo count for this message to f+1
	c.EC = f+1

	// Node0 should now broadcast a ReadyMessage
	node0.ReceiveMessage(senderEcho.Details(), *msgToSend)
	time.Sleep(1 * time.Second)
	
	// Checks
	// 1. node0 should have broadcast a Ready message
	assert.Equal(t, 1, countBroadcastedReadyMessages(transport))
	// 2. ReadySent must now be set to true
	assert.True(t, c.ReadySent)
	// 3. no `Output` message should have been sent
	// So messageCount should be 2 because of initial msg and broadcasted Ready msg
	assert.True(t, node0.messageCount == 2)
}

/*
NOTE
Replacing the check to broadcast a Ready message:
	if c.RC >= f+1 && !c.ReadySent && c.EC >= f+1 {
with this line will make the test run:
	if len(keygen.ReadyStore) >= f+1 && !c.ReadySent && c.EC >= f+1 {
*/
/*
Function: Process
Case: 
- a node receives (f+1)-th `Ready` message - by sending f+1 Ready msgs
- has received f+1 echo messages - by setting c.EC manually
- hasn't broadcast `Ready` message yet
Expects: 
- ReadySent is set to true
- `Ready` message gets broadcasted
*/
func TestMustSendReady2(t *testing.T) {
	// (Node 3 is the dealer of this round and is the one that sends shares to all nodes)
	// Node 0 is the receiver of all the ready msg
	// Node 1 is the sender of the (f+1)-th message
	nodes, transport, node0, round, hash, encodedShares := setupEchoHandlerTest()

	defaultKeygen := NewDefaultKeygen(round)
	keygen, _ := node0.State().KeygenStore.GetOrSetIfNotComplete(round.ID(), defaultKeygen)
	
	curve := common.CurveName(c.Name)
	senderReady1 := nodes[1]
	msgToSend1, _ := NewReadyMessage(round.ID(), encodedShares[senderReady1.id-1], hash, curve)
	msg1 := EchoMessage{
		round.ID(),
		EchoMessageType,
		common.CurveName(c.Name),
		encodedShares[senderReady1.id-1],
		hash,
	}
	cid := msg1.Fingerprint()
	c := common.GetCStore(keygen, cid)
	// Set Echo count for this message to f+1
	c.EC = f+1

	// prep another f Ready messages
	senderReady2 := nodes[2]
	msgToSend2, _ := NewReadyMessage(round.ID(), encodedShares[senderReady2.id-1], hash, curve)
	senderReady3 := nodes[3]
	msgToSend3, _ := NewReadyMessage(round.ID(), encodedShares[senderReady3.id-1], hash, curve)

	// Node0 receives f Ready messages
	node0.ReceiveMessage(senderReady2.Details(), *msgToSend2)
	node0.ReceiveMessage(senderReady3.Details(), *msgToSend3)

	// Then, Ready msg f+1 is sent by the message for which Echo count is f+1
	node0.ReceiveMessage(senderReady1.Details(), *msgToSend1)

	time.Sleep(1 * time.Second)
	
	// Checks
	// 1. node0 should have broadcast a Ready message
	assert.Equal(t, 1, countBroadcastedReadyMessages(transport))
	// 2. ReadySent must now be set to true
	assert.True(t, c.ReadySent)
}

func countBroadcastedReadyMessages(transport *MockTransport) int {
	sentMessages := transport.GetBroadcastedMessages()
	filteredMessages := make([]common.DKGMessage, 0)

	for _, msg := range sentMessages {
		if msg.Method == ReadyMessageType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return len(filteredMessages)
}

/*
Function: Process
Case: keygen already complete
Expects: no action!
- no message sent/broadcast
- keygen.State.Phase should NOT be equal to ENDED
- keygen.State.ReceivedReady does not change
*/
func TestKeygenAlreadtCompleted(t *testing.T) {
	// (Node 3 is the dealer of this round and is the one that sends shares to all nodes)
	// Node 0 is the receiver of all the ready msg
	nodes, transport, node0, round, hash, encodedShares := setupEchoHandlerTest()
	defaultKeygen := NewDefaultKeygen(round)
	keygen, _ := node0.State().KeygenStore.GetOrSetIfNotComplete(round.ID(), defaultKeygen)	
	// Set keygen to completed
	node0.State().KeygenStore.Complete(round.ID())
	
	// Node0 receives n "ready" messages
	for _, nodeDetails := range node0.Nodes() {
		senderEcho := nodes[nodeDetails.Index-1]
		msgToSend, _ := NewReadyMessage(round.ID(), encodedShares[senderEcho.id-1], hash, common.CurveName(c.Name))
		node0.ReceiveMessage(senderEcho.Details(), *msgToSend)
	}
	time.Sleep(1 * time.Second)

	// Checks
	// 1. node0 wasn't triggered to send any message
	assert.Equal(t, n, node0.messageCount) // the initial n messages
	// 2. no `Ready` message should have been broadcasted
	assert.Equal(t, 0, countBroadcastedReadyMessages(transport), "No `Ready` messages should have been broadcasted")
	// 3. keygen.State.ReceivedReady set to false for sender
	assert.False(t, keygen.State.ReceivedReady[node0.id])
	// 4. keygen.State.Phase should NOT be set to ENDED
	assert.NotEqual(t, common.Ended, keygen.State.Phase)
}

/*
Function: Process
Case: with the next message the node0 surpasses the threshold, but received a double `Ready` message
Expects: no action!
- no message sent/broadcast
- keygen.State.Phase should NOT be equal to ENDED
- keygen.State.ReceivedReady does not change
*/
func TestAlreadyReceivedReadyMessage(t *testing.T) {
	// (Node 3 is the dealer of this round and is the one that sends shares to all nodes)
	// Node 0 is the receiver of all the ready msg
	nodes, transport, node0, round, hash, encodedShares := setupEchoHandlerTest()
	defaultKeygen := NewDefaultKeygen(round)
	keygen, _ := node0.State().KeygenStore.GetOrSetIfNotComplete(round.ID(), defaultKeygen)	
	
	// Node0 receives 2*f "ready" messages
	for i :=0 ; i < 2*f ; i++ {
		senderEcho := nodes[i]
		msgToSend, _ := NewReadyMessage(round.ID(), encodedShares[senderEcho.id-1], hash, common.CurveName(c.Name))
		node0.ReceiveMessage(senderEcho.Details(), *msgToSend)
	}
	// And then receives message 2*f+1 
	// but it is one that was already received
	senderEcho := nodes[1]
	msgToSend, _ := NewReadyMessage(round.ID(), encodedShares[senderEcho.id-1], hash, common.CurveName(c.Name))
	node0.ReceiveMessage(senderEcho.Details(), *msgToSend)
	time.Sleep(1 * time.Second)

	// Checks
	// 1. node0 wasn't triggered to send any message
	assert.Equal(t, 2*f+1, node0.messageCount) // the initial 2*f+1 messages
	// 2. no `Ready` message should have been broadcasted
	assert.Equal(t, 0, countBroadcastedReadyMessages(transport), "No `Ready` messages should have been broadcasted")
	// 3. keygen.State.Phase has to be set to ENDED
	assert.NotEqual(t, common.Ended, keygen.State.Phase)
}
