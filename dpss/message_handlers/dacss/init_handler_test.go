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
)

/*
Function: Process

This function checks the happy path:
 1. The node executing the function is an old node.
 2. After executing the Process function, the node should receive B / (n - 2t)
    signals to start the shares of random numbers.
*/
func TestInitMessage(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	defaultSetup := testutils.DefaultTestSetup()
	testDealer := defaultSetup.GetSingleOldNodeFromTestSetup()
	transport := testDealer.Transport

	n, k, _ := testDealer.Params()

	const N_SECRETS int = 30

	msg, err := createTestMsg(testDealer, N_SECRETS)
	if err != nil {
		test.Error("Error creating the init message.")
	}

	msg.Process(testDealer.Details(), testDealer)

	// Wait a bit until all the goroutines are finished.
	time.Sleep(time.Millisecond * 500)

	recvMsgAmmount := len(transport.GetSentMessages())
	realRecvMsgAmmount := N_SECRETS / (n - 2*k)
	assert.Equal(test, realRecvMsgAmmount, recvMsgAmmount)
}

// Creates an init message for testing with a fiven ammount of old shares.
func createTestMsg(testDealer *testutils.PssTestNode, nSecrets int) (*InitMessage, error) {
	id := big.NewInt(1)
	roundID := common.NewPSSRoundID(*id)
	shares, err := generateOldShares(nSecrets, common.SECP256K1)
	if err != nil {
		return nil, err
	}

	msg := &InitMessage{
		RoundID:          roundID,
		OldShares:        shares,
		EphemeralKeypair: testDealer.Keypair,
		Kind:             InitMessageType,
		CurveName:        &common.SECP256K1,
	}
	return msg, nil
}

// Creates multiple shares for the node to simulate that the node holds
// its shares of multiple secrets.
func generateOldShares(nSecrets int, curveName common.CurveName) ([]sharing.ShamirShare, error) {
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
