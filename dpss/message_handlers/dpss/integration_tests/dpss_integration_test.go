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

	// 3. Generate n-t random numbers & Obtain shares for each node for the n-t values
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

	time.Sleep(5 * time.Second)

	broadcastedMsgs := transport.GetBroadcastedMessages()
	assert.True(t, len(broadcastedMsgs) > 0)
	privateRecMessages := make([]common.PSSMessage, 0)

	for _, msg := range broadcastedMsgs {
		if msg.Type == dpss.PrivateRecHandlerType {
			privateRecMessages = append(privateRecMessages, msg)
		}
	}
	assert.True(t, len(privateRecMessages) > 0)

	// Filter broadcasted messages on the ones that are of type PublicRecMsg
	publicRecMessages := make([]common.PSSMessage, 0)

	for _, msg := range broadcastedMsgs {
		if msg.Type == dpss.PublicRecHandlerType {
			publicRecMessages = append(publicRecMessages, msg)
		}
	}
	assert.True(t, len(publicRecMessages) > 0)

	// It doesn't reach the Public Rec Handler
	// TODO continue debugging and add assertions
}
