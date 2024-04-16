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
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Test the happy path
func TestPrivateRecHandlerProcess(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	defaultSetup := testutils.DefaultTestSetup()
	senderNode := defaultSetup.GetSingleOldNodeFromTestSetup()

	testMsg, points, err := GetValidPrivateRecMsgAndPoints(senderNode, defaultSetup)
	assert.Nil(t, err)

	err = senderNode.State().BatchReconStore.UpdateBatchRecState(
		getDPSSBatchRecDetails(senderNode).ToBatchRecID(),
		func(state *common.BatchRecState) {
			state.UStore = points
		},
	)
	assert.Nil(t, err)

	testMsg.Process(senderNode.Details(), senderNode)

	// Wait for all the expected messages to be received
	nOld := defaultSetup.OldCommitteeParams.N
	senderNode.Transport().WaitForMessagesSent(nOld)

	assert.Equal(t, nOld, len(senderNode.Transport().GetSentMessages()))
}

// tests if the shares does not lie on the interpolatng polynomial then early return is triggred
func TestInvalidShare(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	defaultSetup := testutils.DefaultTestSetup()
	senderNode := defaultSetup.GetSingleOldNodeFromTestSetup()

	testMsg, points, err := GetValidPrivateRecMsgAndPoints(senderNode, defaultSetup)
	assert.Nil(t, err)

	// modify the valid points to trigger an early return
	curve := testutils.TestCurve()

	// corrupt the shares
	points[1] = curve.Scalar.Random(rand.Reader)
	points[2] = curve.Scalar.Random(rand.Reader)

	err = senderNode.State().BatchReconStore.UpdateBatchRecState(
		getDPSSBatchRecDetails(senderNode).ToBatchRecID(),
		func(state *common.BatchRecState) {
			state.UStore = points
		},
	)
	assert.Nil(t, err)
	testMsg.Process(senderNode.Details(), senderNode)

	// Wait for all the expected messages to be received
	time.Sleep(2 * time.Second)

	//No msg is send
	assert.Equal(t, 0, len(senderNode.Transport().GetSentMessages()))

}

// test if there is not enough share then it triggers an early return
func TestNotEnoughShare(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	defaultSetup := testutils.DefaultTestSetup()
	senderNode := defaultSetup.GetSingleOldNodeFromTestSetup()
	tOld := defaultSetup.OldCommitteeParams.T
	testMsg, points, err := GetValidPrivateRecMsgAndPoints(senderNode, defaultSetup)
	assert.Nil(t, err)

	// creating insufficient points from the valid points to trigger an early return
	insufficientPoints := make(map[int]curves.Scalar)
	count := 0
	for key, value := range points {
		if count == tOld {
			break
		}
		insufficientPoints[key] = value
		count += 1
	}

	// update the state with insufficient points
	err = senderNode.State().BatchReconStore.UpdateBatchRecState(
		getDPSSBatchRecDetails(senderNode).ToBatchRecID(),
		func(state *common.BatchRecState) {
			state.UStore = insufficientPoints
		},
	)

	assert.Nil(t, err)
	testMsg.Process(senderNode.Details(), senderNode)

	// Wait for all the expected messages to be received
	time.Sleep(2 * time.Second)

	//No msg is send
	assert.Equal(t, 0, len(senderNode.Transport().GetSentMessages()))
}

func getDPSSBatchRecDetails(senderNode *testutils.PssTestNode) *common.DPSSBatchRecDetails {
	pssID := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		Dealer: senderNode.Details(),
		PssID:  common.NewPssID(*pssID),
	}

	dpssBatchRecDetails := common.DPSSBatchRecDetails{
		PSSRoundDetails: pssRoundDetails,
		BatchRecCount:   1,
	}

	return &dpssBatchRecDetails
}

func GetValidPrivateRecMsgAndPoints(senderNode *testutils.PssTestNode, defaultSetup *testutils.TestSetup) (*PrivateRecMsg, map[int]curves.Scalar, error) {

	tOld := defaultSetup.OldCommitteeParams.T
	kOld := defaultSetup.OldCommitteeParams.K
	nOld := defaultSetup.OldCommitteeParams.N

	curve := testutils.TestCurve()

	// Generate a random u_i
	uValue := curve.Scalar.Random(rand.Reader)
	sharingCreator, err := sharing.NewShamir(uint32(kOld), uint32(nOld), curve)

	if err != nil {
		return nil, nil, err
	}

	uShares, err := sharingCreator.Split(uValue, rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	points := make(map[int]curves.Scalar)
	for i := 0; i < tOld+1; i++ {
		share, err := curve.Scalar.SetBytes(uShares[i].Value)

		if err != nil {
			return nil, nil, err
		}

		points[int(uShares[i].Id)] = share
	}

	interpolatePoly, err := common.InterpolatePolynomial(points, curve)
	if err != nil {
		return nil, nil, err
	}

	for i := tOld + 1; i < nOld; i++ {
		value := interpolatePoly.Evaluate(curve.Scalar.New(i))
		points[i] = value
	}

	// valid Private Reconstruction Msg
	testMsg := PrivateRecMsg{
		DPSSBatchRecDetails: *getDPSSBatchRecDetails(senderNode),
		Kind:                PrivateRecMessageType,
		curveName:           testutils.TestCurveName(),
		UShare:              points[senderNode.Details().Index].Bytes(),
	}

	return &testMsg, points, nil
}
