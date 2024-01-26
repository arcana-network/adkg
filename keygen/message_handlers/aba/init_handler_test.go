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
In call from output_handler:
vote is 0 or 1
r is always set to 0
*/

/*
Function: Process
Case: happy path; sender is self, keygen is not complete yet & aba is not started
Expects:
- 1 Est1Message to be broadcasted
- in aba store for this round, "est" for the values should be set to "sent"
*/
func TestProcessInitMessageVote1(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, round := abaTestSetup(r, vote)
	
	// Check est "sent" is not yet set to true
	store, _ := nodes[1].State().ABAStore.GetOrSetIfNotComplete(round.ID(), common.DefaultABAStore())
	assert.False(t, store.Sent("est", r, vote))

	nodes[1].ReceiveMessage(nodes[1].Details(), *msg)
	// Add a small pause so all messages can be sent and received
	time.Sleep(1 * time.Second)

	// Check that NewEst1Msg was broadcasted
	countBroadcastedEst1Msg := getCountMsg(transport, Est1MessageType)
	assert.Equal(t, 1, countBroadcastedEst1Msg, "This node should have broadcasted an Est1MessageType")

	// Check est "sent" is now set to true
	assert.True(t, store.Sent("est", r, vote))
}

func TestProcessInitMessageVote0(t *testing.T) {
	r := 0
	vote := 0
	transport, nodes, msg, _ := abaTestSetup(r, vote)

	nodes[1].ReceiveMessage(nodes[1].Details(), *msg)
	time.Sleep(1 * time.Second)

	// Check that NewEst1Msg was broadcasted
	countBroadcastedEst1Msg := getCountMsg(transport, Est1MessageType)
	assert.Equal(t, 1, countBroadcastedEst1Msg, "This node should have broadcasted an Est1MessageType")
}

func getCountMsg(transport *MockTransport, msgType string) int {
	broadcastedMessages := transport.GetBroadcastedMessages()
	filteredMessages := make([]common.DKGMessage, 0)

	for _, msg := range broadcastedMessages {
		if msg.Method == msgType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	countBroadcasteMsg := len(filteredMessages)
	return countBroadcasteMsg
}

func abaTestSetup(r, vote int) (*MockTransport, []*Node, *common.DKGMessage, common.RoundDetails) {
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

	msg, _ := NewInitMessage(round.ID(), vote, r, common.CurveName(c.Name))
	return transport, nodes, msg, round
}

/*
Function: Process
Case: sender not equal to self
Expects: early return; no msg broadcast
*/
func TestSenderNotSelf(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, _ := abaTestSetup(vote, r)

	nodes[1].ReceiveMessage(nodes[0].Details(), *msg)
	time.Sleep(1 * time.Second)

	countBroadcastedEst1Msg := getCountMsg(transport, Est1MessageType)
	assert.Equal(t, 0, countBroadcastedEst1Msg, "No message should have been broadcast")
}

/*
Function: Process
Case: keygen already completed
Expects: early return; no msg broadcast
*/
func TestKeygenAlreadyCompleted(t *testing.T) {
	r := 0
	vote := 1
	transport, nodes, msg, round := abaTestSetup(vote, r)

	store := nodes[1].State().ABAStore
	store.Complete(round.ID())

	nodes[1].ReceiveMessage(nodes[1].Details(), *msg)
	time.Sleep(1 * time.Second)

	countBroadcastedEst1Msg := getCountMsg(transport, Est1MessageType)
	assert.Equal(t, 0, countBroadcastedEst1Msg, "No message should have been broadcast")
}

// TODO fix
/*
Function: Process
Case: aba already started (store.GetStarted(r) is true)
Expects: early return; no msg broadcast
*/
// func TestABAAlreadyStarted(t *testing.T) {
// 	r := 0
// 	vote := 1
// 	transport, nodes, msg, round := abaTestSetup(vote, r)

// 	store, _ := nodes[1].State().ABAStore.GetOrSetIfNotComplete(round.ID(), common.DefaultABAStore())
	
// 	store.SetStarted(r)
// 	log.Infof("bool %v", store.GetStarted(r) )
	
// 	nodes[1].ReceiveMessage(nodes[1].Details(), *msg)
// 	time.Sleep(1 * time.Second)

// 	countBroadcastedEst1Msg := getCountMsg(transport, Est1MessageType)
// 	assert.Equal(t, 0, countBroadcastedEst1Msg, "No message should have been broadcast")
// }
