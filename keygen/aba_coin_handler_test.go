// The reason for this file to be here instead of in `aba/` is because of import cycle
package keygen

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/aba"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCoinMessageProcessing(t *testing.T) {

	// timeout := time.After(30 * time.Second)
	// done := make(chan bool)

	log.SetLevel(log.DebugLevel)

	// Set up nodes and transport
	id := common.GenerateADKGID(*big.NewInt(int64(1)))
	nodes, transport := setupNodes(7, 0)

	// For each node: simulate processing a CoinMessage
	for _, k := range nodes {
		go func(node *Node) {
			round := common.RoundDetails{
				ADKGID: id,
				Dealer: node.ID(),
				Kind:   "aba",
			}
			var data []byte = []byte{}
			//random data being created
			// data = bytes of randomScalar || secp256k1_Generator || secp256k1_Generator || secp256k1_Generator
			// a correct format to upack
			data = fmt.Append(data,
				randomScalar.Bytes())
			data = fmt.Append(data, common.CurveFromName(common.SECP256K1).Point.Identity().ToAffineCompressed())
			data = fmt.Append(data, common.CurveFromName(common.SECP256K1).Point.Identity().ToAffineCompressed())
			data = fmt.Append(data, common.CurveFromName(common.SECP256K1).Point.Identity().ToAffineCompressed())

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
			// time.Sleep(10 * time.Second)

		}(k)

		go func() {
			count := 0
			var outputs []string = []string{}
			output := <-transport.output
			outputs = append(outputs, output)
			count += 1

			assert.Equal(t, len(outputs), n, "length should be equal")

			// Checks all nodes output the same share
			for _, node1 := range nodes {
				for _, node2 := range nodes {
					t1 := node1.shares[1] != node2.shares[1]
					assert.Equal(t, t1, true, "shoudl be equal")
				}
			}
			// done <- true
		}()

		// select {
		// case <-timeout:
		// 	t.Fatal("Test didn't finish in time")
		// case <-done:
		// }

	}

}
