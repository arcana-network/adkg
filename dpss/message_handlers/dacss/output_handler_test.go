package dacss

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	mrand "math/rand"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	ksharing "github.com/coinbase/kryptology/pkg/sharing"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/torusresearch/bijson"
)

/*
Function: Process

Testcase: the RBC has already ended.

Expectations:
- The node do an early return.
- There is no broadcast of the commitment.
*/
func TestRBCAlreadyEnded(test *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealerNode, testNode := defaultSetup.GetTwoOldNodesFromTestSetup()
	ephemeralKeypairDealer := common.GenerateKeyPair(testutils.TestCurve())

	correctMessage, err := generateCorrectOutputMessage(
		testNode,
		dealerNode,
		ephemeralKeypairDealer,
	)
	assert.Nil(test, err)

	testNode.State().AcssStore.UpdateAccsState(
		correctMessage.AcssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.RBCState.Phase = common.Ended
		},
	)

	correctMessage.Process(testNode.Details(), testNode)

	// We wait here. We can't use the message signal channels because no message
	// is sent here.
	time.Sleep(1 * time.Second)

	broadcastedMsgs := testNode.Transport().BroadcastedMessages
	assert.Zero(test, len(broadcastedMsgs))
}

/*
Function: Process

Testcase: the predicate was not verified correctly.

Expectations:
- If the RBC is not ended, it should remain not ended.
- There is no broadcast of the commitment.
- The commitment sent should be false.
*/
func TestNotVerifiedCorrectly(test *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealerNode, testNode := defaultSetup.GetTwoOldNodesFromTestSetup()
	ephemeralKeypairDealer := common.GenerateKeyPair(testutils.TestCurve())

	correctMessage, err := generateCorrectOutputMessage(
		testNode,
		dealerNode,
		ephemeralKeypairDealer,
	)
	assert.Nil(test, err)

	_, k, _ := testNode.Params()
	fakeMessage, err := constructFakeMessage(
		correctMessage,
		k,
	)
	assert.Nil(test, err)

	fakeMessage.Process(testNode.Details(), testNode)

	// We wait here. We can't use the signal and channels mechanism because no
	// message is sent in this case.
	time.Sleep(2 * time.Second)

	stateNode, found, err := testNode.State().AcssStore.Get(
		correctMessage.AcssRoundDetails.ToACSSRoundID(),
	)
	assert.Nil(test, err)
	assert.True(test, found)

	assert.NotEqual(test, stateNode.RBCState.Phase, common.Ended)

	broadcastedMsgs := testNode.Transport().BroadcastedMessages
	assert.Zero(test, len(broadcastedMsgs))

	assert.False(test, stateNode.CommitmentSent)
}

/*
Function: Process

Testcase: the commitment message was already sent.

Expectations:
- There is no broadcast of the commitment message again.
*/
func TestCommitmentMessageAlreadySent(test *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealerNode, testNode := defaultSetup.GetTwoOldNodesFromTestSetup()
	ephemeralKeypairDealer := common.GenerateKeyPair(testutils.TestCurve())

	correctMessage, err := generateCorrectOutputMessage(
		testNode,
		dealerNode,
		ephemeralKeypairDealer,
	)
	assert.Nil(test, err)

	testNode.State().AcssStore.UpdateAccsState(
		correctMessage.AcssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.CommitmentSent = true
		},
	)

	correctMessage.Process(testNode.Details(), testNode)

	// We wait here. We can't use the signal and channels mechanism because no
	// message is sent in this case.
	time.Sleep(2 * time.Second)

	broadcastedMsgs := testNode.Transport().BroadcastedMessages
	assert.Zero(test, len(broadcastedMsgs))
}

/*
Function: Process

Testcase: the output message is sent twice.

Expectations:
- There is just one broadcasted message.
- One of the messages do an early return.
- RBC should end.
- Commitment sent should be set to true.
*/
func TestCommitmentMessageSentTwice(test *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealerNode, testNode := defaultSetup.GetTwoOldNodesFromTestSetup()
	ephemeralKeypairDealer := common.GenerateKeyPair(testutils.TestCurve())

	correctMessage, err := generateCorrectOutputMessage(
		testNode,
		dealerNode,
		ephemeralKeypairDealer,
	)
	assert.Nil(test, err)

	// Send the Output message twice
	correctMessage.Process(testNode.Details(), testNode)
	correctMessage.Process(testNode.Details(), testNode)

	// We wait here. We can't use the signal and channels mechanism because no
	// message is sent in this case.
	time.Sleep(2 * time.Second)

	broadcastedMsgs := testNode.Transport().BroadcastedMessages
	stateNode, found, err := testNode.State().AcssStore.Get(
		correctMessage.AcssRoundDetails.ToACSSRoundID(),
	)
	assert.Nil(test, err)
	assert.True(test, found)

	assert.Equal(test, stateNode.RBCState.Phase, common.Ended)
	assert.Equal(test, 1, len(broadcastedMsgs))
	assert.True(test, stateNode.CommitmentSent)
}

/*
Function: Process

Testcase: happy path.

Expectations:
- The RBC state has ended.
- There is one broadcast of the commitment message.
- The commitment sent is set to true.
*/
func TestOutputHappyPath(test *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealerNode, testNode := defaultSetup.GetTwoOldNodesFromTestSetup()
	ephemeralKeypairDealer := common.GenerateKeyPair(testutils.TestCurve())

	correctMessage, err := generateCorrectOutputMessage(
		testNode,
		dealerNode,
		ephemeralKeypairDealer,
	)
	assert.Nil(test, err)

	correctMessage.Process(testNode.Details(), testNode)

	dealerNode.Transport().WaitForBroadcastSent(1)

	broadcastedMsgs := testNode.Transport().BroadcastedMessages
	stateNode, found, err := testNode.State().AcssStore.Get(
		correctMessage.AcssRoundDetails.ToACSSRoundID(),
	)
	assert.Nil(test, err)
	assert.True(test, found)

	assert.Equal(test, stateNode.RBCState.Phase, common.Ended)
	assert.Equal(test, 1, len(broadcastedMsgs))
	assert.True(test, stateNode.CommitmentSent)
}

func generateCorrectOutputMessage(
	testNode *testutils.PssTestNode,
	dealerNode *testutils.PssTestNode,
	dealerEphemeralKeys common.KeyPair,
) (DacssOutputMessage, error) {
	// Creates the details for the ACSS protocol.
	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: dealerNode.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: pssRoundDetails,
		ACSSCount:       1,
	}

	n, k, _ := dealerNode.Params()

	// Create a commitment for a secret.
	secret := sharing.GenerateSecret(testutils.TestCurve())
	commitments, shares, err := sharing.GenerateCommitmentAndShares(
		secret,
		uint32(k),
		uint32(n),
		testutils.TestCurve(),
	)
	if err != nil {
		return DacssOutputMessage{}, err
	}

	// Init share map, key: pubkey receiver node, value: share
	shareMap := make(map[string][]byte, n)

	// Encrypt each share with node respective generated symmetric key using Ephemeral Private key and add to share map
	// ADD paper Section 5.3 of https://eprint.iacr.org/2021/777.pdf (making AVSS algorithm ACSS)
	for _, share := range shares {
		nodePublicKey := testNode.GetPublicKeyFor(int(share.Id), testNode.IsNewNode())
		if nodePublicKey == nil {
			return DacssOutputMessage{}, errors.New("Public key not found")
		}

		// Encryption is done with symmetric key Ki = PKi ^ SKd (pubkey of receiver, secret key of sender)
		cipherShare, hmacTag, err := sharing.EncryptSymmetricCalculateKey(share.Bytes(), nodePublicKey, dealerEphemeralKeys.PrivateKey)
		if err != nil {
			log.Errorf("Error while encrypting secret share, err=%v", err)
			return DacssOutputMessage{}, err
		}
		pubkeyHex := common.PointToHex(nodePublicKey)
		cipherShare = sharing.Combine(cipherShare, hmacTag)
		shareMap[pubkeyHex] = cipherShare
	}

	testNode.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.RBCState.Phase = common.Proposing
		},
	)

	compressedCommitments := sharing.CompressCommitments(commitments)

	dealerEphemeralPubKeyBytes := hex.EncodeToString(dealerEphemeralKeys.PublicKey.ToAffineCompressed())
	msgData := common.AcssData{
		Commitments:           compressedCommitments,
		ShareMap:              shareMap,
		DealerEphemeralPubKey: dealerEphemeralPubKeyBytes,
	}

	bytesMsgData, err := bijson.Marshal(msgData)
	if err != nil {
		return DacssOutputMessage{}, err
	}

	outputMsg := DacssOutputMessage{
		AcssRoundDetails: acssRoundDetails,
		kind:             DacssOutputMessageType,
		curveName:        common.CurveName(testutils.TestCurve().Name),
		Data:             bytesMsgData,
	}
	return outputMsg, nil
}

// This function takes a correct OUTPUT message and fakes its information
// returning an OUTPUT message with fake information.
//
// The purpose is to create a message whose commitments and shares do not meet.
// One can change the shares for that, but instead, we choose to change the
// commitments which will have a similar effect.
func constructFakeMessage(correctMsg DacssOutputMessage, numParties int) (
	DacssOutputMessage,
	error,
) {
	var msgData common.AcssData
	bijson.Unmarshal(correctMsg.Data, &msgData)

	uncompressedCommitmnets, err := sharing.DecompressCommitments(
		numParties,
		msgData.Commitments,
		common.CurveFromName(correctMsg.curveName),
	)
	if err != nil {
		return DacssOutputMessage{}, err
	}

	randomIdx := mrand.Intn(len(uncompressedCommitmnets))
	randomPoint := testutils.TestCurve().Point.Random(rand.Reader)
	for randomPoint.Equal(uncompressedCommitmnets[randomIdx]) {
		randomPoint = testutils.TestCurve().Point.Random(rand.Reader)
	}
	uncompressedCommitmnets[randomIdx] = randomPoint
	fakeCommitments := ksharing.FeldmanVerifier{
		Commitments: uncompressedCommitmnets,
	}
	fakeCompressedComm := sharing.CompressCommitments(
		&fakeCommitments,
	)

	fakeMsgData := common.AcssData{
		Commitments:           fakeCompressedComm,
		ShareMap:              msgData.ShareMap,
		DealerEphemeralPubKey: msgData.DealerEphemeralPubKey,
	}

	bytesFakeMsgData, err := bijson.Marshal(fakeMsgData)
	if err != nil {
		return DacssOutputMessage{}, err
	}

	outputFakeMsg := DacssOutputMessage{
		AcssRoundDetails: correctMsg.AcssRoundDetails,
		kind:             DacssOutputMessageType,
		curveName:        common.CurveName(testutils.TestCurve().Name),
		Data:             bytesFakeMsgData,
	}
	return outputFakeMsg, nil
}
