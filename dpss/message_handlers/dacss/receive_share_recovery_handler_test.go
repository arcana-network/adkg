package dacss

import (
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

/*
ReceiveShareRecoveryHandler is WIP, it needs to store share in the state


- Receive 1 share: TestReceiveFirstShareRecovery
- Receive t+1 shares: TestReceiveThresholdShareRecoveryMsgs
- sender self: TestReceiverEqualsSelf
- no acssState found: TestNoAcssState
- no share recovery going on: TestShareRecoveryOngoingFalse
- node already has a valid share (ValidShareOutput true): TestValidShareOutputTrue
- hash of msg.acssData not equal to stored hash: TestValidShareOutputTrue
- incorrect proof format: TestAcssRecoveryDataHashMismatch
- dealerPubkey can't be converted to curve point: TestDealerPubkeyWrongConversion
- symmetric key can't be converted to curve point: TestSymmetricKeyWrongConversion
- proof verification fails: TestProofVerficationFails
- predicate for share fails: TestPredicateForShareFails
*/

/*
Function: Process

Testcase: node receives first share recovery message

Expectations:
- stores ShamirShare in state
- bool ValidShareOutput should not be valid, as t+1 shares are needed for that
*/
func TestReceiveFirstShareRecovery(t *testing.T) {
	// Node 1 is sending their symm key + proof, so node 2 can use it to recover share
	node1, node2, msg := setupSingleRecoveryMsg()

	acssState, _, _ := node2.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.False(t, len(acssState.VerifiedRecoveryShares) > 0)

	msg.Process(node1.Details(), node2)
	time.Sleep(100 * time.Millisecond)
	acssState, _, _ = node2.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())

	assert.False(t, acssState.ValidShareOutput)
	assert.True(t, len(acssState.VerifiedRecoveryShares) > 0)
}

/*
Function: Process

Testcase: node receives t+1 share recovery messages

Expectations:
- bool ValidShareOutput should be valid
- len VerifiedRecoveryShares should be t+1
*/
func TestReceiveThresholdShareRecoveryMsgs(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	oldParams := defaultSetup.OldCommitteeParams
	param_t := oldParams.T
	// we need t+1+1 nodes to test, so 1 node can receive t+1 msgs
	testNodes := defaultSetup.GetXOldCommitteeNodes(param_t + 2)

	receiver := testNodes[0]
	dealer := testNodes[1]

	ephemeralKeypairDealer, msgData, acssRoundDetails := getMsgDataAndStoreHash(dealer, receiver)
	// Receiver node is in Share Recovery phase
	receiver.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.ShareRecoveryOngoing = true
	})

	for i := 1; i < len(testNodes); i++ {
		currentNode := testNodes[i]

		symmKey, _ := sharing.CalculateSharedKey(ephemeralKeypairDealer.PublicKey, currentNode.PrivateKey())
		proof := sharing.GenerateNIZKProof(testutils.TestCurve(), currentNode.LongtermKey.PrivateKey, currentNode.LongtermKey.PublicKey,
			ephemeralKeypairDealer.PublicKey, symmKey, testutils.TestCurve().NewGeneratorPoint())

		msg := ReceiveShareRecoveryMessage{
			ACSSRoundDetails: acssRoundDetails,
			Kind:             ShareMessageType,
			CurveName:        testutils.TestCurveName(),
			SymmetricKey:     symmKey.ToAffineCompressed(),
			Proof:            proof,
			AcssData:         msgData,
		}
		// As long as the threshold is not reached, bool ValidShareOutput should not be valid
		acssState, _, _ := receiver.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
		assert.False(t, acssState.ValidShareOutput)

		msg.Process(currentNode.Details(), receiver)
	}
	time.Sleep(10 * time.Second)
	acssState, _, _ := receiver.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())

	// Check: bool ValidShareOutput should be valid & len VerifiedRecoveryShares should t+1
	assert.True(t, acssState.ValidShareOutput)
	assert.True(t, len(acssState.VerifiedRecoveryShares) == param_t+1)

	receiver.Transport().WaitForBroadcastSent(1)
	broadcastedMsg := receiver.Transport().BroadcastedMessages
	//should broadcast the commit msg to the new committee
	assert.Equal(t, len(broadcastedMsg), 1)
}

/*
Function: Process

Testcase: node receives share recovery message from itself

Expectations: nothing should be stored in VerifiedRecoveryShares
*/
func TestReceiverEqualsSelf(t *testing.T) {
	_, node2, msg := setupSingleRecoveryMsg()

	acssState, _, _ := node2.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.False(t, len(acssState.VerifiedRecoveryShares) > 0)

	msg.Process(node2.Details(), node2)
	time.Sleep(100 * time.Millisecond)
	acssState, _, _ = node2.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.False(t, len(acssState.VerifiedRecoveryShares) > 0)
}

/*
Function: Process

Testcase: no acssState found

Expectations: nothing should be stored in VerifiedRecoveryShares
*/
func TestNoAcssState(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, receiver := defaultSetup.GetThreeOldNodesFromTestSetup()

	ephemeralKeypairDealer, msgData, acssRoundDetails := getMsgData(dealer)

	symmKey, _ := sharing.CalculateSharedKey(ephemeralKeypairDealer.PublicKey, node1.PrivateKey())
	proof := sharing.GenerateNIZKProof(testutils.TestCurve(), node1.LongtermKey.PrivateKey, node1.LongtermKey.PublicKey,
		ephemeralKeypairDealer.PublicKey, symmKey, testutils.TestCurve().NewGeneratorPoint())

	msg := ReceiveShareRecoveryMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ShareMessageType,
		CurveName:        testutils.TestCurveName(),
		SymmetricKey:     symmKey.ToAffineCompressed(),
		Proof:            proof,
		AcssData:         msgData,
	}

	msg.Process(node1.Details(), receiver)
	time.Sleep(100 * time.Millisecond)
	_, found, _ := receiver.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.False(t, found)
}

/*
Function: Process

Testcase: node is not in Share Recovery phase (ShareRecoveryOngoing is false)

Expectations: nothing should be stored in VerifiedRecoveryShares
*/
func TestShareRecoveryOngoingFalse(t *testing.T) {
	node1, receiver, msg := setupSingleRecoveryMsg()

	receiver.State().AcssStore.UpdateAccsState(msg.ACSSRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.ShareRecoveryOngoing = false
	})

	msg.Process(node1.Details(), receiver)
	time.Sleep(2 * time.Second)
	acssState, _, _ := receiver.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.True(t, len(acssState.VerifiedRecoveryShares) == 1)
}

/*
Function: Process

Testcase: node already has a valid share (ValidShareOutput true)

Expectations: nothing should be stored in VerifiedRecoveryShares
*/
func TestValidShareOutputTrue(t *testing.T) {
	node1, receiver, msg := setupSingleRecoveryMsg()

	acssState, _, _ := receiver.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.False(t, len(acssState.VerifiedRecoveryShares) > 0)

	receiver.State().AcssStore.UpdateAccsState(msg.ACSSRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = []byte("hash")
	})

	msg.Process(node1.Details(), receiver)
	time.Sleep(100 * time.Millisecond)
	acssState, _, _ = receiver.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.True(t, len(acssState.VerifiedRecoveryShares) == 0)
}

/*
Function: Process

Testcase: provided proof cannot be unpacked (wrong format)

Expectations: nothing should be stored in VerifiedRecoveryShares
*/
func TestIncorrectProofFormat(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, receiver := defaultSetup.GetThreeOldNodesFromTestSetup()

	ephemeralKeypairDealer, msgData, acssRoundDetails := getMsgData(dealer)

	hash, _ := common.HashAcssData(msgData)
	err := receiver.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = hash
		state.ShareRecoveryOngoing = true
	})
	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
	}
	symmKey, _ := sharing.CalculateSharedKey(ephemeralKeypairDealer.PublicKey, node1.PrivateKey())
	proof := []byte("invalid proof")

	msg := ReceiveShareRecoveryMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ShareMessageType,
		CurveName:        testutils.TestCurveName(),
		SymmetricKey:     symmKey.ToAffineCompressed(),
		Proof:            proof,
		AcssData:         msgData,
	}

	msg.Process(node1.Details(), receiver)
	time.Sleep(100 * time.Millisecond)
	acssState, _, _ := receiver.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.True(t, len(acssState.VerifiedRecoveryShares) == 0)
}

/*
Function: Process

Testcase: hash of msg.acssData not equal to stored hash

Expectations: nothing should be stored in VerifiedRecoveryShares
*/
func TestAcssRecoveryDataHashMismatch(t *testing.T) {
	node1, node2, msg := setupSingleRecoveryMsg()
	// This will cause the hash of received acssData to be distinct from the stored hash
	msg.AcssData.DealerEphemeralPubKey = "0x123"

	msg.Process(node1.Details(), node2)
	time.Sleep(100 * time.Millisecond)
	acssState, _, _ := node2.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.False(t, len(acssState.VerifiedRecoveryShares) > 0)
}

/*
Function: Process

Testcase: dealerPubkey (in acssData) can't be converted to curve point

Expectations: nothing should be stored in VerifiedRecoveryShares
*/
func TestDealerPubkeyWrongConversion(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, receiver := defaultSetup.GetThreeOldNodesFromTestSetup()
	ephemeralKeypairDealer := common.GenerateKeyPair(testutils.TestCurve())
	n, k, _ := dealer.Params()

	secret := sharing.GenerateSecret(testutils.TestCurve())
	commitment, shares, err := sharing.GenerateCommitmentAndShares(
		secret,
		uint32(k),
		uint32(n),
		testutils.TestCurve(),
	)
	if err != nil {
		log.Errorf("Error generating commitments & shares: %v", err)
	}
	compressedCommitments := sharing.CompressCommitments(commitment)
	shareMap := make(map[string][]byte, n)
	for _, share := range shares {
		nodePublicKey := dealer.GetPublicKeyFor(int(share.Id), dealer.IsNewNode())

		cipherShare, _, err := sharing.EncryptSymmetricCalculateKey(
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
		DealerEphemeralPubKey: "0x1234", // Set invalid pubkey, so conversion fails
	}

	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)

	hash, _ := common.HashAcssData(msgData)
	receiver.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = hash
		// make sure node is in share recovery
		state.ShareRecoveryOngoing = true
	})
	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
	}
	symmKey, _ := sharing.CalculateSharedKey(ephemeralKeypairDealer.PublicKey, node1.PrivateKey())
	proof := sharing.GenerateNIZKProof(testutils.TestCurve(), node1.LongtermKey.PrivateKey, node1.LongtermKey.PublicKey,
		ephemeralKeypairDealer.PublicKey, symmKey, testutils.TestCurve().NewGeneratorPoint())

	msg := ReceiveShareRecoveryMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ShareMessageType,
		CurveName:        testutils.TestCurveName(),
		SymmetricKey:     symmKey.ToAffineCompressed(),
		Proof:            proof,
		AcssData:         msgData,
	}
	msg.Process(node1.Details(), receiver)
	time.Sleep(100 * time.Millisecond)
	acssState, _, _ := receiver.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.False(t, len(acssState.VerifiedRecoveryShares) > 0)
}

/*
Function: Process

Testcase: symmetric key can't be converted to curve point

Expectations: nothing should be stored in VerifiedRecoveryShares
*/
func TestSymmetricKeyWrongConversion(t *testing.T) {
	node1, node2, msg := setupSingleRecoveryMsg()

	acssState, _, _ := node2.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.False(t, len(acssState.VerifiedRecoveryShares) > 0)
	// Setting invalid symmetric key so it can't be converted to curve point
	msg.SymmetricKey = []byte("invalid key")

	msg.Process(node1.Details(), node2)
	time.Sleep(100 * time.Millisecond)
	acssState, _, _ = node2.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.False(t, len(acssState.VerifiedRecoveryShares) > 0)
}

/*
Function: Process

Testcase: proof verification fails

Expectations: nothing should be stored in VerifiedRecoveryShares
*/
func TestProofVerficationFails(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, receiver := defaultSetup.GetThreeOldNodesFromTestSetup()

	ephemeralKeypairDealer, msgData, acssRoundDetails := getMsgDataAndStoreHash(dealer, receiver)

	// Receiver node is in Share Recovery phase
	receiver.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.ShareRecoveryOngoing = true
	})

	symmKey, _ := sharing.CalculateSharedKey(ephemeralKeypairDealer.PublicKey, node1.PrivateKey())
	// using a wrong key for proof so verification will fail
	proof := sharing.GenerateNIZKProof(testutils.TestCurve(), dealer.LongtermKey.PrivateKey, node1.LongtermKey.PublicKey,
		ephemeralKeypairDealer.PublicKey, symmKey, testutils.TestCurve().NewGeneratorPoint())

	msg := ReceiveShareRecoveryMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ShareMessageType,
		CurveName:        testutils.TestCurveName(),
		SymmetricKey:     symmKey.ToAffineCompressed(),
		Proof:            proof,
		AcssData:         msgData,
	}
	msg.Process(node1.Details(), receiver)
	time.Sleep(100 * time.Millisecond)
	acssState, _, _ := receiver.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.False(t, len(acssState.VerifiedRecoveryShares) > 0)
}

/*
Function: Process

Testcase: predicate for share fails

Expectations: nothing should be stored in VerifiedRecoveryShares
*/
func TestPredicateForShareFails(t *testing.T) {
	// All shares get encrypted under dealers key, so predicate will fail

	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, receiver := defaultSetup.GetThreeOldNodesFromTestSetup()

	ephemeralKeypairDealer := common.GenerateKeyPair(testutils.TestCurve())
	n, k, _ := dealer.Params()

	secret := sharing.GenerateSecret(testutils.TestCurve())
	commitment, shares, err := sharing.GenerateCommitmentAndShares(
		secret,
		uint32(k),
		uint32(n),
		testutils.TestCurve(),
	)
	if err != nil {
		log.Errorf("Error generating commitments & shares: %v", err)
	}
	compressedCommitments := sharing.CompressCommitments(commitment)
	shareMap := make(map[string][]byte, n)
	for _, share := range shares {
		cipherShare, _, err := sharing.EncryptSymmetricCalculateKey(
			share.Bytes(),
			// All shares get encrypted under dealers key, so predicate will fail
			ephemeralKeypairDealer.PublicKey,
			ephemeralKeypairDealer.PrivateKey,
		)

		if err != nil {
			log.Errorf("Error while encrypting secret share, err=%v", err)
		}
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		nodePublicKey := dealer.GetPublicKeyFor(int(share.Id), dealer.IsNewNode())
		pubkeyHex := common.PointToHex(nodePublicKey)
		shareMap[pubkeyHex] = cipherShare
	}

	msgData := common.AcssData{
		Commitments:           compressedCommitments,
		ShareMap:              shareMap,
		DealerEphemeralPubKey: common.PointToHex(ephemeralKeypairDealer.PublicKey),
	}

	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)
	hash, _ := common.HashAcssData(msgData)
	receiver.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = hash
		state.ShareRecoveryOngoing = true
	})
	symmKey, _ := sharing.CalculateSharedKey(ephemeralKeypairDealer.PublicKey, node1.PrivateKey())
	proof := sharing.GenerateNIZKProof(testutils.TestCurve(), node1.LongtermKey.PrivateKey, node1.LongtermKey.PublicKey,
		ephemeralKeypairDealer.PublicKey, symmKey, testutils.TestCurve().NewGeneratorPoint())

	msg := ReceiveShareRecoveryMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ShareMessageType,
		CurveName:        testutils.TestCurveName(),
		SymmetricKey:     symmKey.ToAffineCompressed(),
		Proof:            proof,
		AcssData:         msgData,
	}
	msg.Process(node1.Details(), receiver)
	time.Sleep(100 * time.Millisecond)
	acssState, _, _ := receiver.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	assert.False(t, len(acssState.VerifiedRecoveryShares) > 0)
}

func setupSingleRecoveryMsg() (*testutils.PssTestNode, *testutils.PssTestNode, ReceiveShareRecoveryMessage) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, receiver := defaultSetup.GetThreeOldNodesFromTestSetup()

	ephemeralKeypairDealer, msgData, acssRoundDetails := getMsgDataAndStoreHash(dealer, receiver)

	// Receiver node is in Share Recovery phase
	receiver.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.ShareRecoveryOngoing = true
	})

	symmKey, _ := sharing.CalculateSharedKey(ephemeralKeypairDealer.PublicKey, node1.PrivateKey())
	proof := sharing.GenerateNIZKProof(testutils.TestCurve(), node1.LongtermKey.PrivateKey, node1.LongtermKey.PublicKey,
		ephemeralKeypairDealer.PublicKey, symmKey, testutils.TestCurve().NewGeneratorPoint())

	msg := ReceiveShareRecoveryMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ShareMessageType,
		CurveName:        testutils.TestCurveName(),
		SymmetricKey:     symmKey.ToAffineCompressed(),
		Proof:            proof,
		AcssData:         msgData,
	}
	return node1, receiver, msg
}

func getMsgDataAndStoreHash(dealer *testutils.PssTestNode, receiver *testutils.PssTestNode) (common.KeyPair, common.AcssData, common.ACSSRoundDetails) {
	ephemeralKeypairDealer, msgData, acssRoundDetails := getMsgData(dealer)

	hash, _ := common.HashAcssData(msgData)
	err := receiver.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = hash
	})
	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
	}
	return ephemeralKeypairDealer, msgData, acssRoundDetails
}

func getMsgData(dealer *testutils.PssTestNode) (common.KeyPair, common.AcssData, common.ACSSRoundDetails) {
	ephemeralKeypairDealer := common.GenerateKeyPair(testutils.TestCurve())
	n, k, _ := dealer.Params()

	secret := sharing.GenerateSecret(testutils.TestCurve())
	commitment, shares, err := sharing.GenerateCommitmentAndShares(
		secret,
		uint32(k),
		uint32(n),
		testutils.TestCurve(),
	)
	if err != nil {
		log.Errorf("Error generating commitments & shares: %v", err)
	}
	compressedCommitments := sharing.CompressCommitments(commitment)
	shareMap := make(map[string][]byte, n)
	for _, share := range shares {
		nodePublicKey := dealer.GetPublicKeyFor(int(share.Id), dealer.IsNewNode())

		cipherShare, hmacTag, err := sharing.EncryptSymmetricCalculateKey(
			share.Bytes(),
			nodePublicKey,
			ephemeralKeypairDealer.PrivateKey,
		)

		if err != nil {
			log.Errorf("Error while encrypting secret share, err=%v", err)
		}
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		pubkeyHex := common.PointToHex(nodePublicKey)
		cipherShare = sharing.Combine(cipherShare, hmacTag)
		shareMap[pubkeyHex] = cipherShare
	}

	msgData := common.AcssData{
		Commitments:           compressedCommitments,
		ShareMap:              shareMap,
		DealerEphemeralPubKey: common.PointToHex(ephemeralKeypairDealer.PublicKey),
	}

	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)
	return ephemeralKeypairDealer, msgData, acssRoundDetails
}
