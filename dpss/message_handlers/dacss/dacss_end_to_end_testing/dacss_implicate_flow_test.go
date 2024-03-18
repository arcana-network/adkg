package dacss

import (
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	log "github.com/sirupsen/logrus"
)

// Dealer sends some wrong shares to the nodes, which triggers the Implicate flow

// Detailed situation:
// - dealer in old committee
// - sharing with new committee
// - 2 of the nodes receive a currupted share
// This means we simulate the dealer sending the wrong shares to the nodes in a ProposeMsg

// WIP
func TestTriggerImplicateFlow(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	testSetUp, _ := DefaultTestSetup()
	nNew := testSetUp.OldCommitteeParams.N
	kNew := testSetUp.OldCommitteeParams.K
	// Dealer is a single node from Old committee
	dealer := testSetUp.GetSingleNode(false)

	ephemeralKeypairDealer := common.GenerateKeyPair(testutils.TestCurve())

	secret := sharing.GenerateSecret(testutils.TestCurve())
	commitment, shares, _ := sharing.GenerateCommitmentAndShares(secret, uint32(kNew), uint32(nNew), testutils.TestCurve())

	compressedCommitments := sharing.CompressCommitments(commitment)
	shareMap := make(map[string][]byte, nNew)
	for _, share := range shares {
		// The receiving nodes are in new committee
		nodePublicKey := dealer.GetPublicKeyFor(int(share.Id), true)

		cipherShare, _ := sharing.EncryptSymmetricCalculateKey(
			share.Bytes(),
			nodePublicKey,
			ephemeralKeypairDealer.PrivateKey,
		)

		pubkeyHex := common.PointToHex(nodePublicKey)
		shareMap[pubkeyHex] = cipherShare
	}

	msgData := common.AcssData{
		Commitments:           compressedCommitments,
		ShareMap:              shareMap,
		DealerEphemeralPubKey: common.PointToHex(ephemeralKeypairDealer.PublicKey),
	}

	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)
	msg := dacss.AcssProposeMessage{
		ACSSRoundDetails:   acssRoundDetails,
		Kind:               dacss.AcssProposeMessageType,
		CurveName:          testutils.TestCurveName(),
		Data:               msgData,
		NewCommittee:       true,
		NewCommitteeParams: testSetUp.NewCommitteeParams,
	}

	// Send the ProposeMsg to each node in new committee
	for _, node := range testSetUp.newCommitteeNetwork {
		go func(node *PssTestNode2) {
			msg.Process(dealer.Details(), node)
		}(node)
	}
	time.Sleep(2 * time.Second)
}
