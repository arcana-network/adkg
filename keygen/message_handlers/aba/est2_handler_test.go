package aba

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

/*
Function: Process
Case: f+1 NewEst2Message messages for this round, r and v have been received and node[n-1] hasn't broadcast NewEst2Message yet
Expects: node[n-1] broadcasts NewEst2Message
*/
func TestSendEst2Msg(t *testing.T) {
	r := 0
	vote := 1

	//setup
	transport, nodes, msg, round := est2TestSetup(r, vote)

	receiverNode := nodes[n-1]

	store, complete := receiverNode.State().ABAStore.GetOrSetIfNotComplete(round.ID(), common.DefaultABAStore())

	assert.Equal(t, complete, false, "should not be complete")

	// node[n-1] will receive NewEst2Message from node0
	// store already received NewEst2Message messages from n-2 other nodes

	for i := 1; i < n-1; i++ {
		store.SetValues("est2", r, vote, nodes[i].id)
	}

	//check that the store sets correctly
	for i := 1; i < n-1; i++ {
		assert.Equal(t, Contains(store.Values("est2", r, vote), nodes[i].id), true)
	}

	store.SetBin("bin", r, vote)

	assert.NotNil(t, store.GetBin("bin", r))
	assert.Equal(t, Contains(store.GetBin("bin", r), 1), true)

	assert.GreaterOrEqual(t, len(store.Values("est2", r, vote)), n-f)

	//check that before sending msg est2 value is not present
	assert.Equal(t, Contains(store.Values("est2", r, vote), nodes[0].id), false)

	// the (f+1)-th message should trigger sending
	receiverNode.ReceiveMessage(nodes[0].Details(), *msg)
	time.Sleep(1 * time.Second)

	// Check that NewEst2Message was broadcasted
	countBroadcastedEst2Msg := getCountMsg(transport, Aux2MessageType)

	assert.Equal(t, 1, countBroadcastedEst2Msg, "This node should have broadcasted an Aux2MessageType")

	//check that after sending msg est2 value is present
	assert.Equal(t, Contains(store.Values("est2", r, vote), nodes[0].id), true)
	assert.Equal(t, store.Sent("est2", r, vote), true)

	assert.Equal(t, Contains(store.Values("est2", r, vote), receiverNode.id), true)

}

/*
Function: Process
Case: receives double Est2 message (from same sender)
Expects: no message broadcast even though f+1 Est2 messages are received
*/
func TestEst2AlreadyReceived(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, round := est2TestSetup(r, vote)

	receiverNode := nodes[n-1]
	// node[n-1] received Est2 messages from nodes 0 through f-1
	store, _ := receiverNode.State().ABAStore.GetOrSetIfNotComplete(round.ID(), common.DefaultABAStore())
	for i := 0; i < f; i++ {
		store.SetValues("est2", r, vote, nodes[i].id)
	}

	// Node0 again sends Est2 message which triggers early return
	receiverNode.ReceiveMessage(nodes[0].Details(), *msg)
	time.Sleep(1 * time.Second)

	// Check whether msg was broadcasted
	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, 0, len(broadcastedMessages), "No message should have been broadcasted")
}

/*
Function: Process
Case: keygen process already completed
Expects: no message broadcast even though f+1 Est2 messages are received
*/
func TestKeygenAlreadyCompleteEst2(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, round := est2TestSetup(r, vote)

	receiverNode := nodes[n-1]

	// Set the keygen state to completed
	state := receiverNode.State()
	state.ABAStore.Complete(round.ID())

	// Send f+1 Est2 messages, which normally should trigger sending NewAux2Message message
	// but since keygen is marked complete won't trigger broadcast
	for i := 0; i < f+1; i++ {
		receiverNode.ReceiveMessage(nodes[i].Details(), *msg)
	}

	time.Sleep(1 * time.Second)

	// Check that whether NewAux2Message was broadcasted
	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, 0, len(broadcastedMessages), "No message should have been broadcasted")
}

func est2TestSetup(r, vote int) (*MockTransport, []*Node, *common.DKGMessage, common.RoundDetails) {
	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)

	nodes, transport := setupNodes(n, 0)
	// The leader doesn't matter
	leaderIndex := 3
	leader := nodes[leaderIndex]

	round := common.RoundDetails{
		ADKGID: id,
		Dealer: leader.ID(),
		Kind:   "aba",
	}

	msg, error := NewEst2Message(round.ID(), vote, r, common.CurveName(c.Name))

	if error != nil {
		fmt.Println("cannot create est2 msg")
	}
	return transport, nodes, msg, round
}
