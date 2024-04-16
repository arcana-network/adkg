package dpss

import (
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dpss"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

func TestDpss(t *testing.T) {
	TestSetUp, transport := DpssIntegrationTestSetup()

	oldNodesNetwork := TestSetUp.OldCommitteeNetwork

	n_old := TestSetUp.OldCommitteeParams.N
	k_old := TestSetUp.OldCommitteeParams.K
	t_old := TestSetUp.OldCommitteeParams.T

	b := 100
	// 1. Generate B secrets
	b_secrets := make([]curves.Scalar, b)
	for i := 0; i < b; i++ {
		b_secrets[i] = sharing.GenerateSecret(testutils.TestCurve())
	}

	// 2. Obtain n_old shares from the B secret & Store shares per node (old committee)
	// Initialize sharestore for each node
	for _, node := range oldNodesNetwork {
		node.State().ShareStore.Initialize(b)
	}

	for i, secret := range b_secrets {
		// Generate commitment and shares
		_, shares, _ := sharing.GenerateCommitmentAndShares(secret, uint32(k_old), uint32(n_old), testutils.TestCurve())
		// Push shares to all nodes
		for j, node := range oldNodesNetwork {
			shareScalar, _ := testutils.TestCurve().Scalar.SetBytes(shares[j].Value)
			node.State().ShareStore.OldShares[i] = shareScalar
		}
	}

	// 3. Generate b/(n-2t) * (n-t) random numbers & Obtain shares for each node for the values
	sharesPerNode := make([][]curves.Scalar, n_old)
	matrixSize := int(math.Ceil(float64(b)/float64(n_old-2*t_old))) * (n_old - t_old)

	for i := 0; i < n_old; i++ {
		sharesPerNode[i] = make([]curves.Scalar, matrixSize)
	}

	for i := 0; i < matrixSize; i++ {
		random_scalar := sharing.GenerateSecret(testutils.TestCurve())
		_, shares, _ := sharing.GenerateCommitmentAndShares(random_scalar, uint32(k_old), uint32(n_old), testutils.TestCurve())
		for j := range oldNodesNetwork {
			shareScalar, _ := testutils.TestCurve().Scalar.SetBytes(shares[j].Value)
			sharesPerNode[j][i] = shareScalar
		}
	}

	// 4. Send DpssHimMessage with the (B / (n - 2*t)) * (n - t)  shares to each node (old committee)
	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: oldNodesNetwork[0].Details(),
	}

	for i, node := range oldNodesNetwork {
		// Per node we have to send
		// (B / (n - 2*t)) * (n - t) shares
		compressedShares := sharing.CompressScalars(sharesPerNode[i])
		msg := dpss.DpssHimMessage{
			PSSRoundDetails: pssRoundDetails,
			Kind:            dpss.DpssHimHandlerType,
			CurveName:       common.CurveName(testutils.TestCurve().Name),
			Shares:          compressedShares,
		}
		msg.Process(node.Details(), node)
	}

	time.Sleep(20 * time.Second)

	// Check step 1: Each HimHandler invocation should send 1 preprocess message
	// in total we expect n PreProcessMessages
	sentMsgs := transport.GetSentMessages()
	preprocessRecMessages := make([]common.PSSMessage, 0)

	for _, msg := range sentMsgs {
		if msg.Type == dpss.PreprocessBatchRecMessageType {
			preprocessRecMessages = append(preprocessRecMessages, msg)
		}
	}
	assert.Equal(t, n_old, len(preprocessRecMessages))

	// Check step 2: Each PreprocessBatchRecMessage should send
	// ceil(B/(n-2t)) InitRecHandlerMessages
	// 34
	nrBatches := math.Ceil(float64(b) / float64(n_old-2*t_old))

	receiveMessageMsgs := transport.GetReceivedMessages()
	assert.True(t, len(receiveMessageMsgs) > 0)
	initRecMessages := make([]common.PSSMessage, 0)

	for _, msg := range receiveMessageMsgs {
		if msg.Type == dpss.InitRecHandlerType {
			initRecMessages = append(initRecMessages, msg)
		}
	}
	nrInitRecMessages := len(initRecMessages)
	// ceil(B/(n-2t)) * n_old
	// 34*7 = 238
	assert.Equal(t, int(nrBatches)*n_old, nrInitRecMessages)

	// Check step 3: From each InitRecHandlerMessage we expect
	// n_old PrivateRecHandlerMessages to be sent
	privateRecMessages := make([]common.PSSMessage, 0)

	for _, msg := range sentMsgs {
		if msg.Type == dpss.PrivateRecMessageType {
			privateRecMessages = append(privateRecMessages, msg)
		}
	}
	// ceil(B/(n-2t)) * n_old InitRecHandlerMessage were sent
	// n_old messages are sent per InitRecHandlerMessage
	// 238 * 7 = 1666
	assert.Equal(t, nrInitRecMessages*n_old, len(privateRecMessages))

	// Filter broadcasted messages on the ones that are of type PublicRecMsg
	publicRecMessages := make([]common.PSSMessage, 0)
	broadcastedMsgs := transport.GetBroadcastedMessages()

	for _, msg := range broadcastedMsgs {
		if msg.Type == dpss.PublicRecMessageType {
			publicRecMessages = append(publicRecMessages, msg)
		}
	}

	// There are ceil(B/(n-2t)) batch rounds
	// There are n_old nodes
	// Each old node should broadcast 1 PublicRecMessage per batch round
	// 34*7 = 238
	assert.Equal(t, int(nrBatches)*n_old, len(publicRecMessages))

	// Final check: the broadcasted LocalComputationMessages
	localComputationMessages := make([]common.PSSMessage, 0)
	for _, msg := range broadcastedMsgs {
		if msg.Type == dpss.LocalComputationMessageType {
			localComputationMessages = append(localComputationMessages, msg)
		}
	}
	// Each node, per batch, broadcasts 1 LocalComputationMessage
	// n_old nodes
	// ceil(B/(n-2t)) nr of batches
	assert.Equal(t, int(nrBatches)*n_old, len(localComputationMessages))

}
