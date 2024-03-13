package dacss

import (
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

/*
Function: Process

Testcase: node receives implicate message and already has the necessary shareMap

Expectations:
- node sends ImplicateExecute message
- nothing stored in ImplicateInformation of node's state
*/
func TestAlreadyHasShareMap(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

	// The dealer creates the shareMap
	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())
	n, k, _ := dealer.Params()
	curve := curves.K256()

	secret := sharing.GenerateSecret(curves.K256())
	commitment, shares, err := sharing.GenerateCommitmentAndShares(
		secret,
		uint32(k),
		uint32(n),
		curves.K256(),
	)
	if err != nil {
		log.Errorf("Error generating commitments & shares: %v", err)
	}
	compressedCommitments := sharing.CompressCommitments(commitment)
	shareMap := make(map[string][]byte, n)
	for _, share := range shares {
		nodePublicKey := dealer.GetPublicKeyFor(int(share.Id), dealer.IsNewNode())

		cipherShare, err := sharing.EncryptSymmetricCalculateKey(
			share.Bytes(),
			nodePublicKey,
			ephemeralKeypairDealer.PrivateKey,
		)

		if err != nil {
			log.Errorf("Error while encrypting secret share, err=%v", err)
		}
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		pubkeyHex := common.PointToHex(nodePublicKey)
		shareMap[pubkeyHex] = cipherShare
	}
	msgData := common.AcssData{
		Commitments:           compressedCommitments,
		ShareMap:              shareMap,
		DealerEphemeralPubKey: common.PointToHex(ephemeralKeypairDealer.PublicKey),
	}

	// Set the round parameters
	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)

	// Assume the shareMap for this round was already received and stored in the Node 1's state
	hash, _ := common.HashAcssData(msgData)
	err = node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = hash
	})
	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
	}

	// Assume in Node 2 the Implicate msg is created
	implicateMsg := createProofAndImplicateMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails)

	// Node 1 receives Implicate msg
	implicateMsg.Process(node2.Details(), node1)

	// In this part we don't actually check whether the Predicate indeed fails, that happens in the ImplicateExecute
	// Check 1: the node itself receives the implicateExecuteMessage
	receivedMsgNode1 := node1.Transport.ReceivedMessages
	assert.Equal(t, 1, len(receivedMsgNode1))
	assert.Equal(t, ImplicateExecuteMessageType, receivedMsgNode1[0].Type)

	// Check 2: *NO* ImplicateInformation is added to Node's state
	acssState, _, _ := node1.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
	assert.True(t, len(acssState.ImplicateInformationSlice) == 0)
}

/*
Function: Process

Testcase: node receives implicate message and doesn't have the necessary shareMap

Expectation:
- node stores the implicate information in the state (to be picked up by ProposeHandler later)
- no msg sent
*/
func TestDoesntHaveShareMap(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()
	// Set the round parameters
	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)

	// Ephemeral pubkey of dealer
	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())
	curve := curves.K256()

	// Assume in Node 2 the Implicate msg is created
	implicateMsg := createProofAndImplicateMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails)

	// Node 1 receives Implicate msg
	// It does not have the shareMap stored. So it shouldn't send an ImplicateExecute message

	// Check 1: no ImplicateExecute message is sent
	implicateMsg.Process(node2.Details(), node1)
	receivedMsgNode1 := node1.Transport.ReceivedMessages
	assert.Equal(t, 0, len(receivedMsgNode1))

	// Check 2: ImplicateInformation is added to Node's state
	acssState, _, _ := node1.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
	assert.True(t, len(acssState.ImplicateInformationSlice) > 0)
}

/*
Function: Process

Testcase: Share recovery is ongoing (ShareRecoveryOngoing is true in Node's state)

Expectation: early return. Nothing gets stored, nothing gets sent.
*/
func TestAlreadyInShareRecovery(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()
	// Set the round parameters
	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)
	// Ephemeral pubkey of dealer
	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())
	curve := curves.K256()

	err := node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.ShareRecoveryOngoing = true
	})
	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
	}

	// Assume in Node 2 the Implicate msg is created
	implicateMsg := createProofAndImplicateMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails)

	// Node 1 receives Implicate msg
	// Does not have the shareMap stored. So it shouldn't send an ImplicateExecute message
	// Check 1: no ImplicateExecute message is sent
	implicateMsg.Process(node2.Details(), node1)
	receivedMsgNode1 := node1.Transport.ReceivedMessages
	assert.Equal(t, 0, len(receivedMsgNode1))

	// Check 2: *NO* ImplicateInformation is added to Node's state
	acssState, _, _ := node1.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
	assert.True(t, len(acssState.ImplicateInformationSlice) == 0)
}

func createProofAndImplicateMsg(ephemeralKeypairDealer common.KeyPair, node2 *testutils.PssTestNode, curve *curves.Curve, acssRoundDetails common.ACSSRoundDetails) ImplicateReceiveMessage {
	symmKey, _ := sharing.CalculateSharedKey(ephemeralKeypairDealer.PublicKey, node2.PrivateKey())
	proof := sharing.GenerateNIZKProof(curve, node2.LongtermKey.PrivateKey, node2.LongtermKey.PublicKey,
		ephemeralKeypairDealer.PublicKey, symmKey, curve.NewGeneratorPoint())

	implicateMsg := createImplicateMsg(acssRoundDetails, symmKey, proof)
	return implicateMsg
}

func createImplicateMsg(acssRoundDetails common.ACSSRoundDetails, symmKey2 curves.Point, proof []byte) ImplicateReceiveMessage {
	implicateMsg := ImplicateReceiveMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ImplicateReceiveMessageType,
		CurveName:        common.SECP256K1,
		SymmetricKey:     symmKey2.ToAffineCompressed(),
		Proof:            proof,
	}
	return implicateMsg
}
