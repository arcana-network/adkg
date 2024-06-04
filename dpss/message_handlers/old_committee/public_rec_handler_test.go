package old_committee

import (
	"crypto/rand"
	"strconv"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/new_committee"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/torusresearch/bijson"
)

// Testing the happy path
func TestPublicRecHandlerProcess(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	defaultSetup := testutils.DefaultTestSetup()
	senderNode := defaultSetup.GetSingleOldNodeFromTestSetup()

	testMsg, points, err := getValidPublicRecMsgAndPoints(senderNode, defaultSetup)
	assert.Nil(t, err)

	nOld, kOld, _ := senderNode.Params()

	// Generate dummy shares for the node. Those shares correspond to shares
	// provided in the init handler. That means that those shares are shares
	// of private keys.
	shares, err := generateSharesMultipleSecrets(
		100,
		senderNode.Details().Index,
		nOld,
		kOld,
		testutils.TestCurve(),
	)
	assert.Nil(t, err)

	// Initialize the sharing store
	senderNode.State().ShareStore.Initialize(len(shares))

	// Store the shares of private keys in the local storage.
	for i, share := range shares {
		senderNode.State().ShareStore.OldShares[i] = common.PrivKeyShare{
			UserIdOwner: "DummyID" + strconv.Itoa(i),
			Share: sharing.ShamirShare{
				Id:    uint32(senderNode.Details().Index),
				Value: share.Bytes(),
			},
		}
	}

	err = senderNode.State().BatchReconStore.UpdateBatchRecState(
		getDPSSBatchRecDetails(senderNode).ToBatchRecID(),
		func(state *common.BatchRecState) {
			state.ReconstructedUStore = points
		},
	)
	assert.Nil(t, err)

	userIdsInternalState := senderNode.State().ShareStore.GetUserIDs()

	testMsg.Process(senderNode.Details(), senderNode)

	senderNode.Transport().WaitForBroadcastSent(1)

	assert.Equal(t, 1, len(senderNode.Transport().GetBroadcastedMessages()))

	// Check that the message sent in the PublicRecHandler contains the
	// user IDs in the state of the sender node.
	for _, message := range senderNode.Transport().GetBroadcastedMessages() {
		// Get the information of the message.
		var localCompMsg new_committee.LocalComputationMsg
		err := bijson.Unmarshal(message.Data, &localCompMsg)
		assert.Nil(t, err)

		assert.Equal(t, localCompMsg.UserIds, userIdsInternalState)
	}

	// Check that the state is cleaned
	assert.Empty(t, senderNode.State().ShareStore.NewShares)
	assert.Empty(t, senderNode.State().ShareStore.OldShares)
	_, found, err := senderNode.State().BatchReconStore.Get(
		getDPSSBatchRecDetails(senderNode).ToBatchRecID(),
	)
	assert.Nil(t, err)
	assert.False(t, found)
}

func getValidPublicRecMsgAndPoints(senderNode *testutils.PssTestNode, defaultSetup *testutils.TestSetup) (*PublicRecMsg, map[int]curves.Scalar, error) {

	// get points and PrivateMsg from getValidPublicRecMsgAndPoints function
	validPrivateRecMsg, points, err := GetValidPrivateRecMsgAndPoints(senderNode, defaultSetup)

	if err != nil {
		return nil, nil, err
	}

	// considering the shares & Points from the validPrivateRecMsg itself as the "reconstructed shares"
	// for unit testing
	testMsg := PublicRecMsg{
		DPSSBatchRecDetails: validPrivateRecMsg.DPSSBatchRecDetails,
		Kind:                PublicRecMessageType,
		curveName:           validPrivateRecMsg.curveName,
		ReconstructedUShare: validPrivateRecMsg.UShare,
	}

	return &testMsg, points, nil
}

func TestInvalidShares(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	defaultSetup := testutils.DefaultTestSetup()
	senderNode := defaultSetup.GetSingleOldNodeFromTestSetup()

	testPrivateMsg, points, err := GetValidPrivateRecMsgAndPoints(senderNode, defaultSetup)
	assert.Nil(t, err)

	// modify the valid points to trigger an early return
	curve := testutils.TestCurve()

	// corrupt the shares
	points[2] = curve.Scalar.Random(rand.Reader)
	points[3] = curve.Scalar.Random(rand.Reader)

	err = senderNode.State().BatchReconStore.UpdateBatchRecState(
		getDPSSBatchRecDetails(senderNode).ToBatchRecID(),
		func(state *common.BatchRecState) {
			state.ReconstructedUStore = points
		},
	)
	assert.Nil(t, err)
	testPublicMsg := PublicRecMsg{
		DPSSBatchRecDetails: testPrivateMsg.DPSSBatchRecDetails,
		Kind:                PublicRecMessageType,
		curveName:           testPrivateMsg.curveName,
		ReconstructedUShare: testPrivateMsg.UShare,
	}

	testPublicMsg.Process(senderNode.Details(), senderNode)

	// Given that there is no sent message, we cannot use the signal/channel
	// strategy. We need to wait.
	time.Sleep(2 * time.Second)

	//No msg is send
	assert.Equal(t, 0, len(senderNode.Transport().GetSentMessages()))

}

func TestNotEnoughShares(t *testing.T) {

	log.SetLevel(log.DebugLevel)
	defaultSetup := testutils.DefaultTestSetup()
	senderNode := defaultSetup.GetSingleOldNodeFromTestSetup()
	tOld := defaultSetup.OldCommitteeParams.T
	testPrivateMsg, points, err := GetValidPrivateRecMsgAndPoints(senderNode, defaultSetup)
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
			state.ReconstructedUStore = insufficientPoints
		},
	)
	assert.Nil(t, err)

	testPublicMsg := PublicRecMsg{
		DPSSBatchRecDetails: testPrivateMsg.DPSSBatchRecDetails,
		Kind:                PublicRecMessageType,
		curveName:           testPrivateMsg.curveName,
		ReconstructedUShare: testPrivateMsg.UShare,
	}

	testPublicMsg.Process(senderNode.Details(), senderNode)

	// Given that there is no sent message, we cannot use the signal/channel
	// strategy. We need to wait.
	time.Sleep(2 * time.Second)

	// No msg is send
	assert.Equal(t, 0, len(senderNode.Transport().GetSentMessages()))

}
