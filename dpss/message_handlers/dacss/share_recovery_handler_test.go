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
Happy path: current node has a valid share, so it sends a ReceiveShareRecoveryMessage to the other nodes.
Also, shareRecoveryOngoing set to true

Other paths:
- sender not self: TestSenderNotSelfShareRecovery
- acssState not found: TestNoAcssDataInState
- acssDataHash length 0: TestAcssDataHashLenZero
- shareRecoveryOngoing already true: TestShareRecoveryAlreadyTrue
- ValidShareOutput false: TestValidShareOutputFalse
*/

/*
Function: Process

Testcase: "Happy path". The current node has already output a valid share,
so it sends a ReceiveShareRecoveryMessage to the other nodes.

Expectations:
1. ReceiveShareRecoveryMessage was sent to all other nodes from committee
2. shareRecoveryOngoing set to true
*/
func TestSendReceiveShareRecoveryMsg(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	node1, _, acssRoundDetails, shareRecoveryMessage := shareRecoveryHappyPathSetup()

	_, err := node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		//To mimic the condition of A node having a valid share
		state.ReceivedShare = &sharing.ShamirShare{
			Id:    0,
			Value: []byte{0, 1, 2, 3},
		}
	})
	assert.Nil(t, err)
	shareRecoveryMessage.Process(node1.Details(), node1)

	nOld, _, _ := node1.Params()
	node1.Transport().WaitForMessagesSent(nOld)

	// Check 1: ReceiveShareRecoveryMessage was sent to all other nodes from committee
	sentMsgs := node1.Transport().GetSentMessages()
	assert.Equal(t, nOld, len(sentMsgs))

	// Check 2: shareRecoveryOngoing set to true
	acssState, _, _ := node1.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
	assert.True(t, acssState.ShareRecoveryOngoing)

	// TODO should we check what's in the message?
}

/*
Function: Process

Testcase: Sender not self

Expectations:
1. No message was sent
2. shareRecoveryOngoing not set to true
*/
func TestSenderNotSelfShareRecovery(t *testing.T) {
	node1, node2, acssRoundDetails, shareRecoveryMessage := shareRecoveryHappyPathSetup()

	shareRecoveryMessage.Process(node2.Details(), node1)

	// Given that no message is sent, we cannot use the signal/channel strategy.
	// Therefore we need to wait.
	time.Sleep(100 * time.Millisecond)

	// Check 1: No message was sent
	sentMsgs := node1.Transport().GetSentMessages()
	assert.Equal(t, 0, len(sentMsgs))

	// Check 2: shareRecoveryOngoing not set to true
	acssState, _, _ := node1.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
	assert.False(t, acssState.ShareRecoveryOngoing)
}

/*
Function: Process

Testcase: Node has no acssState

Expectations:
1. No message was sent
2. acssState still not found afterwards
*/
func TestNoAcssDataInState(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, _ := defaultSetup.GetThreeOldNodesFromTestSetup()

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

	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
	}

	shareRecoveryMessage := ShareRecoveryMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ShareMessageType,
		CurveName:        testutils.TestCurveName(),
		AcssData:         msgData,
	}

	shareRecoveryMessage.Process(node1.Details(), node1)

	// Given that no message is sent, we cannot use the signal/channel strategy.
	// Therefore we need to wait.
	time.Sleep(100 * time.Millisecond)

	// Check 1: No message was sent
	sentMsgs := node1.Transport().GetSentMessages()
	assert.Equal(t, 0, len(sentMsgs))

	// Check 2: acssState still not existing
	_, found, _ := node1.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
	assert.False(t, found)
}

/*
Function: Process

Testcase: acssDataHash has length zero

Expectations:
1. shareRecoveryOngoing is set to true
2. No message was sent
*/
func TestAcssDataHashLenZero(t *testing.T) {
	node1, _, acssRoundDetails, shareRecoveryMessage := shareRecoveryHappyPathSetup()

	node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = []byte{}
	})
	shareRecoveryMessage.Process(node1.Details(), node1)

	// Given that no message is sent, we cannot use the signal/channel strategy.
	// Therefore we need to wait.
	time.Sleep(100 * time.Millisecond)

	// Check 1: shareRecoveryOngoing is set to true
	acssState, _, _ := node1.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
	assert.True(t, acssState.ShareRecoveryOngoing)

	// Check 2: No message was sent
	sentMsgs := node1.Transport().GetSentMessages()
	assert.Equal(t, 0, len(sentMsgs))
}

/*
Function: Process

Testcase: shareRecoveryOngoing already set to true

Expectations:
1. No message was sent
*/
func TestShareRecoveryAlreadyTrue(t *testing.T) {
	node1, _, acssRoundDetails, shareRecoveryMessage := shareRecoveryHappyPathSetup()

	node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.ShareRecoveryOngoing = true
	})
	shareRecoveryMessage.Process(node1.Details(), node1)

	// Given that no message is sent, we cannot use the signal/channel strategy.
	// Therefore we need to wait.
	time.Sleep(100 * time.Millisecond)

	// Check No message was sent
	sentMsgs := node1.Transport().GetSentMessages()
	assert.Equal(t, 0, len(sentMsgs))
}

/*
Function: Process

Testcase: current Node doesn't have a valid share from finished RBC

Expectations:
1. shareRecoveryOngoing is set to true
2. No message was sent
*/
func TestValidShareOutputFalse(t *testing.T) {
	node1, _, acssRoundDetails, shareRecoveryMessage := shareRecoveryHappyPathSetup()

	node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.ValidShareOutput = false
	})
	shareRecoveryMessage.Process(node1.Details(), node1)

	// Given that no message is sent, we cannot use the signal/channel strategy.
	// Therefore we need to wait.
	time.Sleep(5 * time.Second)

	// Check 1: shareRecoveryOngoing is set to true
	acssState, _, _ := node1.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
	assert.True(t, acssState.ShareRecoveryOngoing)

	// Check 2: No message was sent
	sentMsgs := node1.Transport().GetSentMessages()
	assert.Equal(t, 0, len(sentMsgs))
}

func shareRecoveryHappyPathSetup() (*testutils.PssTestNode, *testutils.PssTestNode, common.ACSSRoundDetails, ShareRecoveryMessage) {
	defaultSetup := testutils.DefaultTestSetup()
	dealer, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()

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

	// The current node already completed the full RBC process and ended up with a valid share (ValidShareOutput = true)
	hash, _ := common.HashAcssData(msgData)
	_, err = node1.State().AcssStore.UpdateAccsState(acssRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = hash
		state.ValidShareOutput = true
	})
	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
	}

	shareRecoveryMessage := ShareRecoveryMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ShareMessageType,
		CurveName:        testutils.TestCurveName(),
		AcssData:         msgData,
	}
	return node1, node2, acssRoundDetails, shareRecoveryMessage
}
