package dpss

import (
	"crypto/rand"
	"errors"
	"math/big"
	"slices"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	ksharing "github.com/coinbase/kryptology/pkg/sharing"
	"github.com/stretchr/testify/assert"
)

/*
Function: Process

Testcase: test that HIM process generates the correct number of shares.

Expectations:
  - There are B globally random shares at the end of the protocol.
  - At the end of the protocol, the participant send to itself a batch reconstruct
    message.
*/
func TestHappyPathHIM(test *testing.T) {
	// Setup the parties
	defaultSetup := testutils.DefaultTestSetup()
	testNode, dealerNode := defaultSetup.GetTwoOldNodesFromTestSetup()
	transport := testNode.Transport

	testNode.State().ShareStore.Initialize()

	n, k, t := testNode.Params()

	shares, err := generateSharesMultipleSecrets(
		n-t,
		dealerNode.Details().Index,
		n,
		k,
		curves.K256(),
	)
	assert.Nil(test, err)

	compressedShares := sharing.CompressShares(shares)

	// Set the round parameters
	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: dealerNode.Details(),
	}

	msg := DacssHimMessage{
		PSSRoundDetails: pssRoundDetails,
		Kind:            DpssHimHandlerType,
		CurveName:       common.CurveName(curves.K256().Name),
		Shares:          compressedShares,
	}

	msg.Process(testNode.Details(), testNode)
	time.Sleep(2 * time.Second)

	assert.Equal(test, 1, len(transport.GetSentMessages()))
}

// Generates the shares for one node for a number of random values. The function
// generates r_1, r_2, ..., r_n and returns the shares [r_1]_i, ...[r_n]_i for
// a node i.
func generateSharesMultipleSecrets(nShares, nodeIdx, n, k int, curve *curves.Curve) ([]curves.Scalar, error) {
	shares := make([]curves.Scalar, 0)
	for range nShares {
		randomScalar := curves.K256().Scalar.Random(rand.Reader)
		_, sharesRandScalar, err := sharing.GenerateCommitmentAndShares(
			randomScalar,
			uint32(k),
			uint32(n),
			curve,
		)
		if err != nil {
			return nil, err
		}

		shareIdx := slices.IndexFunc(
			sharesRandScalar,
			func(share *ksharing.ShamirShare) bool {
				return share.Id == uint32(nodeIdx)
			},
		)
		if shareIdx == -1 {
			return nil, errors.New("Index not found in share slice")
		} else {
			value := sharesRandScalar[shareIdx].Value
			shareScalar, err := curves.K256().Scalar.SetBytes(value)
			if err != nil {
				return nil, err
			}
			shares = append(shares, shareScalar)
		}
	}
	return shares, nil
}
