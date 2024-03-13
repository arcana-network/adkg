package dacss

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

/*
Everything that can go "wrong" wrt what would be the happy path:
- sender not self: TestExecuteImplicateSenderNotSelf
- acssState not found: TestAcssStateNotFound
- acssState.AcssDataHash len 0: TestAcssDataUninitialized
- len(share) 0: TestShareLenZero
- acssState.AcssDataHash not equal to hash of received acssData: TestAcssDataHashMismatch
- already in share recovery: TestExecuteImplicateAlreadyInShareRecovery
- dealerPubkey received in msg data invalid for conversion: TestDealerPubkeyInvalid
- symmetric key invalid for conversion: TestSymmKeyInvalid
- SenderPubkeyHex (sender of initial Implicate msg) as passed on in the message invalid for conversion: TestSenderPubkeyInvalid
- proof unpacking fails: TestCantUnpackProof
- empty share: TestShareLenZero
- proof fails: TestZKPVerificationFails
- predicate passes: TestPredicatePasses
*/

/*
Function: Process

Testcase: happy path.
The share for which the implicate was sent, indeed fails the predicate.
The ZKP verifies.

Expectation: node sends ShareRecoveryMessage to self
*/
func TestExecuteHandlerHappyFlow(t *testing.T) {
	// 1. The dealer creates the shareMap
	// 2. Store the shareMap for this round in the Node 1's state
	// 3. Fill the share for node2 with random bytes, so it will fail the predicate
	node1, _, executeMsg, _ := happyPathSetup()

	executeMsg.Process(node1.Details(), node1)

	// Check that the node received a ShareRecoveryMsg from itself
	assert.True(t, len(node1.Transport.ReceivedMessages) == 1)
	assert.Equal(t, node1.Transport.ReceivedMessages[0].Type, ShareRecoveryMessageType)
}

/*
Function: Process

Testcase: sender of ImplicateExecute message is other node than self

Expectation: early return; no msg is sent
*/
func TestExecuteImplicateSenderNotSelf(t *testing.T) {
	node1, node2, executeMsg, _ := happyPathSetup()

	// The execute message is sent by node2 to node1
	// while is should be sent by node1 to node1
	executeMsg.Process(node2.Details(), node1)

	// That's why nothing happens
	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
}

/*
Function: Process

Testcase: in the receiving node, there is no AcssState. This is needed to verify the correctness of acssData in the msg

Expectation: early return; no msg is sent
*/
func TestAcssStateNotFound(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

	// The dealer creates the shareMap
	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())
	// Set the round parameters
	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)
	curve := curves.K256()
	acssData := getCorruptedAcssData(dealer, ephemeralKeypairDealer, node2)

	executeMsg := createProofAndExecuteMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails, acssData)
	executeMsg.Process(node1.Details(), node1)

	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
}

/*
Function: Process

Testcase: there is acssState in the node, but len(acssDataHash) is 0

Expectation: early return; no msg is sent
*/
func TestAcssDataUninitialized(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

	// The dealer creates the shareMap
	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())
	// Set the round parameters
	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)
	curve := curves.K256()
	acssData := getCorruptedAcssData(dealer, ephemeralKeypairDealer, node2)

	// This should trigger creating an empty state.
	node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		// don't add anything to the state
	})

	executeMsg := createProofAndExecuteMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails, acssData)
	executeMsg.Process(node1.Details(), node1)

	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
}

/*
Function: Process

Testcase: the stored acssData hash is not equal to the hash of the received acssData

Expectation: early return; no msg is sent
*/
func TestAcssDataHashMismatch(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

	// The dealer creates the shareMap
	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())
	// Set the round parameters
	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)
	curve := curves.K256()
	acssData := getCorruptedAcssData(dealer, ephemeralKeypairDealer, node2)

	// Store a different hash
	node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = []byte("different")
	})

	executeMsg := createProofAndExecuteMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails, acssData)
	executeMsg.Process(node1.Details(), node1)

	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
}

/*
Function: Process

Testcase: the node is already in share recovery (the next phase of implicate flow),
so there is no need to continue on.

Expectation: early return; no msg is sent
*/
func TestExecuteImplicateAlreadyInShareRecovery(t *testing.T) {
	node1, _, executeMsg, acssRoundDetails := happyPathSetup()

	node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.ShareRecoveryOngoing = true
	})
	executeMsg.Process(node1.Details(), node1)
	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
}

/*
Function: Process

Testcase: the dealer's ephemeral key is invalid for conversion to a point.
This is passed on in the ImplicateExecuteMsg.

Expectation: early return; no msg is sent
*/
func TestDealerPubkeyInvalid(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)
	curve := curves.K256()

	// Set dealer's ephemeral key to invalid value
	acssData := getCorruptedAcssData(dealer, ephemeralKeypairDealer, node2)
	acssData.DealerEphemeralPubKey = "invalid"

	// Store hash of acssData in the receiver's node state
	hash, _ := common.HashAcssData(acssData)
	err := node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = hash
	})
	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
	}

	executeMsg := createProofAndExecuteMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails, acssData)

	executeMsg.Process(node1.Details(), node1)

	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
}

/*
Function: Process

Testcase: the symmetric key is invalid for conversion to a point.
This is passed on in the ImplicateExecuteMsg.

Expectation: early return; no msg is sent
*/
func TestSymmKeyInvalid(t *testing.T) {
	node1, _, executeMsg, _ := happyPathSetup()

	executeMsg.SymmetricKey = []byte("invalid")

	executeMsg.Process(node1.Details(), node1)

	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
}

/*
Function: Process

Testcase: the pubkey of the sender of the implicate message is invalid for conversion to a point.
This is passed on in the ImplicateExecuteMsg.

Expectation: early return; no msg is sent
*/
func TestSenderPubkeyInvalid(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)
	curve := curves.K256()

	// Set dealer's ephemeral key to invalid value
	acssData := getCorruptedAcssData(dealer, ephemeralKeypairDealer, node2)
	acssData.ShareMap["invalid"] = []byte("invalid")

	// Store hash of acssData in the receiver's node state
	hash, _ := common.HashAcssData(acssData)
	err := node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = hash
	})
	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
	}
	executeMsg := createProofAndExecuteMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails, acssData)

	// Set dealer's ephemeral key to invalid value in the msg
	executeMsg.SenderPubkeyHex = "invalid"

	executeMsg.Process(node1.Details(), node1)

	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
}

/*
Function: Process

Testcase: proof is invalid and can't be unpacked.
Proof is passed in the ImplicateExecuteMsg.

Expectation: early return; no msg is sent
*/
func TestCantUnpackProof(t *testing.T) {
	node1, _, executeMsg, _ := happyPathSetup()

	bytes := make([]byte, 66)
	_, _ = rand.Read(bytes)
	executeMsg.Proof = bytes

	executeMsg.Process(node1.Details(), node1)

	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
}

/*
Function: Process

Testcase: the proof fails. It has a valid format, but doesn't proof what it should.
Proof is passed in the ImplicateExecuteMsg.

Expectation: early return; no msg is sent
*/
func TestZKPVerificationFails(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)
	curve := curves.K256()

	acssData := getCorruptedAcssData(dealer, ephemeralKeypairDealer, node2)

	// Store hash of acssData in the receiver's node state
	hash, _ := common.HashAcssData(acssData)
	err := node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = hash
	})
	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
	}

	symmKey, _ := sharing.CalculateSharedKey(ephemeralKeypairDealer.PublicKey, node2.PrivateKey())
	// Generation of proof is wrong; it should use node2.Longtermkey.PrivateKey but uses the key of node1
	proof := sharing.GenerateNIZKProof(curve, node1.LongtermKey.PrivateKey, node2.LongtermKey.PublicKey,
		ephemeralKeypairDealer.PublicKey, symmKey, curve.NewGeneratorPoint())

	executeMsg := createExecuteMsg(acssRoundDetails, symmKey, proof, node2.LongtermKey, acssData)
	executeMsg.Process(node1.Details(), node1)

	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
}

/*
Function: Process

Testcase: the share for which the implicate was sent, passes the predicate.
For the implicate flow to continue, the predicate should fail.

Expectation: early return; no msg is sent
*/
func TestPredicatePasses(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)
	curve := curves.K256()

	n, k, _ := dealer.Params()
	secret := sharing.GenerateSecret(curves.K256())
	commitment, shares, err := sharing.GenerateCommitmentAndShares(
		secret,
		uint32(k),
		uint32(n),
		curves.K256(),
	)
	if err != nil {
		log.Errorf("Error generating commitments & shares: %v\n", err)
	}
	compressedCommitments := sharing.CompressCommitments(commitment)
	shareMap := make(map[string][]byte, n)
	for _, share := range shares {
		nodePublicKey := dealer.GetPublicKeyFor(int(share.Id), dealer.IsNewNode())

		cipherShare, _ := sharing.EncryptSymmetricCalculateKey(
			share.Bytes(),
			nodePublicKey,
			ephemeralKeypairDealer.PrivateKey,
		)

		if err != nil {
			log.Errorf("Error while encrypting secret share, err=%v", err)
		}
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		pubkeyHex := hex.EncodeToString(nodePublicKey.ToAffineCompressed())

		shareMap[pubkeyHex] = cipherShare
	}
	msgData := common.AcssData{
		Commitments:           compressedCommitments,
		ShareMap:              shareMap,
		DealerEphemeralPubKey: hex.EncodeToString(ephemeralKeypairDealer.PublicKey.ToAffineCompressed()),
	}
	// Store hash of acssData in the receiver's node state
	hash, _ := common.HashAcssData(msgData)
	err = node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = hash
	})
	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
	}
	executeMsg := createProofAndExecuteMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails, msgData)
	executeMsg.Process(node1.Details(), node1)

	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
}

/*
FIXME
Function: Process

Testcase: the hash of share for which the implicate was sent, is empty.
This is stored in the node's state (acssState.acssDataHash)

Expectation: early return; no msg is sent
*/
func TestShareHashLenZero(t *testing.T) {
	node1, _, executeMsg, acssRoundDetails := happyPathSetup()

	// Set share to empty
	node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = []byte{}
	})

	executeMsg.Process(node1.Details(), node1)

	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
}

func happyPathSetup() (*testutils.PssTestNode, *testutils.PssTestNode, ImplicateExecuteMessage, common.ACSSRoundDetails) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

	acssRoundDetails := testutils.GetTestACSSRoundDetails(dealer)
	curve := curves.K256()

	acssData := getCorruptedAcssData(dealer, ephemeralKeypairDealer, node2)

	// Store hash of acssData in the receiver's node state
	hash, _ := common.HashAcssData(acssData)
	err := node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = hash
	})
	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
	}

	executeMsg := createProofAndExecuteMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails, acssData)
	return node1, node2, executeMsg, acssRoundDetails
}

func getCorruptedAcssData(dealer *testutils.PssTestNode, ephemeralKeypairDealer common.KeyPair,
	nodeOfCorruptShare *testutils.PssTestNode) common.AcssData {
	n, k, _ := dealer.Params()
	secret := sharing.GenerateSecret(curves.K256())
	commitment, shares, err := sharing.GenerateCommitmentAndShares(
		secret,
		uint32(k),
		uint32(n),
		curves.K256(),
	)
	if err != nil {
		log.Errorf("Error generating commitments & shares: %v\n", err)
	}
	compressedCommitments := sharing.CompressCommitments(commitment)
	shareMap := make(map[string][]byte, n)
	for _, share := range shares {
		nodePublicKey := dealer.GetPublicKeyFor(int(share.Id), dealer.IsNewNode())

		cipherShare := []byte{}
		if share.Id == uint32(nodeOfCorruptShare.Details().Index) {
			bytes := make([]byte, 33)
			_, _ = rand.Read(bytes)
			cipherShare = bytes
		} else {
			cipherShare, _ = sharing.EncryptSymmetricCalculateKey(
				share.Bytes(),
				nodePublicKey,
				ephemeralKeypairDealer.PrivateKey,
			)
		}

		if err != nil {
			log.Errorf("Error while encrypting secret share, err=%v", err)
		}
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		pubkeyHex := hex.EncodeToString(nodePublicKey.ToAffineCompressed())

		shareMap[pubkeyHex] = cipherShare
	}
	msgData := common.AcssData{
		Commitments:           compressedCommitments,
		ShareMap:              shareMap,
		DealerEphemeralPubKey: hex.EncodeToString(ephemeralKeypairDealer.PublicKey.ToAffineCompressed()),
	}
	return msgData
}

func createProofAndExecuteMsg(ephemeralKeypairDealer common.KeyPair, initiatorImplicate *testutils.PssTestNode,
	curve *curves.Curve, acssRoundDetails common.ACSSRoundDetails, acssData common.AcssData) ImplicateExecuteMessage {
	symmKey, _ := sharing.CalculateSharedKey(ephemeralKeypairDealer.PublicKey, initiatorImplicate.PrivateKey())
	proof := sharing.GenerateNIZKProof(curve, initiatorImplicate.LongtermKey.PrivateKey, initiatorImplicate.LongtermKey.PublicKey,
		ephemeralKeypairDealer.PublicKey, symmKey, curve.NewGeneratorPoint())

	implicateExecuteMsg := createExecuteMsg(acssRoundDetails, symmKey, proof, initiatorImplicate.LongtermKey, acssData)
	return implicateExecuteMsg
}

func createExecuteMsg(acssRoundDetails common.ACSSRoundDetails, symmKey curves.Point,
	proof []byte, initiatorImplicatePubkey common.KeyPair, acssData common.AcssData) ImplicateExecuteMessage {
	implicateExecuteMsg := ImplicateExecuteMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ImplicateExecuteMessageType,
		CurveName:        common.SECP256K1,
		SymmetricKey:     symmKey.ToAffineCompressed(),
		Proof:            proof,
		SenderPubkeyHex:  hex.EncodeToString(initiatorImplicatePubkey.PublicKey.ToAffineCompressed()),
		AcssData:         acssData,
	}
	return implicateExecuteMsg
}
