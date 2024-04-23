package old_committee

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
	"github.com/torusresearch/bijson"
)

/*
Function: Process

Testcase: happy path

Expectation:
- for B = x*(n-2t)+1, we expect x+1 InitRecMessages to be sent
*/
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
				shareBytes := testutils.TestCurve().Scalar.Random(rand.Reader).Bytes()
				sharePrivKey := common.PrivKeyShare{
					UserIdOwner: "DummyUserID",
					Share: sharing.ShamirShare{
						Id:    uint32(testNode.Details().Index),
						Value: shareBytes,
					},
				}
				testNode.State().ShareStore.OldShares[i] = sharePrivKey
			}

			testMsg.Process(testNode.Details(), testNode)

			// Wait for all the expected messages to be received
			testNode.Transport().WaitForMessagesReceived(tc.nr_batches_minus_1 + 1)

			receivedMsgs := testNode.Transport().ReceivedMessages
			assert.Equal(t, tc.nr_batches_minus_1+1, len(receivedMsgs))
			// Check: len(receivedMsgs)-1 of the receivedMsgs should have batchSize n-2t
			// Check: 1 of the receivedMsgs should have batchSize 1
			batchSizeNMinus2tCount := 0
			batchSizeOneCount := 0

			for _, pssMsg := range receivedMsgs {
				// Deserialize the Data field into an InitRecMessage
				var initRecMsg InitRecMessage
				err := bijson.Unmarshal(pssMsg.Data, &initRecMsg)
				if err != nil {
					t.Fatalf("Failed to unmarshal InitRecMessage: %v", err)
				}

				// Check the BatchSize and increment counters accordingly
				if initRecMsg.BatchSize == n-2*param_t {
					batchSizeNMinus2tCount++
				} else if initRecMsg.BatchSize == 1 {
					batchSizeOneCount++
				} else {
					t.Errorf("Unexpected BatchSize found: %d", initRecMsg.BatchSize)
				}
			}
			assert.Equal(t, tc.nr_batches_minus_1, batchSizeNMinus2tCount, "Unexpected number of messages with BatchSize n-2t")
			assert.Equal(t, 1, batchSizeOneCount, "Unexpected number of messages with BatchSize 1")

		})
	}
}

/*
Function: Process
Testcase: sender is not self
Expectation: early return
*/
func TestSenderNotSelfPreProcess(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	B := 100

	rScalars := make([]curves.Scalar, B)
	for i := 0; i < B; i++ {
		rScalars[i] = testutils.TestCurve().Scalar.Random(rand.Reader)
	}
	rValues := sharing.CompressScalars(rScalars)

	testNode1, testNode2 := defaultSetup.GetTwoOldNodesFromTestSetup()

	pssID := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		Dealer:    testNode2.Details(),
		PssID:     common.NewPssID(*pssID),
		BatchSize: B,
	}

	testMsg := PreprocessBatchRecMessage{
		PSSRoundDetails: pssRoundDetails,
		Kind:            PreprocessBatchRecMessageType,
		RValues:         rValues,
		CurveName:       testutils.TestCurveName(),
	}

	testNode1.State().ShareStore.Initialize(B)
	for i := 0; i < B; i++ {
		shareBytes := testutils.TestCurve().Scalar.Random(rand.Reader).Bytes()
		sharePrivKey := common.PrivKeyShare{
			UserIdOwner: "DummyUserID",
			Share: sharing.ShamirShare{
				Id:    uint32(testNode1.Details().Index),
				Value: shareBytes,
			},
		}
		testNode1.State().ShareStore.OldShares[i] = sharePrivKey
	}

	// Send message from node1 to node2, which triggers early return
	testMsg.Process(testNode1.Details(), testNode2)

	assert.Equal(t, 0, len(testNode1.Transport().ReceivedMessages))
}

/*
Function: Process
Testcase: decompress Rvalues fails (send some random bytes)
Expectation: early return (no messages sent)
*/
func TestRvaluesInvalid(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	B := 35

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
		RValues:         []byte{0, 1, 2}, // Rvalues are some random bytes
		CurveName:       testutils.TestCurveName(),
	}

	testNode.State().ShareStore.Initialize(B)
	for i := 0; i < B; i++ {
		shareBytes := testutils.TestCurve().Scalar.Random(rand.Reader).Bytes()
		sharePrivKey := common.PrivKeyShare{
			UserIdOwner: "DummyUserID",
			Share: sharing.ShamirShare{
				Id:    uint32(testNode.Details().Index),
				Value: shareBytes,
			},
		}
		testNode.State().ShareStore.OldShares[i] = sharePrivKey
	}

	testMsg.Process(testNode.Details(), testNode)

	// Since the Rvalues are invalid, there should be an early return in Process
	assert.Equal(t, 0, len(testNode.Transport().ReceivedMessages))

}

/*
Function: Process
Testcase: not enough shares in state (less than B)
Expectation: early return (no messages sent)
*/
func TestNotEnoughSharesInState(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	B := 100

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

	testNode.State().ShareStore.Initialize(B - 1)
	// Store only B-1 shares in the node's state.
	for i := 0; i < B-1; i++ {
		shareBytes := testutils.TestCurve().Scalar.Random(rand.Reader).Bytes()
		sharePrivKey := common.PrivKeyShare{
			UserIdOwner: "DummyUserID",
			Share: sharing.ShamirShare{
				Id:    uint32(testNode.Details().Index),
				Value: shareBytes,
			},
		}
		testNode.State().ShareStore.OldShares[i] = sharePrivKey
	}

	testMsg.Process(testNode.Details(), testNode)

	// Early return because there are not enough shares in the state
	assert.Equal(t, 0, len(testNode.Transport().ReceivedMessages))
}
