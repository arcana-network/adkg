package dpss

import (
	"crypto/rand"
	"math/big"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

/*
Function: Process

Testcase: happy path.
*/
func TestInitRec(test *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode, dealerNode := defaultSetup.GetTwoOldNodesFromTestSetup()

	recMessage := createInitRecMessage(
		dealerNode,
		*testutils.TestCurve(),
	)

	recMessage.Process(receiverNode.Details(), receiverNode)
	time.Sleep(10 * time.Second)

	// The ammount of messages sent is n-2t
	n, _, _ := receiverNode.Params()
	assert.Equal(
		test,
		n,
		len(receiverNode.Transport.GetSentMessages()),
	)
}

func createInitRecMessage(dealerNode *testutils.PssTestNode, curve curves.Curve) InitRecMessage {
	pssID := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		Dealer: dealerNode.Details(),
		PssID:  common.NewPssID(*pssID),
	}
	dpssBatchRecDetails := common.DPSSBatchRecDetails{
		PSSRoundDetails: pssRoundDetails,
		BatchRecCount:   1,
	}

	// Create n - 2t random numbers
	n, _, t := dealerNode.Params()
	shareBatch := make([]curves.Scalar, 0)
	for range n - 2*t {
		share := curve.Scalar.Random(rand.Reader)
		shareBatch = append(shareBatch, share)
	}
	shareBatchBytes := sharing.CompressShares(shareBatch)

	msg := InitRecMessage{
		DPSSBatchRecDetails: dpssBatchRecDetails,
		ShareBatch:          shareBatchBytes,
		Curve:               testutils.TestCurveName(),
		Kind:                InitRecHandlerType,
	}

	return msg
}
