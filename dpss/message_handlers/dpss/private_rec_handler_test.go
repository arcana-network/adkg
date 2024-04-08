package dpss

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

// Test the happy path
func TestPrivateRecHandlerProcess(t *testing.T) {

	defaultSetup := testutils.DefaultTestSetup()
	nodesOld := defaultSetup.GetAllOldNodesFromTestSetup()
	senderNode := nodesOld[0]

	pssID := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		Dealer: senderNode.Details(),
		PssID:  common.NewPssID(*pssID),
	}

	dpssBatchRecDetails := common.DPSSBatchRecDetails{
		PSSRoundDetails: pssRoundDetails,
		BatchRecCount:   1,
	}

	tOld := defaultSetup.OldCommitteeParams.T
	nOld := defaultSetup.OldCommitteeParams.N

	curve := curves.K256()
	points := make(map[int]curves.Scalar)

	// creating random t+1 Scalars
	for i := 0; i < tOld+1; i++ {
		points[i] = curve.Scalar.Random(rand.Reader)
	}

	interpolatePoly, err := common.InterpolatePolynomial(points, curve)

	assert.Nil(t, err)

	for i := tOld + 1; i < nOld; i++ {
		value := interpolatePoly.Evaluate(curve.Scalar.New(i))
		points[i] = value
	}

	err = senderNode.State().BatchReconStore.UpdateBatchRecState(
		dpssBatchRecDetails.ToBatchRecID(),
		func(state *common.BatchRecState) {
			state.UStore = points
		},
	)
	assert.Nil(t, err)

	// valid Private Reconstruction Msg
	testMsg := PrivateRecMsg{
		DPSSBatchRecDetails: dpssBatchRecDetails,
		Kind:                InitRecHandlerType,
		curveName:           testutils.TestCurveName(),
		UShare:              points[senderNode.Details().Index].Bytes(),
	}

	testMsg.Process(senderNode.Details(), senderNode)

	// Wait for all the expected messages to be received
	senderNode.Transport.WaitForMessagesReceived(nOld)

	assert.Equal(t, nOld, len(senderNode.Transport.GetSentMessages()))
}
