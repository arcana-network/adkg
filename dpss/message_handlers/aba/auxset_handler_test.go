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

func TestSendAuxset(t *testing.T) {
	r := 0
	vote := 1

	//setup
	transport, nodes, msg, round := AuxsetTestSetup(r, vote)

	receiverNode := nodes[n-1]

	store, complete := receiverNode.State().ABAStore.GetOrSetIfNotComplete(round.ToRoundID(), common.DefaultABAStore())

	assert.Equal(t, complete, false, "should not be complete")

	// node[n-1] will receive NewAuxsetMessage from node0
	// store already received NewAuxsetMessage messages from n-2 other nodes

	for i := 1; i < n-1; i++ {
		store.SetValues("auxset", r, vote, nodes[i].id)
	}

	//check that the store sets correctly
	for i := 1; i < n-1; i++ {
		assert.Equal(t, Contains(store.Values("auxset", r, vote), nodes[i].id), true)
	}

	store.SetBin("bin", r, vote)

	assert.NotNil(t, store.GetBin("bin", r))
	assert.Equal(t, Contains(store.GetBin("bin", r), vote), true)

	assert.GreaterOrEqual(t, len(store.Values("auxset", r, vote)), n-f)

	//check that before sending msg aux value is not present
	assert.Equal(t, Contains(store.Values("auxset", r, vote), nodes[0].id), false)

	// the (f+1)-th message should trigger sending
	receiverNode.ReceiveMessage(nodes[0].Details(), *msg)
	time.Sleep(1 * time.Second)

	// Check that Est2MessageType was broadcasted
	countBroadcastedAuxsetMsg := getCountMsg(transport, Est2MessageType)

	assert.Equal(t, 1, countBroadcastedAuxsetMsg, "This node should have broadcasted an Est2MessageType")

	//check that after sending msg aux value is present
	assert.Equal(t, Contains(store.Values("auxset", r, vote), nodes[0].id), true)

	assert.Equal(t, Contains(store.Values("est2", r, vote), receiverNode.id), true)

}

/*
Function: Process
Case: receives double Auxset message (from same sender)
Expects: no message broadcast even though f+1 Auxset messages are received
*/
func TestAuxsetAlreadyReceived(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, round := AuxsetTestSetup(r, vote)

	receiverNode := nodes[n-1]
	// node[n-1] received Auxset messages from nodes 0 through f-1
	store, _ := receiverNode.State().ABAStore.GetOrSetIfNotComplete(round.ToRoundID(), common.DefaultABAStore())
	for i := 0; i < f; i++ {
		store.SetValues("auxset", r, vote, nodes[i].id)
	}

	// Node0 again sends Auxset message which triggers early return
	receiverNode.ReceiveMessage(nodes[0].Details(), *msg)
	time.Sleep(1 * time.Second)

	// Check whether msg was broadcasted
	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, 0, len(broadcastedMessages), "No message should have been broadcasted")
}

/*
Function: Process
Case: keygen process already completed
Expects: no message broadcast even though f+1 Auxset messages are received
*/
func TestKeygenAlreadyCompleteAuxset(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, round := AuxsetTestSetup(r, vote)

	receiverNode := nodes[n-1]

	// Set the keygen state to completed
	state := receiverNode.State()
	state.ABAStore.Complete(round.ToRoundID())

	// Send f+1 Auxset messages, which normally should trigger sending Auxset message
	// but since keygen is marked complete won't trigger broadcast
	for i := 0; i < n; i++ {
		receiverNode.ReceiveMessage(nodes[i].Details(), *msg)
	}

	time.Sleep(1 * time.Second)

	// Check that whether msg was broadcasted
	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, 0, len(broadcastedMessages), "No message should have been broadcasted")
}

func AuxsetTestSetup(r, vote int) (*MockTransport, []*Node, *common.PSSMessage, common.PSSRoundDetails) {
	id := common.GeneratePSSID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)

	nodes, transport := setupNodes(n, 0)
	// The leader doesn't matter
	leaderIndex := 3
	leader := nodes[leaderIndex]

	round := common.PSSRoundDetails{
		PssID:  id,
		Dealer: leader.Details(),
		Kind:   "aba",
	}

	msg, error := NewAuxsetMessage(round, vote, r, common.CurveName(c.Name))

	if error != nil {
		fmt.Println("cannot create auxset msg")
	}
	return transport, nodes, msg, round
}
