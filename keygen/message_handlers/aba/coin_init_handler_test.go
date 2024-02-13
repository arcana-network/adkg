package aba

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

/*
Function: Process
Case: f+1 CoinInitMessage messages for this round, r and v have been received and node[n-1] hasn't broadcast NewCoinMessage yet
Expects: node[n-1] sends NewCoinMessage
*/

// TODO: Incomplete
func TestSendCoinInit(t *testing.T) {

	//setup
	transport, nodes, msg, round := CoinInitTestSetup()

	receiverNode := nodes[n-1]

	_, complete := receiverNode.State().ABAStore.GetOrSetIfNotComplete(round.ID(), common.DefaultABAStore())

	Store, _ := receiverNode.State().SessionStore.GetOrSetIfNotComplete(round.ADKGID, common.DefaultADKGSession())
	// Store2, _ := nodes[0].State().SessionStore.GetOrSetIfNotComplete(round.ADKGID, common.DefaultADKGSession())

	Store.T[int(round.Dealer)] = 5 //some random value
	// Store2.T[int(round.Dealer)] = 5 //some random value

	Store.S[3] = sharing.ShamirShare{
		Id:    3,
		Value: []byte{155, 76, 0, 68, 124, 111, 88, 80, 158, 27, 75, 191, 11, 203, 22, 125, 223, 231, 128, 135, 16, 74, 60, 220, 41, 83, 174, 44, 89, 180, 145, 49},
	}
	assert.Equal(t, complete, false, "should not be complete")

	receiverNode.ReceiveMessage(nodes[0].Details(), *msg)
	time.Sleep(2 * time.Second)

	countBroadcastedCoinMsg := getCountMsg(transport, CoinMessageType)
	assert.Equal(t, 1, countBroadcastedCoinMsg, "This node should have broadcasted an CoinMessageType")

	sessionStore, complete := nodes[0].State().SessionStore.GetOrSetIfNotComplete(round.ADKGID, common.DefaultADKGSession())
	assert.Equal(t, sessionStore.ABAComplete, false)
}

func CoinInitTestSetup() (*MockTransport, []*Node, *common.DKGMessage, common.RoundDetails) {
	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)

	nodes, transport := setupNodes(n, 0)

	leaderIndex := 3
	leader := nodes[leaderIndex]

	round := common.RoundDetails{
		ADKGID: id,
		Dealer: leader.ID(),
		Kind:   "aba",
	}

	coin_id := "aba_coin"
	msg, error := NewCoinInitMessage(round.ID(), coin_id, common.CurveName(c.Name))

	if error != nil {
		fmt.Println("cannot create aux2 msg")
	}
	return transport, nodes, msg, round
}
