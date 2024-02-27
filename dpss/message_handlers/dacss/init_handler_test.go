package dacss

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
	transport := testDealer.Transport

	n, k, _ := testDealer.Params()

	const N_SECRETS int = 30
	ephemeralKeypair := common.GenerateKeyPair(curves.K256())
	msg, err := createTestMsg(testDealer, N_SECRETS, n, k, ephemeralKeypair)
	if err != nil {
		test.Error("Error creating the init message.")
	}

	msg.Process(testDealer.Details(), testDealer)

	// Wait a bit until all the goroutines are finished.
	time.Sleep(time.Second)

	recvMsgAmmount := len(transport.GetSentMessages())
	realRecvMsgAmmount := N_SECRETS / (n - 2*k)
	assert.Equal(test, realRecvMsgAmmount, recvMsgAmmount)
}

// Tests that the creation of the messages is done correctly.
func TestNewInitMessage(test *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	testDealer := defaultSetup.GetSingleOldNodeFromTestSetup()
	const N_SECRETS int = 30
	n, k, _ := testDealer.Params()

	ephemeralKeypair := common.GenerateKeyPair(curves.K256())
	msg, err := createTestMsg(testDealer, N_SECRETS, n, k, ephemeralKeypair)
	if err != nil {
		test.Errorf("Error creating the reference message.")
	}

	createdMsgBytes, err := NewInitMessage(
		msg.RoundID,
		msg.OldShares,
		*msg.CurveName,
		ephemeralKeypair,
	)
	if err != nil {
		test.Errorf("Error creating the message using the function: %v", err)
	}

	var createdInitMsg InitMessage
	json.Unmarshal(createdMsgBytes.Data, &createdInitMsg)

	// Asserts that the values are the same
	assert.Equal(test, msg.RoundID, createdInitMsg.RoundID)
	assert.Equal(test, msg.OldShares, createdInitMsg.OldShares)
	assert.Equal(test, *msg.CurveName, *createdInitMsg.CurveName)
	assert.Equal(test, msg.EphemeralPublicKey, createdInitMsg.EphemeralPublicKey)
	assert.Equal(test, msg.EphemeralSecretKey, createdInitMsg.EphemeralSecretKey)
}

// Creates an init message for testing with a fiven ammount of old shares.
func createTestMsg(testDealer *testutils.PssTestNode, nSecrets, n, k int, ephemeralKeypair common.KeyPair) (*InitMessage, error) {
	id := big.NewInt(1)
	roundID := common.NewPSSRoundID(*id)
	shares, err := generateOldShares(nSecrets, n, k, common.SECP256K1)
	if err != nil {
		return nil, err
	}

	msg := &InitMessage{
		RoundID:            roundID,
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
func generateOldShares(nSecrets, n, k int, curveName common.CurveName) ([]sharing.ShamirShare, error) {
	curve := common.CurveFromName(curveName)
	shares := make([]sharing.ShamirShare, nSecrets)
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
		shares[i] = *sharesSecret[0]
	}
	return shares, nil
}
