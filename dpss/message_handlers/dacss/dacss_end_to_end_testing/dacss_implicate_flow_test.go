package dacss

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

/*
Testcase: Dealer sends some wrong shares to the nodes, which triggers the Implicate flow

Detailed situation:
- dealer in old committee
- sharing with new committee
- 2 of the nodes receive a currupted share
This means we simulate the dealer sending the wrong shares to the nodes in a ProposeMsg

Expected outcome:
- The nodes that received the wrong shares will be able to recover their share
- the original secret can be reconstructed
- all new nodes have rbcState = Ended
*/
func TestTriggerImplicateFlow(t *testing.T) {
	log.SetLevel(log.InfoLevel)

	testSetUp, _ := DacssIntegrationTestSetup()
	nNew := testSetUp.NewCommitteeParams.N
	kNew := testSetUp.NewCommitteeParams.K
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

	// Corrupting the shareMap, 3 nodes will get the same encrypted shares
	// This will trigger the Implicate flow for 2 of those 3 nodes
	pubkey1Hex, err := testSetUp.newCommitteeNetwork[0].Details().ToHexString(testutils.TestCurveName())
	if err != nil {
		t.Fatal(err)
	}
	pubkey2Hex, _ := testSetUp.newCommitteeNetwork[1].Details().ToHexString(testutils.TestCurveName())
	pubkey3Hex, _ := testSetUp.newCommitteeNetwork[2].Details().ToHexString(testutils.TestCurveName())

	shareNode0 := shareMap[pubkey1Hex]
	// Node 2 & 3 will have to recover their share with help of the other nodes
	shareMap[pubkey2Hex] = shareNode0
	shareMap[pubkey3Hex] = shareNode0

	msgData := common.AcssData{
		Commitments:           compressedCommitments,
		ShareMap:              shareMap,
		DealerEphemeralPubKey: hex.EncodeToString(ephemeralKeypairDealer.PublicKey.ToAffineCompressed()),
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
		go func(node *testutils.IntegrationTestNode) {
			msg.Process(dealer.Details(), node)
		}(node)
	}

	time.Sleep(15 * time.Second)

	var receivedShares []*sharing.ShamirShare
	for _, n := range testSetUp.newCommitteeNetwork {

		state, _, err := n.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
		assert.Nil(t, err)

		rbcState := state.RBCState.Phase
		assert.Equal(t, rbcState, common.Ended)
		share := state.ReceivedShare
		assert.NotNil(t, share)
		receivedShares = append(receivedShares, (*sharing.ShamirShare)(share))
	}

	// Reconstruct secret
	shamir, err := sharing.NewShamir(uint32(kNew), uint32(nNew), testutils.TestCurve())
	assert.Nil(t, err)

	reconstructedSecret, err := shamir.Combine(receivedShares...)
	assert.Nil(t, err)

	assert.Equal(t, secret, reconstructedSecret)

}
