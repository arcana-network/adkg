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
Case: f+1 NewAux1Message messages for this round, r and v have been received and node[n-1] hasn't broadcast NewAux1Message yet
Expects: node[n-1] broadcasts AuxsetMessage
*/
func TestSendAux1Msg(t *testing.T) {
	r := 0
	vote := 1

	//setup
	transport, nodes, msg, round := AuxTestSetup(r, vote)

	receiverNode := nodes[n-1]

	store, complete := receiverNode.State().ABAStore.GetOrSetIfNotComplete(round.ID(), common.DefaultABAStore())

	assert.Equal(t, complete, false, "should not be complete")

	// node[n-1] will receive NewAux1Message from node0
	// store already received NewAux1Message messages from n-2 other nodes

	for i := 1; i < n-1; i++ {
		store.SetValues("aux", r, vote, nodes[i].id)
	}

	store.SetBin("bin", r, vote)

	assert.NotNil(t, store.GetBin("bin", r))
	assert.GreaterOrEqual(t, len(store.Values("aux", r, vote)), n-f)

	// the (f+1)-th message should trigger sending
	receiverNode.ReceiveMessage(nodes[0].Details(), *msg)
	time.Sleep(1 * time.Second)

	// Check that AuxsetMessageType was broadcasted
	countBroadcastedAux1Msg := getCountMsg(transport, AuxsetMessageType)

	assert.Equal(t, 1, countBroadcastedAux1Msg, "This node should have broadcasted an Aux1MessageType")
}

/*
Function: Process
Case: receives double Aux1 message (from same sender)
Expects: no message broadcast even though f+1 Aux messages are received
*/
func TestAuxAlreadyReceived(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, round := AuxTestSetup(r, vote)

	receiverNode := nodes[n-1]
	// node[n-1] received Aux messages from nodes 0 through f-1
	store, _ := receiverNode.State().ABAStore.GetOrSetIfNotComplete(round.ID(), common.DefaultABAStore())
	for i := 0; i < f; i++ {
		store.SetValues("aux", r, vote, nodes[i].id)
	}

	// Node0 again sends Aux message which triggers early return
	receiverNode.ReceiveMessage(nodes[0].Details(), *msg)
	time.Sleep(1 * time.Second)

	// Check that Aux1MessageType was broadcasted
	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, 0, len(broadcastedMessages), "No message should have been broadcasted")
}

/*
Function: Process
Case: keygen process already completed
Expects: no message broadcast even though f+1 Aux1 messages are received
*/
func TestKeygenAlreadyCompleteAux(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, round := AuxTestSetup(r, vote)

	receiverNode := nodes[n-1]

	// Set the keygen state to completed
	state := receiverNode.State()
	state.ABAStore.Complete(round.ID())

	// Send f+1 Aux1 messages, which normally should trigger sending Aux1 message
	// but since keygen is marked complete won't trigger broadcast
	for i := 0; i < f+1; i++ {
		receiverNode.ReceiveMessage(nodes[i].Details(), *msg)
	}

	time.Sleep(1 * time.Second)

	// Check that whether NewAux1Message was broadcasted
	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, 0, len(broadcastedMessages), "No message should have been broadcasted")
}

func AuxTestSetup(r, vote int) (*MockTransport, []*Node, *common.DKGMessage, common.RoundDetails) {
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

	msg, error := NewAux1Message(round.ID(), vote, r, common.CurveName(c.Name))

	if error != nil {
		fmt.Println("cannot create aux1 msg")
	}
	return transport, nodes, msg, round
}
