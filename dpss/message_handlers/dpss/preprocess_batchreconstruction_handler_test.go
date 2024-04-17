package dpss

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

func TestProcessPreprocessBatchRecMessage(t *testing.T) {
	// Running for multiple test cases
	testCases := []struct {
		name               string
		nr_batches_minus_1 int
	}{
		{"ThreeBatches", 3 - 1},
		{"FourBatches", 4 - 1},
		{"HundredBatches", 100 - 1},
		{"ThousandBatches", 1000 - 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defaultSetup := testutils.DefaultTestSetup()
			n := defaultSetup.OldCommitteeParams.N
			param_t := defaultSetup.OldCommitteeParams.T

			// We make the total batch size such that it is 1 larger
			// than the multiple of the batchsize
			B := tc.nr_batches_minus_1*(n-2*param_t) + 1

			rScalars := make([]curves.Scalar, B)
			for i := 0; i < B; i++ {
				rScalars[i] = testutils.TestCurve().Scalar.Random(rand.Reader)
			}
			rValues := sharing.CompressScalars(rScalars)

			testNode := defaultSetup.GetSingleOldNodeFromTestSetup()

			pssID := big.NewInt(1)
			pssRoundDetails := common.PSSRoundDetails{
				Dealer:    testNode.Details(),
				PssID:     common.NewPssID(*pssID),
				BatchSize: B,
			}

			testMsg := PreprocessBatchRecMessage{
				PSSRoundDetails: pssRoundDetails,
				Kind:            PreprocessBatchRecMessageType,
				RValues:         rValues,
				CurveName:       testutils.TestCurveName(),
			}

			testNode.State().ShareStore.Initialize(B)
			for i := 0; i < B; i++ {
				testNode.State().ShareStore.OldShares[i] = testutils.TestCurve().Scalar.Random(rand.Reader)
			}

			testMsg.Process(testNode.Details(), testNode)

			// Wait for all the expected messages to be received
			testNode.Transport().WaitForMessagesReceived(tc.nr_batches_minus_1 + 1)

			assert.Equal(t, tc.nr_batches_minus_1+1, len(testNode.Transport().ReceivedMessages))
		})
	}
}
