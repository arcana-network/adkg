package acss

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/messages"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

/*
Function: Process

Happy path check;
1. sender equals self
2. keygen (state) will be initialized
3. keygen state "Started" must equal false before and true after processing
4. each node must have received a ProposeMessage from all the other nodes (including itself)
5. all ProposeMessages must have a share for all nodes in the sharesMap
6. the shares & commitments in the sharesMap can be verified
*/
func processShareMsg(t *testing.T) {
	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)

	nodes, transport := setupNodes(7, 0)

	timeout := time.After(30 * time.Second)
	done := make(chan bool)

	// For each round: create a round for which this node is the leader
	// the node will be triggered to create a secret, and distribute the shares with the other nodes
	// all the other nodes process the share as `acssProposal` with current node as leader
	for _, n := range nodes {

		go func(node *Node) {
			round := common.RoundDetails{
				ADKGID: id,
				Dealer: node.ID(),
				Kind:   "acss",
			}
			msg, err := NewShareMessage(
				round.ID(),
				common.SECP256K1,
			)
			if err != nil {
				log.WithError(err).Error("EndBlock:Acss.NewShareMessage")
			}

			// Checks:
			// 1 (implicit): sender equals self. If sender equals self, something will be stored in the keygen (otherwise it will return early)
			// 2: keygen will be initialized
			// Check nothing was stored before sending the msg
			state := node.State().KeygenStore
			state_before, bool_before := state.Get(round.ID())
			assert.Nil(t, state_before, "State should be Nil initially.")
			assert.False(t, bool_before, "The state shouldn't have been initiated")

			// trigger node to create a secret and secret share it with the other nodes
			node.ReceiveMessage(node.Details(), *msg)
			// Add a small pause so all messages can be sent and received
			time.Sleep(1 * time.Second)

			state_after, bool_after := state.Get(round.ID())
			assert.NotNil(t, state_after, "State should have been initialized")
			assert.True(t, bool_after, "The state should have been initialized.")

			// Check 3: keygen state "Started" must equal true
			assert.True(t, state_after.Started, "Keygen state started should be True")

			// Check 4: each node must have received 7 ProposeMessages (and thus have sent 7 ProposeMessages)
			broadcastedMessages := transport.GetBroadcastedMessages()
			filteredMessages := make([]common.DKGMessage, 0)

			for _, msg := range broadcastedMessages {
				if msg.Method == ProposeMessageType {
					filteredMessages = append(filteredMessages, msg)
				}
			}
			assert.Equal(t, len(filteredMessages), 7, "This node should have received 7 ProposeMsgs")

			// Check 5: all ProposeMessages must have 7 shares in their sharesMap and they must all belong to different id's
			for _, msg := range filteredMessages {
				var proposeMsg ProposeMessage
				err := json.Unmarshal(msg.Data, &proposeMsg)
				if err != nil {
					log.Fatalf("Error parsing ProposeMsg JSON: %v", err)
				}

				dataField := proposeMsg.Data
				var msgData messages.MessageData
				err = json.Unmarshal(dataField, &msgData)
				if err != nil {
					fmt.Println("Error during Unmarshal():", err)
					return
				}
				assert.Equal(t, 7, len(msgData.ShareMap))

				// Check 6: the shares and commitments can be verified
				// TODO fixme
				// _, k, _ := node.Params()
				// sender_id, _ := msg.RoundID.Leader()
				// node_sk := node.PrivateKey()
				// sender_pk := nodes[sender_id.Int64()].keypair.PublicKey
				// aes_key := acss.SharedKey(node_sk, sender_pk)
				// curve := common.CurveFromName(common.SECP256K1)
				// _, _, verified := acss.Predicate(aes_key[:], msgData.ShareMap[uint32(node.ID())][:],
				// 	msgData.Commitments[:], k, curve)
				// assert.True(t, verified)
			}

			done <- true
		}(n)

	}

	select {
	case <-timeout:
		t.Fatal("Test didn't finish in time")
	case <-done:
	}
}

/*
Function: Process
Case: sender not equal to self
Expects: early return

We send a message as node0 to node1
Checks:
1. keygen state will not be initialized
2. no ProposeMessage will be sent
*/
func processSenderNotNode(t *testing.T) {
	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)

	nodes, transport := setupNodes(7, 0)

	node0 := nodes[0]
	node1 := nodes[1]

	round := common.RoundDetails{
		ADKGID: id,
		Dealer: node0.ID(),
		Kind:   "acss",
	}
	msg, err := NewShareMessage(
		round.ID(),
		common.SECP256K1,
	)
	if err != nil {
		log.WithError(err).Error("EndBlock:Acss.NewShareMessage")
	}

	state := node1.State().KeygenStore
	state_before, bool_before := state.Get(round.ID())
	assert.Nil(t, state_before, "State should be Nil initially.")
	assert.False(t, bool_before, "The state shouldn't have been initiated")

	node1.ReceiveMessage(node0.Details(), *msg)
	// Add a small pause so all messages can be sent and received
	time.Sleep(500 * time.Millisecond)

	// Check 1: keygen state will not be initialized
	state_after, bool_after := state.Get(round.ID())
	assert.Nil(t, state_after, "State shouldn't have been changed.")
	assert.False(t, bool_after, "The state shouldn't have been initiated")

	// Check 2: no ProposeMessage will be sent
	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, len(broadcastedMessages), 0)
}

/*
Function: Process
Case: keygen already completed
Expects: early return

Checks:
1. no ProposeMessage will be sent
*/
func processKeygenAlreadyCompleted(t *testing.T) {
	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)

	nodes, transport := setupNodes(7, 0)

	node0 := nodes[0]

	round := common.RoundDetails{
		ADKGID: id,
		Dealer: node0.ID(),
		Kind:   "acss",
	}
	msg, err := NewShareMessage(
		round.ID(),
		common.SECP256K1,
	)
	if err != nil {
		log.WithError(err).Error("EndBlock:Acss.NewShareMessage")
	}

	state := node0.State().KeygenStore
	state_before, bool_before := state.Get(round.ID())
	assert.Nil(t, state_before, "State should be Nil initially.")
	assert.False(t, bool_before, "The state shouldn't have been initiated")
	// Manually set the keygen state to complete, to trigger early return
	state.Complete(round.ID())

	node0.ReceiveMessage(node0.Details(), *msg)
	// Add a small pause so all messages can be sent and received
	time.Sleep(500 * time.Millisecond)

	// Check 1: no ProposeMessage will be sent
	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, len(broadcastedMessages), 0)
}

/*
Function: Process
Case: keygen state "Started" already true
Expects: early return
*/
func processKeygenAlreadyStarted(t *testing.T) {
	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)

	nodes, transport := setupNodes(7, 0)

	node0 := nodes[0]

	round := common.RoundDetails{
		ADKGID: id,
		Dealer: node0.ID(),
		Kind:   "acss",
	}
	msg, err := NewShareMessage(
		round.ID(),
		common.SECP256K1,
	)
	if err != nil {
		log.WithError(err).Error("EndBlock:Acss.NewShareMessage")
	}

	state := node0.State().KeygenStore
	state_before, bool_before := state.Get(round.ID())
	assert.Nil(t, state_before, "State should be Nil initially.")
	assert.False(t, bool_before, "The state shouldn't have been initiated")

	// Set keygen started to "true" in the node state
	defaultKeygen := &common.SharingStore{
		RoundID: round.ID(),
		State: common.RBCState{
			Phase:         common.Initial,
			ReceivedReady: make(map[int]bool),
			ReceivedEcho:  make(map[int]bool),
		},
		CStore:  make(map[string]*common.CStore),
		Started: false,
	}
	keygen, _ := state.GetOrSetIfNotComplete(round.ID(), defaultKeygen)
	keygen.Started = true

	node0.ReceiveMessage(node0.Details(), *msg)
	// Add a small pause so all messages can be sent and received
	time.Sleep(500 * time.Millisecond)

	// Check 1: no ProposeMessage will be sent
	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, len(broadcastedMessages), 0)
}

type testFunc func(*testing.T)

func TestACSS(t *testing.T) {
	// To prevent any issues between different testcases, we call them in this way, instead of separate test functions

	var wg sync.WaitGroup

	// Tests to be executed
	tests := []testFunc{
		// Function: Process. Test Happy Path for catching "ShareMessage" msg
		processShareMsg,
		// Function: Process. Testcase: sender node not equal to node receiving the msg
		processSenderNotNode,
		// Function: Process. Testcase: keygen was already completed (triggers early return)
		processKeygenAlreadyCompleted,
		// Function: Process. Testcase: keygen already started (triggers early return)
		processKeygenAlreadyStarted,
		// Add any new test functions here...
	}

	// Iterate over the test functions and execute them concurrently
	for _, test := range tests {
		wg.Add(1)
		go func(tf testFunc) {
			defer wg.Done()
			tf(t)
		}(test)
	}

	wg.Wait()
}
