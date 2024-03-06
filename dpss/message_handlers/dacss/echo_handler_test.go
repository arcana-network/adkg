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
*/
func TestIncrement(test *testing.T) {
	// Setup the parties
	defaultSetup := testutils.DefaultTestSetup()
	testSender, testRecvr := defaultSetup.GetTwoOldNodesFromTestSetup()

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

	// Initialize the AcssStore as default
	testRecvr.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(as *common.AccsState) {},
	)

	echoMsg, err := getTestEchoMsg(
		testSender,
		testRecvr,
		ephemeralKeypairSender,
		acssRoundDetails,
	)
	if err != nil {
		test.Errorf("Error creating the echo message: %v", err)
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
	testRecvr.State().AcssStore.Unlock()
}

func getTestEchoMsg(
	sender *testutils.PssTestNode,
	receiver *testutils.PssTestNode,
	ephemeralKey common.KeyPair,
	acssRoundDetails common.ACSSRoundDetails,
) (DacssEchoMessage, error) {

	n, k, _ := sender.Params()

	secret := sharing.GenerateSecret(curves.K256())
	commitment, shares, err := sharing.GenerateCommitmentAndShares(
		secret,
		uint32(k),
		uint32(n),
		curves.K256(),
	)
	if err != nil {
		log.Errorf("Error creating the secret sharing: %v", err)
		return DacssEchoMessage{}, err
	}

	shards, hashMsg, err := computeReedSolomonShardsAndHash(
		commitment,
		sender,
		shares,
		ephemeralKey,
		n,
		k,
	)
	if err != nil {
		log.Errorf("Error creating the RS shards sharing: %v", err)
		return DacssEchoMessage{}, err
	}

	shardReceiver := shards[receiver.Details().Index]

	receiver.State().AcssStore.Lock()

	recvState, stateExists, err := receiver.State().AcssStore.Get(acssRoundDetails.ToACSSRoundID())
	if !stateExists {
		log.Errorf("Error retrieving the sharing - : %v", err)
		return DacssEchoMessage{}, errors.New("Error retrieving the state of the node. The state does not exists")
	}
	if err != nil {
		log.Errorf("Error retrieving the sharing: %v", err)
		return DacssEchoMessage{}, err
	}

	// Store the shards in the internal state of the node.
	recvState.RBCState.HashMsg = hashMsg
	recvState.RBCState.OwnReedSolomonShard = shardReceiver

	receiver.State().AcssStore.Unlock()

	msg := DacssEchoMessage{
		ACSSRoundDetails: acssRoundDetails,
		NewCommittee:     sender.IsOldNode(),
		Kind:             DacssEchoMessageType,
		CurveName:        common.CurveName(curves.K256().Name),
		Share:            shardReceiver,
		Hash:             hashMsg,
	}
	return msg, nil
}

func computeReedSolomonShardsAndHash(
	commitment *krsharing.FeldmanVerifier,
	sender *testutils.PssTestNode,
	shares []*krsharing.ShamirShare,
	ephemeralKey common.KeyPair,
	n int,
	k int,
) ([]infectious.Share, []byte, error) {
	compressedCommitments := sharing.CompressCommitments(commitment)
	shareMap := make(map[string][]byte, n)
	for _, share := range shares {
		nodePublicKey := sender.GetPublicKeyFor(int(share.Id), sender.IsOldNode())
		if nodePublicKey == nil {
			log.Errorf("Couldn't obtain public key for node with id=%v", share.Id)
			return []infectious.Share{}, []byte{}, errors.New("Public key is nil")
		}

		cipherShare, err := sharing.EncryptSymmetricCalculateKey(
			share.Bytes(),
			nodePublicKey,
			ephemeralKey.PrivateKey,
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
		DealerEphemeralPubKey: hex.EncodeToString(ephemeralKey.PrivateKey.Bytes()),
	}

	msgBytes, err := bijson.Marshal(msgData)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	msgHash := common.HashByte(msgBytes)

	fec, err := infectious.NewFEC(k, n)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	shards, err := acss.Encode(fec, msgHash)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	return shards, msgHash, nil
}
