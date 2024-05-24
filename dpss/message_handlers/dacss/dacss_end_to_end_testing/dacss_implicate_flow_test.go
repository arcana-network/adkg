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

		cipherShare, hmacTag, _ := sharing.EncryptSymmetricCalculateKey(
			share.Bytes(),
			nodePublicKey,
			ephemeralKeypairDealer.PrivateKey,
		)
		//combine the hmac and encrypted shares
		cipherShare = sharing.Combine(cipherShare, hmacTag)
		pubkeyHex := common.PointToHex(nodePublicKey)
		shareMap[pubkeyHex] = cipherShare
	}

	// Corrupting the shareMap, 3 nodes will get the same encrypted shares
	// This will trigger the Implicate flow for 2 of those 3 nodes
	pubkey1Hex, err := testSetUp.NewCommitteeNetwork[0].Details().ToHexString(testutils.TestCurveName())
	if err != nil {
		t.Fatal(err)
	}

	pubkey2Hex, _ := testSetUp.NewCommitteeNetwork[1].Details().ToHexString(testutils.TestCurveName())
	pubkey3Hex, _ := testSetUp.NewCommitteeNetwork[2].Details().ToHexString(testutils.TestCurveName())

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

	//TODO: Probably it's a good idea to Initialize an empty state for all the new and old committee at the beginning of the protocol ie in initHandler??
	// Initialize the empty state for all the new nodes
	for _, node := range testSetUp.NewCommitteeNetwork {
		node.State().AcssStore.UpdateAccsState(
			acssRoundDetails.ToACSSRoundID(),
			func(as *common.AccsState) {},
		)
	}

	// Initialize the empty state for all the old nodes
	for _, node := range testSetUp.OldCommitteeNetwork {
		node.State().AcssStore.UpdateAccsState(
			acssRoundDetails.ToACSSRoundID(),
			func(as *common.AccsState) {},
		)
	}

	// Send the ProposeMsg to each node in new committee
	for _, node := range testSetUp.NewCommitteeNetwork {
		go func(node *testutils.IntegrationTestNode) {
			msg.Process(dealer.Details(), node)
		}(node)
	}

	time.Sleep(8 * time.Second)

	var receivedShares []*sharing.ShamirShare
	for _, n := range testSetUp.NewCommitteeNetwork {

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

/*
Testcase: Dealer sends some wrong shares to the nodes, which triggers the Implicate flow. 1 node doesn't receive shares at all.

Detailed situation:
- dealer in old committee
- sharing with new committee
- 1 of the nodes receives share with a delay (to trigger execution upon storing the implicate information)
- 2 of the nodes receive a currupted share

Expected outcome:
- The node that received the wrong shares will be able to recover their share
- the original secret can be reconstructed
- all new nodes have rbcState = Ended
*/
func TestTriggerImplicateFlow2(t *testing.T) {
	log.SetLevel(log.DebugLevel)

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

		cipherShare, hmacTag, _ := sharing.EncryptSymmetricCalculateKey(
			share.Bytes(),
			nodePublicKey,
			ephemeralKeypairDealer.PrivateKey,
		)
		//combine the hmac and encrypted shares
		cipherShare = sharing.Combine(cipherShare, hmacTag)
		pubkeyHex := common.PointToHex(nodePublicKey)
		shareMap[pubkeyHex] = cipherShare
	}

	// Corrupting the shareMap, 3 nodes will get the same encrypted shares
	// This will trigger the Implicate flow for 2 of those 3 nodes
	pubkey1Hex, err := testSetUp.NewCommitteeNetwork[0].Details().ToHexString(testutils.TestCurveName())
	if err != nil {
		t.Fatal(err)
	}

	pubkey2Hex, _ := testSetUp.NewCommitteeNetwork[1].Details().ToHexString(testutils.TestCurveName())
	pubkey3Hex, _ := testSetUp.NewCommitteeNetwork[2].Details().ToHexString(testutils.TestCurveName())

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

	// Initialize the empty state for all the new nodes
	for _, node := range testSetUp.NewCommitteeNetwork {
		node.State().AcssStore.UpdateAccsState(
			acssRoundDetails.ToACSSRoundID(),
			func(as *common.AccsState) {},
		)
	}

	// Initialize the empty state for all the old nodes
	for _, node := range testSetUp.OldCommitteeNetwork {
		node.State().AcssStore.UpdateAccsState(
			acssRoundDetails.ToACSSRoundID(),
			func(as *common.AccsState) {},
		)
	}

	// Send the ProposeMsg to each node in new committee

	for i, node := range testSetUp.NewCommitteeNetwork {
		if i != len(testSetUp.NewCommitteeNetwork)-1 { // Skip the last node
			go func(node *testutils.IntegrationTestNode) {
				msg.Process(dealer.Details(), node)
			}(node)
		}
	}

	time.Sleep(1 * time.Second)
	log.Info("Send missing ProposeMsg to the last node")
	go msg.Process(dealer.Details(), testSetUp.NewCommitteeNetwork[len(testSetUp.NewCommitteeNetwork)-1])
	time.Sleep(8 * time.Second)
	log.Info("Start assertions in test")
	var receivedShares []*sharing.ShamirShare
	for _, n := range testSetUp.NewCommitteeNetwork {

		state, _, err := n.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
		assert.Nil(t, err)

		rbcState := state.RBCState.Phase
		log.Info("Node ", n.Details().Index)
		assert.Equal(t, common.Ended, rbcState)
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

/*
Testcase: Dealer sends 2 wrong shares: to 1 node immediately and to 1 node with a small delay

Detailed situation:
- dealer in old committee
- sharing with new committee
- 1 of the nodes receives a corrupt share with a delay
- 1 of the nodes receives a currupted share

Expected outcome:
- The node that received the wrong shares will be able to recover their share
- the original secret can be reconstructed
- all new nodes have rbcState = Ended

This is expected to fail because we need to fix
*/
func TestTriggerImplicateFlow3(t *testing.T) {
	log.SetLevel(log.DebugLevel)

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

		cipherShare, hmacTag, _ := sharing.EncryptSymmetricCalculateKey(
			share.Bytes(),
			nodePublicKey,
			ephemeralKeypairDealer.PrivateKey,
		)
		//combine the hmac and encrypted shares
		cipherShare = sharing.Combine(cipherShare, hmacTag)
		pubkeyHex := common.PointToHex(nodePublicKey)
		shareMap[pubkeyHex] = cipherShare
	}

	// Corrupting the shareMap, 3 nodes will get the same encrypted shares
	// This will trigger the Implicate flow for 2 of those 3 nodes
	pubkey1Hex, err := testSetUp.NewCommitteeNetwork[0].Details().ToHexString(testutils.TestCurveName())
	if err != nil {
		t.Fatal(err)
	}

	pubkey2Hex, _ := testSetUp.NewCommitteeNetwork[1].Details().ToHexString(testutils.TestCurveName())
	pubkey3Hex, _ := testSetUp.NewCommitteeNetwork[2].Details().ToHexString(testutils.TestCurveName())

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

	// Initialize the empty state for all the new nodes
	for _, node := range testSetUp.NewCommitteeNetwork {
		node.State().AcssStore.UpdateAccsState(
			acssRoundDetails.ToACSSRoundID(),
			func(as *common.AccsState) {},
		)
	}

	// Initialize the empty state for all the old nodes
	for _, node := range testSetUp.OldCommitteeNetwork {
		node.State().AcssStore.UpdateAccsState(
			acssRoundDetails.ToACSSRoundID(),
			func(as *common.AccsState) {},
		)
	}

	// Send the ProposeMsg to each node in new committee

	for i, node := range testSetUp.NewCommitteeNetwork {
		if i != len(testSetUp.NewCommitteeNetwork) { // Skip node with index 1
			if i != 1 {
				go func(node *testutils.IntegrationTestNode) {
					msg.Process(dealer.Details(), node)
				}(node)
			}
		}
	}

	time.Sleep(1 * time.Second)
	log.Info("Send missing ProposeMsg to the last node")
	go msg.Process(dealer.Details(), testSetUp.NewCommitteeNetwork[1])
	time.Sleep(8 * time.Second)
	log.Info("Start assertions in test")
	var receivedShares []*sharing.ShamirShare
	for _, n := range testSetUp.NewCommitteeNetwork {

		state, _, err := n.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
		assert.Nil(t, err)

		rbcState := state.RBCState.Phase
		log.Info("Node ", n.Details().Index)
		assert.Equal(t, common.Ended, rbcState)
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
