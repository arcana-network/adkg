package dacss

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Nodes(Old and New both) receives a Propose msg from the dealer
// stores the state
// verifies the shares and commitment
// if everything passes, send an eco msg

func TestProcessProposeMessage(t *testing.T) {

	defaultSetup := testutils.DefaultTestSetup()

	SingleOldNode := defaultSetup.GetSingleOldNodeFromTestSetup()
	singleNewNode := defaultSetup.GetSingleNewNodeFromTestSetup()

	transport := SingleOldNode.Transport

	msgOldCommittee := getTestValidProposeMsg(SingleOldNode, defaultSetup, false)

	// Call the process on the msg
	msgOldCommittee.Process(SingleOldNode.Details(), SingleOldNode)

	sent_msg := transport.GetSentMessages()
	assert.Equal(t, len(sent_msg), defaultSetup.OldCommitteeParams.N)

	msgNewCommittee := getTestValidProposeMsg(singleNewNode, defaultSetup, true)
	msgNewCommittee.Process(singleNewNode.Details(), singleNewNode)

	sent_msg = transport.GetSentMessages()

	for i := 0; i < defaultSetup.NewCommitteeParams.N+defaultSetup.OldCommitteeParams.N; i++ {
		assert.Equal(t, sent_msg[i].Type, DacssEchoMessageType)

	}
	//total length of the transport msg = length of the old msgs + new msgs
	assert.Equal(t, len(sent_msg), defaultSetup.NewCommitteeParams.N+defaultSetup.OldCommitteeParams.N)

	//check states update correctly
	acssState, _, _ := SingleOldNode.State().AcssStore.Get(msgOldCommittee.ACSSRoundDetails.ToACSSRoundID())
	assert.Equal(t, len(acssState.ImplicateInformationSlice), 0)

	acssState, _, _ = singleNewNode.State().AcssStore.Get(msgNewCommittee.ACSSRoundDetails.ToACSSRoundID())
	assert.Equal(t, len(acssState.ImplicateInformationSlice), 0)
}

// Test for invalid share
// if not verified, then the node should send implicate to all the other node
func TestInvalidShare(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()

	//invalid msg construction
	SingleOldNode := defaultSetup.GetSingleOldNodeFromTestSetup()
	transport := SingleOldNode.Transport

	msgOldCommittee := getTestValidProposeMsg(SingleOldNode, defaultSetup, false)

	//constructing an invalid share for node0
	X := SingleOldNode.Details().PubKey.X
	Y := SingleOldNode.Details().PubKey.Y
	pubKeyPoint, _ := curves.K256().NewIdentityPoint().Set(&X, &Y)

	bytes := make([]byte, 33)
	_, _ = rand.Read(bytes)

	msgOldCommittee.Data.ShareMap[hex.EncodeToString(pubKeyPoint.ToAffineCompressed())] = sharing.ShamirShare{
		Id:    0,
		Value: bytes, //random bytes instead of actual share
	}.Bytes()

	// Call the process on the msg
	// should send an implicate
	msgOldCommittee.Process(SingleOldNode.Details(), SingleOldNode)

	sent_msg := transport.GetSentMessages()

	for i := 0; i < defaultSetup.OldCommitteeParams.N; i++ {
		assert.Equal(t, sent_msg[i].Type, ImplicateReceiveMessageType)

	}

	assert.Equal(t, len(sent_msg), defaultSetup.OldCommitteeParams.N)

	// check state updates correctly
	acssState, _, _ := SingleOldNode.State().AcssStore.Get(msgOldCommittee.ACSSRoundDetails.ToACSSRoundID())

	assert.Equal(t, len(acssState.ImplicateInformationSlice), 0)

}

// when sender is not dealer then it should trigger an early return
func TestSenderNotEqualToDealer(t *testing.T) {

	defaultSetup := testutils.DefaultTestSetup()
	node0, node1 := defaultSetup.GetTwoOldNodesFromTestSetup()
	transport := node1.Transport
	msgOldCommittee := getTestValidProposeMsg(node0, defaultSetup, false)

	// Call the process on the msg
	// should trigger an early return since dealer != sender
	msgOldCommittee.Process(node1.Details(), node0)

	sent_msg := transport.GetSentMessages()
	assert.Equal(t, len(sent_msg), 0)

	//check states should not be present
	_, isPresent, _ := node0.State().AcssStore.Get(msgOldCommittee.ACSSRoundDetails.ToACSSRoundID())
	assert.Equal(t, isPresent, false)

}

// Test if the shares already received then it should trigger an early return
func TestShareAlreadyReceived(t *testing.T) {

	defaultSetup := testutils.DefaultTestSetup()
	node0, node1 := defaultSetup.GetTwoOldNodesFromTestSetup()
	transport := node0.Transport
	msgOldCommittee := getTestValidProposeMsg(node0, defaultSetup, false)

	// The node already received the shares, and thus stored the acssDataHash
	node1.State().AcssStore.UpdateAccsState(msgOldCommittee.ACSSRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = []byte("test")
	})

	// Call the process on the msg
	msgOldCommittee.Process(node0.Details(), node1)

	//shold trigger an early return
	sent_msg := transport.GetSentMessages()
	assert.Equal(t, len(sent_msg), 0)

	//state should be updated to true
	_, isPresent, _ := node1.State().AcssStore.Get(msgOldCommittee.ACSSRoundDetails.ToACSSRoundID())
	assert.Equal(t, isPresent, true)

}

/*
Function: Process

Testcase: 2 ImplicateInformations have been stored in Node's state.

Expectation: Node should send itself 2 implicateExecuteMessages
*/
func TestAlreadyInImplicateFlow(t *testing.T) {

	defaultSetup := testutils.DefaultTestSetup()
	node0, node1, node2 := defaultSetup.GetThreeOldNodesFromTestSetup()
	transport := node0.Transport
	msgOldCommittee := getTestValidProposeMsg(node0, defaultSetup, false)

	curveName := msgOldCommittee.CurveName
	curve := common.CurveFromName(curveName)

	// create dummy implicate information
	randomSymmetricKey := curve.Point.Random(rand.Reader).ToAffineCompressed()

	randomProofBytes := make([]byte, 99)
	_, _ = rand.Read(randomProofBytes)

	implicateInfo1 := common.ImplicateInformation{
		SymmetricKey:    randomSymmetricKey,
		Proof:           randomProofBytes,
		SenderPubkeyHex: hex.EncodeToString(node0.LongtermKey.PublicKey.ToAffineCompressed()),
	}

	implicateInfo2 := common.ImplicateInformation{
		SymmetricKey:    randomSymmetricKey,
		Proof:           randomProofBytes,
		SenderPubkeyHex: hex.EncodeToString(node2.LongtermKey.PublicKey.ToAffineCompressed()),
	}

	// Add both pieces of implicateInformation to the receiver's node state
	err := node1.State().AcssStore.UpdateAccsState(msgOldCommittee.ACSSRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.ImplicateInformationSlice = []common.ImplicateInformation{implicateInfo1, implicateInfo2}
	})

	assert.Nil(t, err)

	// Call the process on the msg
	msgOldCommittee.Process(node0.Details(), node1)

	// Check: Node should send itself 2 implicateExecuteMessages
	sent_msgs := transport.ReceivedMessages
	implicateExecuteMessages := []common.PSSMessage{}
	for _, msg := range sent_msgs {
		if msg.Type == ImplicateExecuteMessageType {
			implicateExecuteMessages = append(implicateExecuteMessages, msg)
		}
	}

	assert.Equal(t, 2, len(implicateExecuteMessages))
}

func getTestValidProposeMsg(SingleNode *testutils.PssTestNode, defaultSetup *testutils.TestSetup, newCommittee bool) AcssProposeMessage {

	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: SingleNode.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: pssRoundDetails,
		ACSSCount:       1,
	}

	k := 0
	n := 0

	if newCommittee {
		k = defaultSetup.NewCommitteeParams.K
		n = defaultSetup.NewCommitteeParams.K
	} else {
		k = defaultSetup.OldCommitteeParams.K
		n = defaultSetup.OldCommitteeParams.N
	}

	DealerEphemeralKey := common.GenerateKeyPair(curves.K256())
	testSecret := sharing.GenerateSecret(curves.K256())
	commitments, shares, _ := sharing.GenerateCommitmentAndShares(testSecret, uint32(k), uint32(n), curves.K256())
	compressedCommitments := sharing.CompressCommitments(commitments)
	shareMap := make(map[string][]byte, n)

	for _, share := range shares {

		nodePublicKey := SingleNode.GetPublicKeyFor(int(share.Id), newCommittee)
		if nodePublicKey == nil {
			log.Errorf("Couldn't obtain public key for node with id=%v", share.Id)
		}

		cipherShare, err := sharing.EncryptSymmetricCalculateKey(share.Bytes(), nodePublicKey, DealerEphemeralKey.PrivateKey)
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
		DealerEphemeralPubKey: hex.EncodeToString(DealerEphemeralKey.PublicKey.ToAffineCompressed()),
	}

	msg := AcssProposeMessage{
		ACSSRoundDetails:   acssRoundDetails,
		Kind:               AcssProposeMessageType,
		CurveName:          common.CurveName(curves.K256().Name),
		Data:               msgData,
		NewCommittee:       newCommittee,
		NewCommitteeParams: defaultSetup.NewCommitteeParams,
	}
	return msg
}
