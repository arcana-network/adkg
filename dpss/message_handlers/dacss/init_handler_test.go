package dacss

import (
	"crypto/rand"
	"math/big"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/torusresearch/bijson"
)

/*
Function: Process

This function checks the happy path:
 1. The node executing the function is an old node.
 2. After executing the Process function, the node should receive B / (n - 2t)
    signals to start the shares of random numbers.
*/
func TestProcessInitMessage(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	defaultSetup := testutils.DefaultTestSetup()
	testDealer := defaultSetup.GetSingleOldNodeFromTestSetup()
	transport := testDealer.Transport()

	n, k, t := testDealer.Params()

	const N_SECRETS int = 30
	ephemeralKeypair := common.GenerateKeyPair(testutils.TestCurve())
	msg, err := createTestMsg(testDealer, N_SECRETS, n, k, ephemeralKeypair)
	if err != nil {
		test.Error("Error creating the init message.")
	}

	go msg.Process(testDealer.Details(), testDealer)

	realRecvMsgAmmount := N_SECRETS / (n - 2*t)

	// Wait for the messages to be sent.
	transport.WaitForMessagesSent(realRecvMsgAmmount)

	recvMsgAmmount := len(transport.GetSentMessages())
	assert.Equal(test, realRecvMsgAmmount, recvMsgAmmount)
}

// Tests that the creation of the messages is done correctly.
func TestNewInitMessage(test *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	testDealer := defaultSetup.GetSingleOldNodeFromTestSetup()
	const N_SECRETS int = 30
	n, k, _ := testDealer.Params()

	ephemeralKeypair := common.GenerateKeyPair(testutils.TestCurve())
	msg, err := createTestMsg(testDealer, N_SECRETS, n, k, ephemeralKeypair)
	if err != nil {
		test.Errorf("Error creating the reference message.")
	}

	createdMsgBytes, err := NewInitMessage(
		msg.PSSRoundDetails,
		msg.OldShares,
		*msg.CurveName,
		ephemeralKeypair,
		defaultSetup.NewCommitteeParams,
	)
	if err != nil {
		test.Errorf("Error creating the message using the function: %v", err)
	}

	var createdInitMsg InitMessage
	err = bijson.Unmarshal(createdMsgBytes.Data, &createdInitMsg)
	assert.Nil(test, err)

	// Asserts that the values are the same
	assert.Equal(test, msg.PSSRoundDetails, createdInitMsg.PSSRoundDetails)
	assert.Equal(test, msg.OldShares, createdInitMsg.OldShares)
	assert.Equal(test, *msg.CurveName, *createdInitMsg.CurveName)
	assert.Equal(test, msg.EphemeralPublicKey, createdInitMsg.EphemeralPublicKey)
	assert.Equal(test, msg.EphemeralSecretKey, createdInitMsg.EphemeralSecretKey)
}

// Tests that if the node is from the new committee, it should do nothing.
func TestNewCommitteeDoNothing(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	defaultSetup := testutils.DefaultTestSetup()
	testDealer := defaultSetup.GetSingleNewNodeFromTestSetup()
	transport := testDealer.Transport()

	n, k, _ := testDealer.Params()
	ephemeralKey := common.GenerateKeyPair(common.CurveFromName(common.SECP256K1))

	const N_SECRETS int = 30

	msg, err := createTestMsg(testDealer, N_SECRETS, n, k, ephemeralKey)
	if err != nil {
		test.Error("Error creating the init message.")
	}

	go msg.Process(testDealer.Details(), testDealer)

	// Wait a bit until all the goroutines are finished. We use time here
	// because no message is sent.
	time.Sleep(time.Second)

	// The party should not send any message
	recvMsgAmmount := len(transport.GetSentMessages())
	assert.Equal(test, 0, recvMsgAmmount)
}

// Creates an init message for testing with a fiven ammount of old shares.
func createTestMsg(testDealer *testutils.PssTestNode, nSecrets, n, k int, ephemeralKeypair common.KeyPair) (*InitMessage, error) {
	id := big.NewInt(1)
	roundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: testDealer.Details(),
	}

	shares, err := generateOldShares(nSecrets, n, k, common.SECP256K1)
	if err != nil {
		return nil, err
	}

	msg := &InitMessage{
		PSSRoundDetails:    roundDetails,
		OldShares:          shares,
		EphemeralSecretKey: ephemeralKeypair.PrivateKey.Bytes(),
		EphemeralPublicKey: ephemeralKeypair.PublicKey.ToAffineCompressed(),
		Kind:               InitMessageType,
		CurveName:          &common.SECP256K1,
	}
	return msg, nil
}

// Creates multiple shares for the node to simulate that the node holds
// its shares of multiple secrets.
func generateOldShares(nSecrets, n, k int, curveName common.CurveName) ([]common.PrivKeyShare, error) {
	curve := common.CurveFromName(curveName)
	shares := make([]common.PrivKeyShare, nSecrets)
	shamir, err := sharing.NewShamir(uint32(k), uint32(n), curve)
	if err != nil {
		return nil, err
	}
	for i := range nSecrets {
		secret := curve.Scalar.Random(rand.Reader)
		sharesSecret, err := shamir.Split(secret, rand.Reader)
		if err != nil {
			return nil, err
		}
		shares[i] = common.PrivKeyShare{
			UserIdOwner: "userID0",
			Share:       *sharesSecret[0],
		}
	}
	return shares, nil
}
