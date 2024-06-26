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
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

// Testing the happy path
func TestLocalComputationProcess(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode := defaultSetup.GetSingleNewNodeFromTestSetup()
	senderNode := defaultSetup.GetSingleOldNodeFromTestSetup()
	senderNode.SetDefaultBatchSize(6)

	testMsg1, testMsg2 := getTestMsgAndUpdateState(receiverNode, senderNode)
	testMsg1.Process(senderNode.Details(), receiverNode)
	testMsg2.Process(senderNode.Details(), receiverNode)
	time.Sleep(2 * time.Second)

	t.Logf("sender=%d", senderNode.Details().Index)
	actualKeyIndexes := []int{}
	expectedKeyIndexes := []int{}
	for k := range receiverNode.GetRefreshedShares() {
		actualKeyIndexes = append(actualKeyIndexes, k)
	}
	sort.Ints(actualKeyIndexes)

	pssIndex := common.GetIndexFromPSSID(testMsg1.DPSSBatchRecDetails.PSSRoundDetails.PssID)
	for i := range append(testMsg1.Coefficients, testMsg2.Coefficients...) {
		keyIndex := (pssIndex * senderNode.DefaultBatchSize()) + i
		expectedKeyIndexes = append(expectedKeyIndexes, keyIndex)
	}

	assert.Equal(t, senderNode.DefaultBatchSize(), len(receiverNode.GetUID()))
	assert.Equal(t, senderNode.DefaultBatchSize(), len(receiverNode.GetRefreshedShares()))
	assert.Equal(t, expectedKeyIndexes, actualKeyIndexes)
}

func TestInvalidUID(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode := defaultSetup.GetSingleNewNodeFromTestSetup()
	senderNode1, senderNode2 := defaultSetup.GetTwoOldNodesFromTestSetup()
	receiverNode.SetDefaultBatchSize(6)
	senderNode1.SetDefaultBatchSize(6)
	senderNode2.SetDefaultBatchSize(6)
	testMsg1, testMsg2 := getTestMsgAndUpdateState(receiverNode, senderNode1)

	wrongUIDS1 := []string{}
	for i := range testMsg1.DPSSBatchRecDetails.PSSRoundDetails.BatchSize {
		wrongUIDS1 = append(wrongUIDS1, fmt.Sprintf("wrong_user:%d", i))
	}
	wrongUIDS2 := []string{}
	for i := range testMsg1.DPSSBatchRecDetails.PSSRoundDetails.BatchSize {
		wrongUIDS2 = append(wrongUIDS2, fmt.Sprintf("wrong_user:%d", i))
	}

	correctUIDs1 := testMsg1.UserIds
	correctUIDs2 := testMsg2.UserIds
	// change to different userIDS so ids doesnt match
	testMsg1.UserIds = wrongUIDS1[:3]
	testMsg2.UserIds = wrongUIDS2[:3]

	testMsg1.Process(senderNode1.Details(), receiverNode)
	testMsg2.Process(senderNode1.Details(), receiverNode)
	time.Sleep(2 * time.Second)

	// uid should not be processed with incorrect data but shares should
	assert.Equal(t, len(receiverNode.GetUID()), 0)
	assert.Equal(t, len(receiverNode.GetRefreshedShares()), senderNode1.DefaultBatchSize())

	// uid should be processed even if share data is already processed
	testMsg1.UserIds = correctUIDs1
	testMsg2.UserIds = correctUIDs2
	testMsg1.Process(senderNode2.Details(), receiverNode)
	testMsg2.Process(senderNode2.Details(), receiverNode)
	time.Sleep(2 * time.Second)

	assert.Equal(t, 6, len(receiverNode.GetUID()))
}

func TestInvalidCoefficients(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode := defaultSetup.GetSingleNewNodeFromTestSetup()
	receiverNode.SetDefaultBatchSize(6)

	senderNode := defaultSetup.GetSingleOldNodeFromTestSetup()
	senderNode.SetDefaultBatchSize(6)

	testMsg1, testMsg2 := getTestMsgAndUpdateState(receiverNode, senderNode)

	// change to different coeff so hash doesnt match
	coeff1 := []string{}
	for range testMsg1.DPSSBatchRecDetails.PSSRoundDetails.BatchSize {
		s := common.CurveFromName(testutils.TestCurveName()).Scalar.Random(rand.Reader)
		coeff1 = append(coeff1, s.BigInt().Text(16))
	}
	coeff2 := []string{}
	for range testMsg1.DPSSBatchRecDetails.PSSRoundDetails.BatchSize {
		s := common.CurveFromName(testutils.TestCurveName()).Scalar.Random(rand.Reader)
		coeff2 = append(coeff2, s.BigInt().Text(16))
	}

	testMsg1.Coefficients = coeff1
	testMsg1.Process(senderNode.Details(), receiverNode)
	testMsg2.Coefficients = coeff2
	testMsg2.Process(senderNode.Details(), receiverNode)

	time.Sleep(2 * time.Second)

	assert.Equal(t, 0, len(receiverNode.GetRefreshedShares()))
}

func getDPSSBatchRecDetails(senderNode *testutils.PssTestNode, batchRecCount int) common.DPSSBatchRecDetails {
	pssID := big.NewInt(0)
	pssRoundDetails := common.PSSRoundDetails{
		Dealer:    senderNode.Details(),
		PssID:     common.NewPssID(*pssID),
		BatchSize: 6,
	}

	dpssBatchRecDetails := common.DPSSBatchRecDetails{
		PSSRoundDetails: pssRoundDetails,
		BatchRecCount:   batchRecCount,
	}

	return dpssBatchRecDetails
}

func getTestMsgAndUpdateState(n *testutils.PssTestNode, sender *testutils.PssTestNode) (LocalComputationMsg, LocalComputationMsg) {
	r := getDPSSBatchRecDetails(sender, 0)
	state, _ := n.State().PSSStore.GetOrSetIfNotComplete(r.PSSRoundDetails.PssID)
	state.Lock()
	defer state.Unlock()
	curve := common.CurveFromName(testutils.TestCurveName())
	state.LocalCompReceived[fmt.Sprintf("%d:%d", 3, 0)] = true
	state.LocalCompReceived[fmt.Sprintf("%d:%d", 3, 1)] = true
	state.LocalCompReceived[fmt.Sprintf("%d:%d", 4, 0)] = true
	state.LocalCompReceived[fmt.Sprintf("%d:%d", 4, 1)] = true

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
	coeff := []curves.Scalar{}
	coeffStr := []string{}
	users := []string{}
	for i := range r.PSSRoundDetails.BatchSize {
		s := curve.Scalar.Random(rand.Reader)
		coeff = append(coeff, s)
		coeffStr = append(coeffStr, s.BigInt().Text(16))
		users = append(users, fmt.Sprintf("user:%d", i))
	}

	hash := getHash(coeff[:3])
	state.LocalComp[0] = &common.LocalComputation{
		Hash:  hash,
		Count: 2,
	}
	hash = getHash(coeff[3:])
	state.LocalComp[1] = &common.LocalComputation{
		Hash:  hash,
		Count: 2,
	}

	for i, u := range users {
		state.UserIDs[u] = &common.LocalComputationUserIDS{
			ID:    i,
			Count: 2,
		}
	}
	msg1 := LocalComputationMsg{
		DPSSBatchRecDetails: getDPSSBatchRecDetails(sender, 0),
		Kind:                LocalComputationMessageType,
		CurveName:           testutils.TestCurveName(),
		Coefficients:        coeffStr[:3],
		UserIds:             users[:3],
		T:                   []int{1, 2, 3, 4, 5},
	}
	msg2 := LocalComputationMsg{
		DPSSBatchRecDetails: getDPSSBatchRecDetails(sender, 1),
		Kind:                LocalComputationMessageType,
		CurveName:           testutils.TestCurveName(),
		Coefficients:        coeffStr[3:],
		UserIds:             users[3:],
		T:                   []int{1, 2, 3, 4, 5},
	}
	return msg1, msg2
}
