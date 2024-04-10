package dpss

import (
	"crypto/rand"
	"math"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

/*
Function: Process

Testcase: happy path.
*/
func TestInitRec(test *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode, dealerNode := defaultSetup.GetTwoOldNodesFromTestSetup()

	recMessage := createInitRecMessage(
		dealerNode,
		*testutils.TestCurve(),
	)

	recMessage.Process(receiverNode.Details(), receiverNode)

	// The ammount of messages sent is n
	n, _, _ := receiverNode.Params()
	receiverNode.Transport.WaitForMessagesSent(n)
	assert.Equal(
		test,
		n,
		len(receiverNode.Transport().GetSentMessages()),
	)
}

/*
Function: Process

Testecase: The sender node and the receiver node are different. Therefore, we
should expect that no message is sent at the end of the process function.
*/
func TestInitRecDiffrerentSender(test *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	nodes := defaultSetup.GetAllOldNodesFromTestSetup()
	dealerNode := nodes[0]
	senderNode := nodes[1]
	receiverNode := nodes[2]

	recMessage := createInitRecMessage(
		dealerNode,
		*testutils.TestCurve(),
	)

	recMessage.Process(senderNode.Details(), receiverNode)

	// The ammount of messages sent is n
	assert.Equal(
		test,
		0,
		len(receiverNode.Transport.GetSentMessages()),
	)
}

// Creates a random InitRecMessage for testing.
func createInitRecMessage(dealerNode *testutils.PssTestNode, curve curves.Curve) InitRecMessage {
	pssID := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		Dealer: dealerNode.Details(),
		PssID:  common.NewPssID(*pssID),
	}
	dpssBatchRecDetails := common.DPSSBatchRecDetails{
		PSSRoundDetails: pssRoundDetails,
		BatchRecCount:   1,
	}

	// Create n - 2t random numbers
	n, _, t := dealerNode.Params()
	shareBatch := make([]curves.Scalar, 0)
	sizeBatch := n - 2*t // Value of B
	nShares := int(math.Ceil(float64(sizeBatch)/float64(n-2*t))) * (n - t)
	for range nShares {
		share := curve.Scalar.Random(rand.Reader)
		shareBatch = append(shareBatch, share)
	}
	shareBatchBytes := sharing.CompressScalars(shareBatch)

	msg := InitRecMessage{
		DPSSBatchRecDetails: dpssBatchRecDetails,
		ShareBatch:          shareBatchBytes,
		Curve:               testutils.TestCurveName(),
		Kind:                InitRecHandlerType,
	}

	return msg
}
