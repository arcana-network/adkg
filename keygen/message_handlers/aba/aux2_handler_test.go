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
Case: f+1 NewAux2Message messages for this round, r and v have been received and node[n-1] hasn't broadcast NewAux2Message yet
Expects: node[n-1] sends NewCoinInitMessage
*/
//TODO: incomplete
func TestSendAux2Msg(t *testing.T) {
	r := 0
	vote := 1

	//setup
	_, nodes, msg, round := Aux2TestSetup(r, vote)

	receiverNode := nodes[n-1]

	store, complete := receiverNode.State().ABAStore.GetOrSetIfNotComplete(round.ID(), common.DefaultABAStore())

	assert.Equal(t, complete, false, "should not be complete")

	for i := 1; i < n-1; i++ {
		store.SetValues("aux2", r, 0, nodes[i].id)
		store.SetValues("aux2", r, 1, nodes[i].id)
		store.SetValues("aux2", r, 2, nodes[i].id)
	}

	//check that the store sets correctly
	for i := 1; i < n-1; i++ {
		assert.Equal(t, Contains(store.Values("aux2", r, 0), nodes[i].id), true)
		assert.Equal(t, Contains(store.Values("aux2", r, 1), nodes[i].id), true)
		assert.Equal(t, Contains(store.Values("aux2", r, 2), nodes[i].id), true)
	}

	store.SetBin("bin2", r, 0)
	store.SetBin("bin2", r, 1)
	store.SetBin("bin2", r, 2)

	assert.NotNil(t, store.GetBin("bin2", r))
	assert.Equal(t, Contains(store.GetBin("bin2", r), 0), true)
	assert.Equal(t, Contains(store.GetBin("bin2", r), 1), true)
	assert.Equal(t, Contains(store.GetBin("bin2", r), 2), true)

	assert.GreaterOrEqual(t, len(store.Values("aux2", r, vote)), n-f)

	//check that before sending msg aux2 value is not present
	assert.Equal(t, Contains(store.Values("aux2", r, vote), nodes[0].id), false)

	// the (f+1)-th message should trigger sending
	receiverNode.ReceiveMessage(nodes[0].Details(), *msg)
	time.Sleep(2 * time.Second)

	// assert.Equal(t, receiverNode.messageCount, 1)

	// Check that AuxsetMessageType was broadcasted

	//TODO: msg is not broadcasted but Received
	//TODO: expected an init msg
	// countReceivedAux2Msg := receiverNode.GetReceivedMessages(InitMessageType)
	// assert.Equal(t, 1, len(countReceivedAux2Msg), "This node should have broadcasted an CoinInitMessageType")

	//check that after sending msg aux value is present
	assert.Equal(t, Contains(store.Values("aux2", r, vote), nodes[0].id), true)
	sessionStore, complete := nodes[0].State().SessionStore.GetOrSetIfNotComplete(round.ADKGID, common.DefaultADKGSession())
	assert.Equal(t, sessionStore.ABAComplete, false)
}

func Aux2TestSetup(r, vote int) (*MockTransport, []*Node, *common.DKGMessage, common.RoundDetails) {
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

	msg, error := NewAux2Message(round.ID(), vote, r, common.CurveName(c.Name))

	if error != nil {
		fmt.Println("cannot create aux2 msg")
	}
	return transport, nodes, msg, round
}
