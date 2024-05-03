package aba

import (
	"math/big"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

/*
Function: Process
Case: f+1 est messages for this round, r and v have been received and node[n-1] hasn't broadcast est1 yet
Expects: node[n-1] broadcasts est1
*/
func TestSendEst1(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, round := estTestSetup(r, vote)

	receiverNode := nodes[n-1]
	// node[n-1] will receive est1 from node0
	// store already received est messages from f other nodes
	// then, the message from node0 triggers broadcasting est1 by node[n-1]
	store, _ := receiverNode.State().ABAStore.GetOrSetIfNotComplete(round.ToRoundID(), common.DefaultABAStore())
	for i := 1; i < f+1; i++ {
		store.SetValues("est", r, vote, nodes[i].id)
	}

	// the (f+1)-th message should trigger sending
	receiverNode.ReceiveMessage(nodes[0].Details(), *msg)
	time.Sleep(1 * time.Second)

	// Check that NewEst1Msg was broadcasted
	countBroadcastedEst1Msg := getCountMsg(transport, Est1MessageType)
	assert.Equal(t, 1, countBroadcastedEst1Msg, "This node should have broadcasted an Est1MessageType")
}

/*
Function: Process
Case: 2f+1 est messages for this round, r and v have been received
Expects: node[n-1] broadcasts aux1
*/
func TestSendAux1(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, round := estTestSetup(r, vote)

	receiverNode := nodes[n-1]
	// node[n-1] will receive est1 from node0
	// store already received est messages from 2f other nodes
	// then, the message from node0 triggers broadcasting aux1 by node[n-1]
	store, _ := receiverNode.State().ABAStore.GetOrSetIfNotComplete(round.ToRoundID(), common.DefaultABAStore())
	// mark the EST1 msg as sent
	store.SetSent("est", r, vote)

	for i := 1; i < (2*f + 1); i++ {
		store.SetValues("est", r, vote, nodes[i].id)
	}

	// the (2f+1)-th message should trigger sending
	receiverNode.ReceiveMessage(nodes[0].Details(), *msg)
	time.Sleep(1 * time.Second)

	// Check that Aux1Msg was broadcasted
	countBroadcastedAux1Msg := getCountMsg(transport, Aux1MessageType)
	assert.Equal(t, 1, countBroadcastedAux1Msg, "This node should have broadcasted an Aux1 msg")
}

func estTestSetup(r, vote int) (*MockTransport, []*Node, *common.PSSMessage, common.PSSRoundDetails) {
	id := common.GeneratePSSID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)

	nodes, transport := setupNodes(n, 0)
	// The leader doesn't matter
	leaderIndex := 3
	leader := nodes[leaderIndex]

	round := common.PSSRoundDetails{
		PssID:  id,
		Dealer: leader.Details(),
	}

	msg, _ := NewEst1Message(round, vote, r, common.CurveName(c.Name))
	return transport, nodes, msg, round
}

/*
Function: Process
Case: keygen process already completed
Expects: no message broadcast even though f+1 est1 messages are received
*/
func TestKeygenAlreadyComplete(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, round := estTestSetup(r, vote)

	receiverNode := nodes[n-1]

	// Set the keygen state to completed
	state := receiverNode.State()
	state.ABAStore.Complete(round.ToRoundID())

	// Send f+1 Est1 messages, which normally should trigger sending Est1 message
	// but since keygen is marked complete won't trigger broadcast
	for i := 0; i < f+1; i++ {
		receiverNode.ReceiveMessage(nodes[i].Details(), *msg)
	}

	time.Sleep(1 * time.Second)

	// Check that NewEst1Msg was broadcasted
	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, 0, len(broadcastedMessages), "No message should have been broadcasted")
}

/*
Function: Process
Case: receives double est1 message (from same sender)
Expects: no message broadcast even though f+1 est1 messages are received
*/
func TestEstAlreadyReceived(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, round := estTestSetup(r, vote)

	receiverNode := nodes[n-1]
	// node[n-1] received est1 messages from nodes 0 through f-1
	store, _ := receiverNode.State().ABAStore.GetOrSetIfNotComplete(round.ToRoundID(), common.DefaultABAStore())
	for i := 0; i < f; i++ {
		store.SetValues("est", r, vote, nodes[i].id)
	}

	// Node0 again sends est1 message which triggers early return
	receiverNode.ReceiveMessage(nodes[0].Details(), *msg)
	time.Sleep(1 * time.Second)

	// Check that NewEst1Msg was broadcasted
	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, 0, len(broadcastedMessages), "No message should have been broadcasted")
}
