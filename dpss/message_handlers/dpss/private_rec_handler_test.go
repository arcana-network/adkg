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
	kOld := defaultSetup.OldCommitteeParams.K
	nOld := defaultSetup.OldCommitteeParams.N

	curve := testutils.TestCurve()

	// Generate a random u_i
	uValue := curve.Scalar.Random(rand.Reader)
	sharingCreator, err := sharing.NewShamir(uint32(kOld), uint32(nOld), curve)
	assert.Nil(t, err)

	uShares, err := sharingCreator.Split(uValue, rand.Reader)
	assert.Nil(t, err)

	points := make(map[int]curves.Scalar)
	for i := 0; i < tOld+1; i++ {
		share, err := curve.Scalar.SetBytes(uShares[i].Value)
		assert.Nil(t, err)
		points[int(uShares[i].Id)] = share
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
	senderNode.Transport.WaitForMessagesSent(nOld)

	assert.Equal(t, nOld, len(senderNode.Transport.GetSentMessages()))
}
