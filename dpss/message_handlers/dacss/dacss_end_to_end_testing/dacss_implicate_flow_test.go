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
	log.SetLevel(log.InfoLevel)

	testSetUp, _ := DefaultTestSetup()
	nOld := testSetUp.OldCommitteeParams.N
	kOld := testSetUp.OldCommitteeParams.K
	// Dealer is a single node from Old committee
	dealer := testSetUp.GetSingleNode(false)

	ephemeralKeypairDealer := common.GenerateKeyPair(testutils.TestCurve())

	secret := sharing.GenerateSecret(testutils.TestCurve())
	commitment, shares, _ := sharing.GenerateCommitmentAndShares(secret, uint32(kOld), uint32(nOld), testutils.TestCurve())

	compressedCommitments := sharing.CompressCommitments(commitment)
	shareMap := make(map[string][]byte, nOld)
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

	// Corrupting the shareMap, 3 nodes will get the same encrypted shares
	// This will trigger the Implicate flow for 2 of those 3 nodes
	pubkey0Hex, err := testSetUp.newCommitteeNetwork[0].Details().ToHexString(testutils.TestCurveName())
	if err != nil {
		t.Fatal(err)
	}
	pubkey1Hex, _ := testSetUp.newCommitteeNetwork[1].Details().ToHexString(testutils.TestCurveName())
	pubkey2Hex, _ := testSetUp.newCommitteeNetwork[2].Details().ToHexString(testutils.TestCurveName())

	shareNode0 := shareMap[pubkey0Hex]
	shareMap[pubkey1Hex] = shareNode0
	shareMap[pubkey2Hex] = shareNode0

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

	// TODO verify if this works & add assertions
	time.Sleep(15 * time.Second)
}
