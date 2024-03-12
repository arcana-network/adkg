package dacss

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	"github.com/coinbase/kryptology/pkg/core/curves"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TODO: Incomplete
func TestDpss(t *testing.T) {
	// timeout := time.After(30 * time.Second)
	// done := make(chan bool)

	curve := curves.ED25519()
	log.SetLevel(log.InfoLevel)

	//default setup and mock transport
	TestSetUp, _ := DefaultTestSetup()

	nodesOld := TestSetUp.oldCommitteeNetwork
	// nodesNew := TestSetUp.newCommitteeNetwork

	nOld := TestSetUp.OldCommitteeParams.N
	kOld := TestSetUp.OldCommitteeParams.K

	//generating random old secrets of old nodes which will be re-shared
	//only one share for each node is considered
	secret := curve.Scalar.Random(rand.Reader)
	_, shares, err := sharing.GenerateCommitmentAndShares(secret, uint32(kOld), uint32(nOld), curve)

	OldSharesOfNodes := make(map[common.NodeDetailsID]*sharing.ShamirShare, nOld)

	for i, n := range nodesOld {

		//convert from kryptoSharing to sharing
		share := sharing.ShamirShare{
			Id:    shares[i].Id,
			Value: shares[i].Value,
		}

		OldSharesOfNodes[n.details.ToNodeDetailsID()] = &share
	}

	assert.Nil(t, err)

	//generating emhemeral keypair for old committee nodes
	ephemeralKeyOfNodes := make(map[common.NodeDetailsID]*common.KeyPair, nOld)

	for _, n := range nodesOld {
		secretKey := curve.Scalar.Random(rand.Reader)
		pubKey := curve.Point.Generator().Mul(secretKey)

		ephemeralKeyOfNodes[n.Details().GetNodeDetailsID()] = &common.KeyPair{
			PublicKey:  pubKey,
			PrivateKey: secretKey,
		}
	}

	id := common.NewPssID(*big.NewInt(int64(1)))

	for _, n := range nodesOld {
		go func(node *PssTestNode2) {
			round := common.PSSRoundDetails{
				PssID:  id,
				Dealer: node.Details(),
			}

			var OldSharesArray []sharing.ShamirShare
			OldShares := OldSharesOfNodes[node.details.GetNodeDetailsID()]

			OldSharesArray = append(OldSharesArray, *OldShares)

			msg, err := dacss.NewInitMessage(
				round,
				OldSharesArray,
				common.CurveName(curve.Name),
				*ephemeralKeyOfNodes[node.details.GetNodeDetailsID()],
			)

			if err != nil {
				log.WithError(err).Error("EndBlock:Acss.NewShareMessage")
			}

			node.ReceiveMessage(node.Details(), *msg)
		}(n)
	}
}
