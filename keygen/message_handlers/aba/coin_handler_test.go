package aba

import (
	"fmt"
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
)

//TODO!

// func TestCoinMessageProcessing(t *testing.T) {
// 	//setup
// 	transport, nodes, msg, round := CoinTestSetup()

// 	receiverNode := nodes[n-1]

// 	_, complete := receiverNode.State().ABAStore.GetOrSetIfNotComplete(round.ID(), common.DefaultABAStore())
// 	assert.Equal(t, complete, false, "should not be complete")
// 	assert.Nil(t, receiverNode.State(), false)

// 	receiverNode.ReceiveMessage(nodes[0].Details(), *msg)
// 	time.Sleep(1 * time.Second)

// 	// Check that AuxsetMessageType was broadcasted
// 	countBroadcastedCoinMsg := getCountMsg(transport, keyderivation.InitMessageType)
// 	assert.Equal(t, 1, countBroadcastedCoinMsg, "This node should have broadcasted an Aux1MessageType")

// }

func CoinTestSetup() (*MockTransport, []*Node, *common.DKGMessage, common.RoundDetails) {
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

	data := make([]byte, 0)

	//random data being created
	// data = bytes of randomScalar || secp256k1_Generator || secp256k1_Generator || secp256k1_Generator
	// a correct format to upack
	data = append(data, randomScalar.Bytes()...)
	data = append(data, common.CurveFromName(common.SECP256K1).Point.Generator().ToAffineCompressed()...)
	data = append(data, common.CurveFromName(common.SECP256K1).Point.Generator().ToAffineCompressed()...)
	data = append(data, common.CurveFromName(common.SECP256K1).Point.Generator().ToAffineCompressed()...)

	msg, error := NewCoinMessage(round.ID(), data, common.CurveName(c.Name))

	if error != nil {
		fmt.Println("cannot create aux1 msg")
	}
	return transport, nodes, msg, round
}
