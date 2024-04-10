package dpss

import (
	"testing"

	"github.com/arcana-network/dkgnode/common"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TODO: Incomplete
// Testing the happy path
func TestPublicRecHandlerProcess(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	defaultSetup := testutils.DefaultTestSetup()
	senderNode := defaultSetup.GetSingleOldNodeFromTestSetup()

	testMsg, points, err := getValidPublicRecMsgAndPoints(senderNode, defaultSetup)
	assert.Nil(t, err)

	err = senderNode.State().BatchReconStore.UpdateBatchRecState(
		getDPSSBatchRecDetails(senderNode).ToBatchRecID(),
		func(state *common.BatchRecState) {
			state.ReconstructedUStore = points
		},
	)
	assert.Nil(t, err)
	testMsg.Process(senderNode.Details(), senderNode)

}

func getValidPublicRecMsgAndPoints(senderNode *testutils.PssTestNode, defaultSetup *testutils.TestSetup) (*PublicRecMsg, map[int]curves.Scalar, error) {

	// get points and PrivateMsg from getValidPublicRecMsgAndPoints function
	validPrivateRecMsg, points, err := GetValidPrivateRecMsgAndPoints(senderNode, defaultSetup)

	if err != nil {
		return nil, nil, err
	}

	// considering the shares & Points from the validPrivateRecMsg itself as the "reconstructed shares"
	testMsg := PublicRecMsg{
		DPSSBatchRecDetails: validPrivateRecMsg.DPSSBatchRecDetails,
		Kind:                PublicRecHandlerType,
		curveName:           validPrivateRecMsg.curveName,
		UShare:              validPrivateRecMsg.UShare,
	}

	return &testMsg, points, nil
}
