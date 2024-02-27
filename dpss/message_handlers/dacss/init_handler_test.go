package dacss

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
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

	const N_SECRETS int = 30

	id := big.NewInt(1)

	node, transport := getSingleNode(false)

	round := common.PSSRoundDetails{
		PSSRoundID: common.GeneratePSSRoundID(*id),
		Dealer:     node.details.Index,
		Kind:       InitMessageType,
	}

	n, k, _ := node.Params()

	curve := common.CurveFromName(common.SECP256K1)
	reader := rand.Reader

	// Creates multiple shares for the node to simulate that the node holds
	// its shares of multiple secrets.
	shares := make([]sharing.ShamirShare, N_SECRETS)
	shamir, err := sharing.NewShamir(uint32(k), uint32(n), curve)
	if err != nil {
		test.Errorf("Error while generating the Shamir builder: %v", err)
	}
	for i := range N_SECRETS {
		secret := curve.Scalar.Random(reader)
		sharesSecret, err := shamir.Split(secret, reader)
		if err != nil {
			test.Errorf("Error creating the shares for the secret: %v", err)
		}
		shares[i] = *sharesSecret[0]
	}

	msg, err := NewInitMessage(
		round.ID(),
		shares,
		common.SECP256K1,
	)

	if err != nil {
		test.Error("Error creating the init message.")
	}

	node.ReceiveMessage(node.Details(), *msg)

	// Checks: the node now should have B / (n - 2t) messages to start the ACSS
	// protocol.
	recvMsgAmmount := len(transport.receivedMessages)
	realRecvMsgAmmount := N_SECRETS / (n - 2*k)
	assert.Equal(test, realRecvMsgAmmount, recvMsgAmmount)
}
