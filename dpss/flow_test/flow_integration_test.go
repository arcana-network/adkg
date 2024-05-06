package dpss

import (
	"crypto/rand"
	"math/big"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/old_committee"
)

/*
1 -> [a1, b1, c1]
2 -> [a2, b2, c2]
...
*/
func getOldShares(n, k uint32, numOfShares int) map[int][]common.PrivKeyShare {
	shareList := make(map[int][]common.PrivKeyShare)
	c := common.CurveFromName(common.SECP256K1)
	for i := range numOfShares {
		log.Infof("Generating share: %d", i)
		s, err := sharing.NewShamir(k, n, c)
		if err != nil {
			log.Error(err)
		}
		shares, err := s.Split(c.Scalar.Random(rand.Reader), rand.Reader)
		if err != nil {
			log.Error(err)
		}
		for i, share := range shares {
			shareList[i+1] = append(shareList[i+1], common.PrivKeyShare{UserIdOwner: "test", Share: *share})
		}
	}

	return shareList
}

func TestFlow(t *testing.T) {
	// timeout := time.After(30 * time.Second)
	// done := make(chan bool)
	log.SetLevel(log.InfoLevel)

	testSetup, transport := DefaultTestSetup()

	shareList := getOldShares(uint32(testSetup.OldCommitteeParams.N), uint32(testSetup.OldCommitteeParams.K), 1)

	pssID := common.GeneratePSSID(*big.NewInt(1))
	for _, n := range testSetup.oldCommitteeNetwork {
		go func(node *PssTestNode2) {
			shares := shareList[node.details.Index]
			pssRoundDetails := common.PSSRoundDetails{
				PssID:     pssID,
				Dealer:    node.details,
				BatchSize: 1,
			}
			ephemeralKeypairDealer := common.GenerateKeyPair(common.CurveFromName(common.SECP256K1))
			m, _ := dacss.NewInitMessage(pssRoundDetails, shares, common.SECP256K1, ephemeralKeypairDealer, testSetup.NewCommitteeParams)
			n.ReceiveMessage(node.Details(), *m)
		}(n)
	}

	time.Sleep(time.Second * 5)

	// Add conditions to check
	receivedMsg := transport.GetReceivedMessages()
	count := 0

	// count the number of him msg sent
	for _, msg := range receivedMsg {
		if msg.Type == old_committee.DpssHimHandlerType {
			count++
		}
	}
	// expect him message to be sent
	//check number of him msg sent is equal to the number of old-committee nodes
	assert.Equal(t, count, len(testSetup.oldCommitteeNetwork))

	var TSet [][]int

	for _, node := range testSetup.oldCommitteeNetwork {
		state, _ := node.State().PSSStore.Get(pssID)
		nodeTSet := state.GetTSet(testSetup.OldCommitteeParams.N, testSetup.OldCommitteeParams.T)

		TSet = append(TSet, nodeTSet)
	}

	// expect n-f nodes to have agreed (Check decision and TSet)
	// Since it checks the happy path, all node should agree
	for i := 1; i < len(testSetup.oldCommitteeNetwork); i++ {

		assert.Equal(t, TSet[0], TSet[i])
	}

	// select {
	// case <-timeout:
	// 	t.Fatal("Test didn't finish in time")
	// case <-done:
	// }
}
