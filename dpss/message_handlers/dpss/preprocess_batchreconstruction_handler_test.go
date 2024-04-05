package dpss

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

func TestProcessPreprocessBatchRecMessage(t *testing.T) {
	// Test with batchsize B = 2 * (n-2t) + 1

	defaultSetup := testutils.DefaultTestSetup()
	n := defaultSetup.OldCommitteeParams.N
	param_t := defaultSetup.OldCommitteeParams.T
	B := 2*(n-2*param_t) + 1

	rScalars := make([]curves.Scalar, B)
	for i := 0; i < B; i++ {
		rScalars[i] = testutils.TestCurve().Scalar.Random(rand.Reader)
	}
	rValues := sharing.CompressScalars(rScalars)

	// Will test with a single node being the dealer for the DPSS round,
	// and initiating this batchrec round.
	testNode := defaultSetup.GetSingleOldNodeFromTestSetup()

	pssID := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		Dealer: testNode.Details(),
		PssID:  common.NewPssID(*pssID),
	}

	testMsg := PreprocessBatchRecMessage{
		PSSRoundDetails: pssRoundDetails,
		Kind:            PreprocessBatchRecMessageType,
		RValues:         rValues,
		CurveName:       testutils.TestCurveName(),
	}

	// secret := sharing.GenerateSecret(curves.K256())
	// commitments, shares, err := sharing.GenerateCommitmentAndShares(
	// 	secret,
	// 	uint32(defaultSetup.OldCommitteeParams.K),
	// 	uint32(n),
	// 	curves.K256(),
	// )
	// testNode.State().ShareStore.Initialize()
	// for i := 0; i < len(shares); i++ {
	// 	testNode.State().ShareStore.OldShares[int(shares[i].Id)] = shares[i].Value
	// }

	testMsg.Process(testNode.Details(), testNode)

}
