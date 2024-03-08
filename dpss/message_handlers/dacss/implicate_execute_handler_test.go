package dacss

// import (
// 	"crypto/rand"
// 	"encoding/hex"
// 	"testing"

// 	"github.com/arcana-network/dkgnode/common"
// 	"github.com/arcana-network/dkgnode/common/sharing"
// 	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
// 	"github.com/coinbase/kryptology/pkg/core/curves"
// 	log "github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// )

// /*
// Everything that can go "wrong" wrt what would be the happy path:
// - sender not self: TestExecuteImplicateSenderNotSelf
// - acssState not found: TestAcssStateNotFound
// - acssState.acssData uninitialized: TestAcssDataUninitialized
// - already in share recovery: TestExecuteImplicateAlreadyInShareRecovery
// - dealerPubkey as stored in Node's state invalid for conversion: TestDealerPubkeyInvalid
// - symmetric key invalid for conversion: TestSymmKeyInvalid
// - SenderPubkeyHex (sender of initial Implicate msg) as passed on in the message invalid for conversion: TestSenderPubkeyInvalid
// - proof unpacking fails: TestCantUnpackProof
// - empty share: TestShareLenZero
// - proof fails: TestZKPVerificationFails
// - predicate passes: TestPredicatePasses
// */

// /*
// Function: Process

// Testcase: happy path.
// The share for which the implicate was sent, indeed fails the predicate.
// The ZKP verifies.

// Expectation: node sends ShareRecoveryMessage to self
// */
// func TestExecuteHandlerHappyFlow(t *testing.T) {
// 	// 1. The dealer creates the shareMap
// 	// 2. Store the shareMap for this round in the Node 1's state
// 	// 3. Fill the share for node2 with random bytes, so it will fail the predicate
// 	node1, _, executeMsg, _ := happyPathSetup()

// 	executeMsg.Process(node1.Details(), node1)

// 	// Check that the node received a ShareRecoveryMsg from itself
// 	assert.True(t, len(node1.Transport.ReceivedMessages) == 1)
// 	assert.Equal(t, node1.Transport.ReceivedMessages[0].Type, ShareRecoveryMessageType)
// }

// /*
// Function: Process

// Testcase: sender of ImplicateExecute message is other node than self

// Expectation: early return; no msg is sent
// */
// func TestExecuteImplicateSenderNotSelf(t *testing.T) {
// 	node1, node2, executeMsg, _ := happyPathSetup()

// 	// The execute message is sent by node2 to node1
// 	// while is should be sent by node1 to node1
// 	executeMsg.Process(node2.Details(), node1)

// 	// That's why nothing happens
// 	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
// }

// /*
// Function: Process

// Testcase: in the receiving node, there is no AcssState. This is needed to verify the predicate

// Expectation: early return; no msg is sent
// */
// func TestAcssStateNotFound(t *testing.T) {
// 	defaultSetup := testutils.DefaultTestSetup()
// 	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

// 	// The dealer creates the shareMap
// 	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())
// 	// Set the round parameters
// 	acssRoundDetails := getTestACSSRoundDetails(dealer)
// 	curve := curves.K256()

// 	executeMsg := createProofAndExecuteMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails)
// 	executeMsg.Process(node1.Details(), node1)

// 	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
// }

// /*
// Function: Process

// Testcase: there is acssState in the node, but the acssData is empty.

// Expectation: early return; no msg is sent
// */
// func TestAcssDataUninitialized(t *testing.T) {
// 	defaultSetup := testutils.DefaultTestSetup()
// 	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

// 	// The dealer creates the shareMap
// 	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())
// 	// Set the round parameters
// 	acssRoundDetails := getTestACSSRoundDetails(dealer)
// 	curve := curves.K256()

// 	// This should trigger creating an empty state.
// 	node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
// 		// don't add anything to the state
// 	})

// 	executeMsg := createProofAndExecuteMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails)
// 	executeMsg.Process(node1.Details(), node1)

// 	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
// }

// /*
// Function: Process

// Testcase: the node is already in share recovery (the next phase of implicate flow),
// so there is no need to continue on.

// Expectation: early return; no msg is sent
// */
// func TestExecuteImplicateAlreadyInShareRecovery(t *testing.T) {
// 	node1, _, executeMsg, acssRoundDetails := happyPathSetup()

// 	node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
// 		state.ShareRecoveryOngoing = true
// 	})
// 	executeMsg.Process(node1.Details(), node1)
// 	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
// }

// /* FIXME
// Function: Process

// Testcase: the dealer's ephemeral key is invalid for conversion to a point.
// This is stored in the node's state (acssData).

// Expectation: early return; no msg is sent
// */
// // func TestDealerPubkeyInvalid(t *testing.T) {
// // 	node1, _, executeMsg, acssRoundDetails := happyPathSetup()

// // 	// Set dealer's ephemeral key to invalid value
// // 	node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
// // 		state.AcssData.DealerEphemeralPubKey = "invalid"
// // 	})

// // 	executeMsg.Process(node1.Details(), node1)

// // 	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
// // }

// /*
// Function: Process

// Testcase: the symmetric key is invalid for conversion to a point.
// This is passed on in the ImplicateExecuteMsg.

// Expectation: early return; no msg is sent
// */
// func TestSymmKeyInvalid(t *testing.T) {
// 	node1, _, executeMsg, _ := happyPathSetup()

// 	executeMsg.SymmetricKey = []byte("invalid")

// 	executeMsg.Process(node1.Details(), node1)

// 	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
// }

// /*FIXME
// Function: Process

// Testcase: the pubkey of the sender of the implicate message is invalid for conversion to a point.
// This is passed on in the ImplicateExecuteMsg.

// Expectation: early return; no msg is sent
// */
// // func TestSenderPubkeyInvalid(t *testing.T) {
// // 	invalidPubkeyhex := "invalid"

// // 	// Create an otherwise valid shareMap, with exception of the share with id = 2
// // 	// the share will be stored un an invalid pubkey
// // 	// In the ImplicateExecuteMsg the invalid pubkey is sent along.
// // 	// this invalid pubkey can't be transformed to a point

// // 	defaultSetup := testutils.DefaultTestSetup()
// // 	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

// // 	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

// // 	acssRoundDetails := getTestACSSRoundDetails(dealer)
// // 	curve := curves.K256()

// // 	n, k, _ := dealer.Params()
// // 	secret := sharing.GenerateSecret(curves.K256())
// // 	commitment, shares, err := sharing.GenerateCommitmentAndShares(
// // 		secret,
// // 		uint32(k),
// // 		uint32(n),
// // 		curves.K256(),
// // 	)
// // 	if err != nil {
// // 		log.Errorf("Error generating commitments & shares: %v\n", err)
// // 	}
// // 	compressedCommitments := sharing.CompressCommitments(commitment)
// // 	shareMap := make(map[string][]byte, n)
// // 	for _, share := range shares {
// // 		nodePublicKey := dealer.GetPublicKeyFor(int(share.Id), !dealer.IsOldNode())

// // 		cipherShare, err := sharing.EncryptSymmetricCalculateKey(
// // 			share.Bytes(),
// // 			nodePublicKey,
// // 			ephemeralKeypairDealer.PrivateKey,
// // 		)

// // 		if err != nil {
// // 			log.Errorf("Error while encrypting secret share, err=%v", err)
// // 		}
// // 		log.Debugf("CIPHER_SHARE=%v", cipherShare)
// // 		pubkeyHex := hex.EncodeToString(nodePublicKey.ToAffineCompressed())

// // 		if share.Id == 2 {
// // 			bytes := make([]byte, 33)
// // 			_, _ = rand.Read(bytes)
// // 			shareMap[invalidPubkeyhex] = bytes
// // 		} else {
// // 			shareMap[pubkeyHex] = cipherShare
// // 		}
// // 	}
// // 	msgData := common.AcssData{
// // 		Commitments:           compressedCommitments,
// // 		ShareMap:              shareMap,
// // 		DealerEphemeralPubKey: hex.EncodeToString(ephemeralKeypairDealer.PublicKey.ToAffineCompressed()),
// // 	}

// // 	err = node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
// // 		state.AcssData = msgData
// // 	})
// // 	if err != nil {
// // 		log.Errorf("Error updating AcssData in state: %v", err)
// // 	}

// // 	executeMsg := createProofAndExecuteMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails)
// // 	executeMsg.SenderPubkeyHex = invalidPubkeyhex

// // 	executeMsg.Process(node1.Details(), node1)
// // 	// No messages should be sent, because the pubkey of the initiator
// // 	// of the implicate flow can't be tranformed to a point
// // 	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
// // }

// /*
// Function: Process

// Testcase: proof is invalid and can't be unpacked.
// Proof is passed in the ImplicateExecuteMsg.

// Expectation: early return; no msg is sent
// */
// func TestCantUnpackProof(t *testing.T) {
// 	node1, _, executeMsg, _ := happyPathSetup()

// 	bytes := make([]byte, 66)
// 	_, _ = rand.Read(bytes)
// 	executeMsg.Proof = bytes

// 	executeMsg.Process(node1.Details(), node1)

// 	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
// }

// /*
// Function: Process

// Testcase: the proof fails. It has a valid format, but doesn't proof what it should.
// Proof is passed in the ImplicateExecuteMsg.

// Expectation: early return; no msg is sent
// */
// func TestZKPVerificationFails(t *testing.T) {
// 	defaultSetup := testutils.DefaultTestSetup()
// 	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

// 	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

// 	acssRoundDetails := getTestACSSRoundDetails(dealer)
// 	curve := curves.K256()

// 	storeShareMapForNode(dealer, ephemeralKeypairDealer, node1, acssRoundDetails)

// 	invalidateShareForNode(node2, node1, acssRoundDetails)

// 	symmKey, _ := sharing.CalculateSharedKey(ephemeralKeypairDealer.PublicKey, node2.PrivateKey())
// 	// Generation of proof is wrong; it should use node2.Longtermkey.PrivateKey but uses the key of node1
// 	proof := sharing.GenerateNIZKProof(curve, node1.LongtermKey.PrivateKey, node2.LongtermKey.PublicKey,
// 		ephemeralKeypairDealer.PublicKey, symmKey, curve.NewGeneratorPoint())

// 	executeMsg := createExecuteMsg(acssRoundDetails, symmKey, proof, node2.LongtermKey)
// 	executeMsg.Process(node1.Details(), node1)

// 	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
// }

// /*
// Function: Process

// Testcase: the share for which the implicate was sent, passes the predicate.
// For the implicate flow to continue, the predicate should fail.

// Expectation: early return; no msg is sent
// */
// func TestPredicatePasses(t *testing.T) {
// 	defaultSetup := testutils.DefaultTestSetup()
// 	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

// 	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

// 	acssRoundDetails := getTestACSSRoundDetails(dealer)
// 	curve := curves.K256()

// 	storeShareMapForNode(dealer, ephemeralKeypairDealer, node1, acssRoundDetails)

// 	executeMsg := createProofAndExecuteMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails)
// 	executeMsg.Process(node1.Details(), node1)

// 	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
// }

// /*
// FIXME
// Function: Process

// Testcase: the hash of share for which the implicate was sent, is empty.
// This is stored in the node's state (acssState.acssDataHash)

// Expectation: early return; no msg is sent
// */
// func TestShareHashLenZero(t *testing.T) {
// 	node1, _, executeMsg, acssRoundDetails := happyPathSetup()

// 	// Set share to empty
// 	node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
// 		state.AcssDataHash = []byte{}
// 	})

// 	executeMsg.Process(node1.Details(), node1)

// 	assert.True(t, len(node1.Transport.ReceivedMessages) == 0)
// }

// func happyPathSetup() (*testutils.PssTestNode, *testutils.PssTestNode, ImplicateExecuteMessage, common.ACSSRoundDetails) {
// 	defaultSetup := testutils.DefaultTestSetup()
// 	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

// 	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

// 	acssRoundDetails := getTestACSSRoundDetails(dealer)
// 	curve := curves.K256()

// 	storeShareMapForNode(dealer, ephemeralKeypairDealer, node1, acssRoundDetails)

// 	invalidateShareForNode(node2, node1, acssRoundDetails)

// 	executeMsg := createProofAndExecuteMsg(ephemeralKeypairDealer, node2, curve, acssRoundDetails)
// 	return node1, node2, executeMsg, acssRoundDetails
// }

// func storeShareMapForNode(dealer *testutils.PssTestNode, ephemeralKeypairDealer common.KeyPair, node1 *testutils.PssTestNode, acssRoundDetails common.ACSSRoundDetails) {
// 	n, k, _ := dealer.Params()
// 	secret := sharing.GenerateSecret(curves.K256())
// 	commitment, shares, err := sharing.GenerateCommitmentAndShares(
// 		secret,
// 		uint32(k),
// 		uint32(n),
// 		curves.K256(),
// 	)
// 	if err != nil {
// 		log.Errorf("Error generating commitments & shares: %v\n", err)
// 	}
// 	compressedCommitments := sharing.CompressCommitments(commitment)
// 	shareMap := make(map[string][]byte, n)
// 	for _, share := range shares {
// 		nodePublicKey := dealer.GetPublicKeyFor(int(share.Id), !dealer.IsOldNode())

// 		cipherShare, err := sharing.EncryptSymmetricCalculateKey(
// 			share.Bytes(),
// 			nodePublicKey,
// 			ephemeralKeypairDealer.PrivateKey,
// 		)

// 		if err != nil {
// 			log.Errorf("Error while encrypting secret share, err=%v", err)
// 		}
// 		log.Debugf("CIPHER_SHARE=%v", cipherShare)
// 		pubkeyHex := hex.EncodeToString(nodePublicKey.ToAffineCompressed())

// 		shareMap[pubkeyHex] = cipherShare
// 	}
// 	msgData := common.AcssData{
// 		Commitments:           compressedCommitments,
// 		ShareMap:              shareMap,
// 		DealerEphemeralPubKey: hex.EncodeToString(ephemeralKeypairDealer.PublicKey.ToAffineCompressed()),
// 	}

// 	hash, err := HashAcssData(msgData)
// 	err = node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
// 		state.AcssDataHash = hash
// 	})
// 	if err != nil {
// 		log.Errorf("Error updating AcssData in state: %v", err)
// 	}
// }

// func invalidateShareForNode(node2 *testutils.PssTestNode, node1 *testutils.PssTestNode, acssRoundDetails common.ACSSRoundDetails) {
// 	bytes := make([]byte, 33)
// 	_, _ = rand.Read(bytes)
// 	node2PubkeyHex := common.PointToHex(node2.LongtermKey.PublicKey)
// 	hash, err := HashAcssData(msgData)

// 	node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
// 		state.AcssDataHash = hash
// 	})
// }

// func createProofAndExecuteMsg(ephemeralKeypairDealer common.KeyPair, initiatorImplicate *testutils.PssTestNode, curve *curves.Curve, acssRoundDetails common.ACSSRoundDetails) ImplicateExecuteMessage {
// 	symmKey, _ := sharing.CalculateSharedKey(ephemeralKeypairDealer.PublicKey, initiatorImplicate.PrivateKey())
// 	proof := sharing.GenerateNIZKProof(curve, initiatorImplicate.LongtermKey.PrivateKey, initiatorImplicate.LongtermKey.PublicKey,
// 		ephemeralKeypairDealer.PublicKey, symmKey, curve.NewGeneratorPoint())

// 	implicateExecuteMsg := createExecuteMsg(acssRoundDetails, symmKey, proof, initiatorImplicate.LongtermKey)
// 	return implicateExecuteMsg
// }

// func createExecuteMsg(acssRoundDetails common.ACSSRoundDetails, symmKey curves.Point, proof []byte, initiatorImplicatePubkey common.KeyPair) ImplicateExecuteMessage {
// 	implicateExecuteMsg := ImplicateExecuteMessage{
// 		ACSSRoundDetails: acssRoundDetails,
// 		Kind:             ImplicateExecuteMessageType,
// 		CurveName:        common.SECP256K1,
// 		SymmetricKey:     symmKey.ToAffineCompressed(),
// 		Proof:            proof,
// 		SenderPubkeyHex:  hex.EncodeToString(initiatorImplicatePubkey.PublicKey.ToAffineCompressed()),
// 	}
// 	return implicateExecuteMsg
// }
