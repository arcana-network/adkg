package dacss

import (
	"math/big"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/torusresearch/bijson"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TODO: add assertions
func TestDacss(t *testing.T) {
	log.SetLevel(log.InfoLevel)

	//default setup and mock transport
	TestSetUp, _ := DefaultTestSetup()

	nodesOld := TestSetUp.oldCommitteeNetwork

	nOld := TestSetUp.OldCommitteeParams.N
	kOld := TestSetUp.OldCommitteeParams.K

	// The old committee has shares of a single secret
	testSecret := sharing.GenerateSecret(curves.K256())
	_, shares, _ := sharing.GenerateCommitmentAndShares(testSecret, uint32(kOld), uint32(nOld), testutils.TestCurve())

	// Each node in old committee starts dACSS
	// That means that for the single share each node has,
	// ceil(nrShare/(nrOldNodes-2*recThreshold)) = 1 random values are sampled
	// and shared to both old & new committee

	for index, n := range nodesOld {
		go func(index int, node *PssTestNode2) {
			ephemeralKeypair := common.GenerateKeyPair(curves.K256())
			share := sharing.ShamirShare{Id: shares[index].Id, Value: shares[index].Value}
			initMsg := getTestInitMsgSingleShare(n, *big.NewInt(int64(index)), &share, ephemeralKeypair, TestSetUp.NewCommitteeParams)

			pssMsgData, err := bijson.Marshal(initMsg)
			assert.Nil(t, err)

			InitPssMessage := common.PSSMessage{
				PSSRoundDetails: initMsg.PSSRoundDetails,
				Type:            initMsg.Kind,
				Data:            pssMsgData,
			}
			node.ReceiveMessage(node.Details(), InitPssMessage)
		}(index, n)
	}

	time.Sleep(10 * time.Second)

	// TODO add assertions when the initial sampled values in InitHandler get stored and they can be checked
}

func getTestInitMsgSingleShare(testDealer *PssTestNode2, pssRoundIndex big.Int, share *sharing.ShamirShare, ephemeralKeypair common.KeyPair, newCommitteeParams common.CommitteeParams) *dacss.InitMessage {
	roundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(pssRoundIndex),
		Dealer: testDealer.Details(),
	}
	msg := &dacss.InitMessage{
		PSSRoundDetails:    roundDetails,
		OldShares:          []sharing.ShamirShare{*share},
		EphemeralSecretKey: ephemeralKeypair.PrivateKey.Bytes(),
		EphemeralPublicKey: ephemeralKeypair.PublicKey.ToAffineCompressed(),
		Kind:               dacss.InitMessageType,
		CurveName:          &common.SECP256K1,
		NewCommitteeParams: newCommitteeParams,
	}
	return msg
}
