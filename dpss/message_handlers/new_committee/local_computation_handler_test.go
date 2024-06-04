package new_committee

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sort"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/stretchr/testify/assert"
)

// Testing the happy path
func TestLocalComputationProcess(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode := defaultSetup.GetSingleNewNodeFromTestSetup()
	senderNode := defaultSetup.GetSingleOldNodeFromTestSetup()

	testMsg := getTestMsgAndUpdateState(receiverNode, senderNode)
	testMsg.Process(senderNode.Details(), receiverNode)

	time.Sleep(2 * time.Second)

	actualKeyIndexes := []int{}
	expectedKeyIndexes := []int{}
	for k := range receiverNode.GetRefreshedShares() {
		actualKeyIndexes = append(actualKeyIndexes, k)
	}
	sort.Ints(actualKeyIndexes)

	pssIndex := common.GetIndexFromPSSID(testMsg.DPSSBatchRecDetails.PSSRoundDetails.PssID)
	for i := range testMsg.coefficients {
		keyIndex := (pssIndex * senderNode.DefaultBatchSize()) + i
		expectedKeyIndexes = append(expectedKeyIndexes, keyIndex)
	}

	assert.Equal(t, len(receiverNode.GetUID()), senderNode.DefaultBatchSize())
	assert.Equal(t, len(receiverNode.GetRefreshedShares()), senderNode.DefaultBatchSize())
	assert.Equal(t, expectedKeyIndexes, actualKeyIndexes)
}

func TestInvalidUID(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode := defaultSetup.GetSingleNewNodeFromTestSetup()
	senderNode1, senderNode2 := defaultSetup.GetTwoOldNodesFromTestSetup()

	testMsg := getTestMsgAndUpdateState(receiverNode, senderNode1)

	wrongUIDS := []string{}
	for i := range testMsg.DPSSBatchRecDetails.PSSRoundDetails.BatchSize {
		wrongUIDS = append(wrongUIDS, fmt.Sprintf("wrong_user:%d", i))
	}

	correctUIDs := testMsg.UserIds
	// change to different userIDS so ids doesnt match
	testMsg.UserIds = wrongUIDS

	testMsg.Process(senderNode1.Details(), receiverNode)
	time.Sleep(2 * time.Second)

	// uid shouldnt be processed with incorrect data but shares should
	assert.Equal(t, len(receiverNode.GetUID()), 0)
	assert.Equal(t, len(receiverNode.GetRefreshedShares()), senderNode1.DefaultBatchSize())

	// uid should be processed even if share data is already processed
	testMsg.UserIds = correctUIDs
	testMsg.Process(senderNode2.Details(), receiverNode)
	time.Sleep(2 * time.Second)

	assert.Equal(t, len(receiverNode.GetUID()), senderNode1.DefaultBatchSize())
}

func TestInvalidCoefficients(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode := defaultSetup.GetSingleNewNodeFromTestSetup()
	senderNode := defaultSetup.GetSingleOldNodeFromTestSetup()

	testMsg := getTestMsgAndUpdateState(receiverNode, senderNode)

	// change to different coeff so hash doesnt match
	coeff := [][]byte{}
	for range testMsg.DPSSBatchRecDetails.PSSRoundDetails.BatchSize {
		coeff = append(coeff, common.CurveFromName(testutils.TestCurveName()).Scalar.Random(rand.Reader).Bytes())
	}

	testMsg.coefficients = coeff
	testMsg.Process(senderNode.Details(), receiverNode)

	time.Sleep(2 * time.Second)

	// If hash doesn't match, shudn't process uid info
	assert.Equal(t, 0, len(receiverNode.GetUID()))
	assert.Equal(t, 0, len(receiverNode.GetRefreshedShares()))

}

func getDPSSBatchRecDetails(senderNode *testutils.PssTestNode) *common.DPSSBatchRecDetails {
	pssID := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		Dealer:    senderNode.Details(),
		PssID:     common.NewPssID(*pssID),
		BatchSize: senderNode.DefaultBatchSize(),
	}

	dpssBatchRecDetails := common.DPSSBatchRecDetails{
		PSSRoundDetails: pssRoundDetails,
		BatchRecCount:   1,
	}

	return &dpssBatchRecDetails
}

func getTestMsgAndUpdateState(n *testutils.PssTestNode, sender *testutils.PssTestNode) LocalComputationMsg {
	r := *getDPSSBatchRecDetails(sender)
	state, _ := n.State().PSSStore.GetOrSetIfNotComplete(r.PSSRoundDetails.PssID)
	state.Lock()
	defer state.Unlock()
	curve := common.CurveFromName(testutils.TestCurveName())
	state.LocalCompReceived[3] = true

	for i := range r.PSSRoundDetails.BatchSize {
		state.KeysetMap[i] = &common.ACSSKeysetMap{
			ShareStore: map[int]*sharing.ShamirShare{},
		}
		for j := range testutils.DefaultN_old {
			state.KeysetMap[i].ShareStore[j] = &sharing.ShamirShare{
				Id:    uint32(n.Details().Index),
				Value: curve.Scalar.Random(rand.Reader).Bytes(),
			}
		}
	}
	coeff := [][]byte{}
	users := []string{}
	for i := range r.PSSRoundDetails.BatchSize {
		coeff = append(coeff, curve.Scalar.Random(rand.Reader).Bytes())
		users = append(users, fmt.Sprintf("user:%d", i))
	}

	hash := getHash(coeff)
	state.LocalComp[hash] = 1

	for _, u := range users {
		state.UserIDs[u] = 1
	}
	msg := LocalComputationMsg{
		DPSSBatchRecDetails: *getDPSSBatchRecDetails(sender),
		Kind:                LocalComputationMessageType,
		curveName:           testutils.TestCurveName(),
		coefficients:        coeff,
		UserIds:             users,
		T:                   []int{1, 2},
	}
	return msg
}
