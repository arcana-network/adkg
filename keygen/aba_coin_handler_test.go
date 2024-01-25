// The reason for this file to be here instead of in `aba/` is because of import cycle
package keygen

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/aba"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/acss"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCoinMessageProcessing(t *testing.T) {

	timeout := time.After(60 * time.Second)
	done := make(chan bool)
	log.SetLevel(log.DebugLevel)

	// Set up nodes and transport
	id := common.GenerateADKGID(*big.NewInt(int64(1)))
	nodes, _ := setupNodes(7, 0)
	var round common.RoundDetails = common.RoundDetails{}
	// For each node: simulate processing a CoinMessage
	for _, n := range nodes {
		go func(node *Node) {
			round = common.RoundDetails{
				ADKGID: id,
				Dealer: node.ID(),
				Kind:   "aba",
			}
			data := make([]byte, 0)

			//random data being created
			// data = bytes of randomScalar || secp256k1_Generator || secp256k1_Generator || secp256k1_Generator
			// a correct format to upack
			data = append(data, randomScalar.Bytes()...)
			data = append(data, common.CurveFromName(common.SECP256K1).Point.Generator().ToAffineCompressed()...)
			data = append(data, common.CurveFromName(common.SECP256K1).Point.Generator().ToAffineCompressed()...)
			data = append(data, common.CurveFromName(common.SECP256K1).Point.Generator().ToAffineCompressed()...)

			// Create a sample CoinMessage
			coinMsg, err := aba.NewCoinMessage(
				round.ID(),
				data,
				common.SECP256K1,
			)
			if err != nil {
				log.WithError(err).Error("Error creating CoinMessage")
				return
			}
			fmt.Println(coinMsg)

			// check the state before sending the msg
			state := node.State().KeygenStore
			state_before, bool_before := state.Get(round.ID())
			assert.Nil(t, state_before, "State should be Nil initially.")
			assert.False(t, bool_before, "The state shouldn't have been initiated")

			// Process the CoinMessage
			node.ReceiveMessage(node.Details(), *coinMsg)

			// check the state after state
			state_after, bool_after := state.Get(round.ID())
			assert.NotNil(t, state_after, "State should have been initialized")
			assert.True(t, bool_after, "The state should have been initialized.")

			assert.True(t, state_after.Started, "Keygen state started should be True")

			// Allow time for nodes to process messages
			time.Sleep(1 * time.Second)
		}(n)
	}
	go func() {
		for {

			// Checks all nodes output the same share
			for _, node1 := range nodes {
				for _, node2 := range nodes {

					//check that ABAstate is same for all nodes
					state2, _ := node1.state.ABAStore.Get(round.ID())
					state1, _ := node1.state.ABAStore.Get(round.ID())
					assert.Equal(t, state1, state2)

					t1 := node1.shares[1] == node2.shares[1]
					assert.Equal(t, t1, true, "shoudl be equal")
				}
			}
			done <- true
		}

	}()

	select {
	case <-timeout:
		t.Fatal("Test didn't finish in time")
	case <-done:

	}

}

func TestABACompletness(t *testing.T) {

	// The ACSS is done to get the Full flow and get the ABA state
	timeout := time.After(10 * time.Second)
	done := make(chan bool)

	log.SetLevel(log.DebugLevel)
	var round common.RoundDetails = common.RoundDetails{}

	nodes, _ := setupNodes(7, 0)

	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	for _, n := range nodes {
		go func(node *Node) {
			round = common.RoundDetails{
				ADKGID: id,
				Dealer: node.ID(),
				Kind:   "acss",
			}

			msg, err := acss.NewShareMessage(
				round.ID(),
				common.SECP256K1,
			)
			if err != nil {
				log.WithError(err).Error("EndBlock:Acss.NewShareMessage")
			}
			node.ReceiveMessage(node.Details(), *msg)
		}(n)

		//Once the acss is completed for all nodes and broadcasted,
		// each node will have the set Ti to be agreed upon using ABA
	}

	go func() {
		for {
			for i := range nodes {
				state, _ := nodes[i].state.ABAStore.Get(round.ID())
				state2, _ := nodes[(i+1)%7].state.ABAStore.Get(round.ID())
				assert.Equal(t, state, state2, "should be equal")

			}
			done <- true
		}
	}()

	select {
	case <-timeout:
		t.Fatal("Test didn't finish in time")
	case <-done:
	}
}
