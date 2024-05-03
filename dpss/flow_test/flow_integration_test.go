package dpss

import (
	"crypto/rand"
	"math/big"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
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
	timeout := time.After(30 * time.Second)
	done := make(chan bool)
	log.SetLevel(log.InfoLevel)

	testSetup, _ := DefaultTestSetup()

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
	// expect him message to be sent
	// expect n-f nodes to have agreed (Check decision and TSet)
	select {
	case <-timeout:
		t.Fatal("Test didn't finish in time")
	case <-done:
	}
}
