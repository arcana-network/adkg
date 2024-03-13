package dacss

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/stretchr/testify/assert"
	"github.com/torusresearch/bijson"

	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	krsharing "github.com/coinbase/kryptology/pkg/sharing"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

/*
Function: Process

Testcase: Successful reception of echo message. An old node receives the ECHO
message which is a matching message. The party will change the RBC state
accordingly and send a message

Expectations:
- the node increments the counter of echo messages.
- no READY message is sent.
*/
func TestIncrement(test *testing.T) {
	// Setup the parties
	defaultSetup := testutils.DefaultTestSetup()
	testSender, testRecvr := defaultSetup.GetTwoOldNodesFromTestSetup()
	transport := testRecvr.Transport

	// Set the round parameters
	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: testSender.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: pssRoundDetails,
		ACSSCount:       1,
	}

	ephemeralKeypairSender := common.GenerateKeyPair(curves.K256())

	shards, hashMsg, err := createShardAndHash(
		testSender,
		ephemeralKeypairSender,
	)
	if err != nil {
		test.Errorf("Error generating the shard: %v", err)
	}

	shardReceiver := shards[testRecvr.Details().Index]
	testRecvr.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.AcssDataHash = hashMsg
			state.RBCState.OwnReedSolomonShard = shardReceiver
		},
	)

	echoMsg := DacssEchoMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             DacssEchoMessageType,
		CurveName:        common.CurveName(curves.K256().Name),
		Share:            shardReceiver,
		Hash:             hashMsg,
		NewCommittee:     testRecvr.IsNewNode(),
	}

	echoMsg.Process(testSender.Details(), testRecvr)

	testRecvr.State().AcssStore.Lock()

	acssState, stateExists, err := testRecvr.State().AcssStore.Get(
		echoMsg.ACSSRoundDetails.ToACSSRoundID(),
	)
	if !stateExists {
		test.Errorf("The state does not exist")
	}
	if err != nil {
		test.Errorf("Error retrieving the state: %v", err)
	}
	echoDatabase := acssState.RBCState.ReceivedEcho
	for id, received := range echoDatabase {
		if id == testSender.Details().Index {
			assert.Equal(test, true, received)
		} else {
			assert.Equal(test, false, received)
		}
	}

	assert.Equal(test, 0, len(transport.BroadcastedMessages))
	assert.Equal(test, 1, acssState.RBCState.CountEcho())
	testRecvr.State().AcssStore.Unlock()
}

/*
Function: Process

Testcase: the same party sends 2t + 1 matching ECHO messages. The ECHO count
should not increment more than 1 message and no READY message should be sent.

Expectations:
- the node increments the counter of echo messages just once.
- no READY message is sent.
*/
func TestCounterDoesNotIncrement(test *testing.T) {
	// Setup the parties
	defaultSetup := testutils.DefaultTestSetup()
	testSender, testRecvr := defaultSetup.GetTwoOldNodesFromTestSetup()
	transport := testRecvr.Transport

	// Set the round parameters
	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: testSender.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: pssRoundDetails,
		ACSSCount:       1,
	}

	ephemeralKeypairSender := common.GenerateKeyPair(curves.K256())

	shards, hashMsg, err := createShardAndHash(
		testSender,
		ephemeralKeypairSender,
	)

	shardReceiver := shards[testRecvr.Details().Index]
	testRecvr.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.RBCState.OwnReedSolomonShard = shardReceiver
			state.AcssDataHash = hashMsg
		},
	)

	if err != nil {
		test.Errorf("Error generating the shard: %v", err)
	}

	echoMsg := DacssEchoMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             DacssEchoMessageType,
		CurveName:        common.CurveName(curves.K256().Name),
		Share:            shardReceiver,
		Hash:             hashMsg,
		NewCommittee:     testRecvr.IsNewNode(),
	}

	if err != nil {
		test.Errorf("Error creating the echo message: %v", err)
	}

	_, _, t := testSender.Params()

	// Executes the process message simulating that the sender node sends the
	// same matching message 2t + 1 times.
	for range 2*t + 1 {
		echoMsg.Process(testSender.Details(), testRecvr)
	}

	testRecvr.State().AcssStore.Lock()

	acssState, stateExists, err := testRecvr.State().AcssStore.Get(
		echoMsg.ACSSRoundDetails.ToACSSRoundID(),
	)
	if !stateExists {
		test.Errorf("The state does not exist")
	}
	if err != nil {
		test.Errorf("Error retrieving the state: %v", err)
	}
	echoDatabase := acssState.RBCState.ReceivedEcho
	for id, received := range echoDatabase {
		if id == testSender.Details().Index {
			assert.Equal(test, true, received)
		} else {
			assert.Equal(test, false, received)
		}
	}

	assert.Equal(test, 0, len(transport.BroadcastedMessages))
	assert.Equal(test, 1, acssState.RBCState.CountEcho())
	testRecvr.State().AcssStore.Unlock()
}

/*
Function: Process

Test case: happy path. The receiver node receives 2t + 1 ECHO messages and then,
it sends the corresponding ready message to all parties.

Expectations:
- the node increments the counter of echo messages.
- the node sends a READY message to all parties.
*/
func TestCounterEchoMessages(test *testing.T) {
	const oldCommittee = true
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode, senderGroup := defaultSetup.GetDealerAnd2kPlusOneNodes(oldCommittee)
	transport := receiverNode.Transport

	// The dealer node will be the first node in the set of 2k + 1 nodes.
	dealerNode := senderGroup[0]

	// Set the round parameters
	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: dealerNode.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: pssRoundDetails,
		ACSSCount:       1,
	}

	receiverNode.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(as *common.AccsState) {},
	)

	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

	shards, hashMsg, err := createShardAndHash(
		dealerNode,
		ephemeralKeypairDealer,
	)
	if err != nil {
		test.Errorf("Error generating the shard: %v", err)
	}

	shardReceiver := shards[receiverNode.Details().Index]

	receiverNode.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.AcssDataHash = hashMsg
			state.RBCState.OwnReedSolomonShard = shardReceiver
		},
	)

	for _, senderNode := range senderGroup {

		echoMsg := DacssEchoMessage{
			ACSSRoundDetails: acssRoundDetails,
			Kind:             DacssEchoMessageType,
			CurveName:        common.CurveName(curves.K256().Name),
			Share:            shardReceiver,
			Hash:             hashMsg,
			NewCommittee:     receiverNode.IsNewNode(),
		}
		echoMsg.Process(senderNode.Details(), receiverNode)
	}

	receiverNode.State().AcssStore.Lock()
	defer receiverNode.State().AcssStore.Unlock()

	acssState, stateExists, err := receiverNode.State().AcssStore.Get(
		acssRoundDetails.ToACSSRoundID(),
	)
	if !stateExists {
		test.Errorf("The state does not exist")
	}
	if err != nil {
		test.Errorf("Error retrieving the state: %v", err)
	}
	_, _, t := dealerNode.Params()

	// Tests that the eco count is 2t + 1.
	assert.Equal(test, 2*t+1, acssState.RBCState.CountEcho())

	// Test that a ready message was sent.
	broadcastedMsgs := transport.BroadcastedMessages
	assert.Equal(test, 1, len(broadcastedMsgs))
	assert.Equal(test, AcssReadyMessageType, broadcastedMsgs[0].Type)
}

/*
Function: Process

Test case: the receiver of the ECHO message receives 2t + 1 messages but he
already sent a READY message.

Expectations:
- The counter increments because the party receives 2t + 1 echo messages.
- No READY message is broadcasted.
*/
func TestNotSendIfReadyMessageAlreadySent(test *testing.T) {
	const oldCommittee = true
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode, senderGroup := defaultSetup.GetDealerAnd2kPlusOneNodes(oldCommittee)
	transport := receiverNode.Transport

	// The dealer node will be the first node in the set of 2k + 1 nodes.
	dealerNode := senderGroup[0]

	// Set the round parameters
	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: dealerNode.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: pssRoundDetails,
		ACSSCount:       1,
	}

	receiverNode.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(as *common.AccsState) {},
	)

	receiverNode.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.RBCState.IsReadyMsgSent = true
		},
	)

	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())
	shards, hashMsg, err := createShardAndHash(
		dealerNode,
		ephemeralKeypairDealer,
	)
	if err != nil {
		test.Errorf("Error generating the shard: %v", err)
	}

	shardReceiver := shards[receiverNode.Details().Index]
	receiverNode.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.AcssDataHash = hashMsg
			state.RBCState.OwnReedSolomonShard = shardReceiver
		},
	)

	for _, senderNode := range senderGroup {
		echoMsg := DacssEchoMessage{
			ACSSRoundDetails: acssRoundDetails,
			Kind:             DacssEchoMessageType,
			CurveName:        common.CurveName(curves.K256().Name),
			Share:            shardReceiver,
			Hash:             hashMsg,
			NewCommittee:     receiverNode.IsNewNode(),
		}

		echoMsg.Process(senderNode.Details(), receiverNode)
	}

	receiverNode.State().AcssStore.Lock()
	defer receiverNode.State().AcssStore.Unlock()

	acssState, stateExists, err := receiverNode.State().AcssStore.Get(
		acssRoundDetails.ToACSSRoundID(),
	)
	if !stateExists {
		test.Errorf("The state does not exist")
	}
	if err != nil {
		test.Errorf("Error retrieving the state: %v", err)
	}
	_, _, t := dealerNode.Params()

	// Tests that the eco count is 2t + 1.
	assert.Equal(test, 2*t+1, acssState.RBCState.CountEcho())

	// Test that a ready message was sent.
	broadcastedMsgs := transport.BroadcastedMessages
	assert.Equal(test, 0, len(broadcastedMsgs))
}

func createShardAndHash(
	dealerNode *testutils.PssTestNode,
	ephemeralKeypairDealer common.KeyPair,
) ([]infectious.Share, []byte, error) {
	// Creates the Reed-Solomon shards for the message.
	n, k, _ := dealerNode.Params()
	secret := sharing.GenerateSecret(curves.K256())
	commitments, shares, err := sharing.GenerateCommitmentAndShares(
		secret,
		uint32(k),
		uint32(n),
		curves.K256(),
	)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	shards, hashMsg, err := computeReedSolomonShardsAndHash(
		commitments,
		dealerNode,
		shares,
		ephemeralKeypairDealer,
	)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	return shards, hashMsg, nil
}

// Computes the Reed-Solomon shards and hash of a given commitment and shares.
func computeReedSolomonShardsAndHash(
	commitment *krsharing.FeldmanVerifier,
	dealer *testutils.PssTestNode,
	shares []*krsharing.ShamirShare,
	dealerEphemeralKey common.KeyPair,
) ([]infectious.Share, []byte, error) {
	n, _, t := dealer.Params()
	compressedCommitments := sharing.CompressCommitments(commitment)
	shareMap := make(map[string][]byte, n)
	for _, share := range shares {
		nodePublicKey := dealer.GetPublicKeyFor(int(share.Id), dealer.IsNewNode())
		if nodePublicKey == nil {
			log.Errorf("Couldn't obtain public key for node with id=%v", share.Id)
			return []infectious.Share{}, []byte{}, errors.New("Public key is nil")
		}

		cipherShare, err := sharing.EncryptSymmetricCalculateKey(
			share.Bytes(),
			nodePublicKey,
			dealerEphemeralKey.PrivateKey,
		)

		if err != nil {
			log.Errorf("Error while encrypting secret share, err=%v", err)
			return []infectious.Share{}, []byte{}, errors.New("Can't been able to encrypt the shares")
		}
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		pubkeyHex := common.PointToHex(nodePublicKey)
		shareMap[pubkeyHex] = cipherShare
	}

	msgData := common.AcssData{
		Commitments:           compressedCommitments,
		ShareMap:              shareMap,
		DealerEphemeralPubKey: hex.EncodeToString(dealerEphemeralKey.PrivateKey.Bytes()),
	}

	msgBytes, err := bijson.Marshal(msgData)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	msgHash := common.HashByte(msgBytes)

	fec, err := infectious.NewFEC(t+1, n)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	shards, err := acss.Encode(fec, msgBytes)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	return shards, msgHash, nil
}
