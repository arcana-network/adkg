package dacss

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/torusresearch/bijson"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TODO: Incomplete
func TestDacss(t *testing.T) {
	// timeout := time.After(30 * time.Second)
	// done := make(chan bool)

	// curve := curves.K256()
	log.SetLevel(log.DebugLevel)

	//default setup and mock transport
	TestSetUp, _ := DefaultTestSetup()

	nodesOld := TestSetUp.oldCommitteeNetwork
	// nodesNew := TestSetUp.newCommitteeNetwork

	nOld := TestSetUp.OldCommitteeParams.N
	kOld := TestSetUp.OldCommitteeParams.K

	//generating random old secrets of old nodes which will be re-shared
	// secret := curve.Scalar.Random(rand.Reader)
	// _, shares, err := sharing.GenerateCommitmentAndShares(secret, uint32(kOld), uint32(nOld), curve)

	InitMsgs := make(map[common.NodeDetailsID]*dacss.InitMessage, nOld)

	for _, n := range nodesOld {
		ephemeralKeypair := common.GenerateKeyPair(curves.K256())
		InitMsg, err := createTestMsg(n, 10, nOld, kOld, ephemeralKeypair)
		assert.Nil(t, err)

		//save the init msg against the nodes
		InitMsgs[n.details.GetNodeDetailsID()] = InitMsg
	}

	// id := common.NewPssID(*big.NewInt(int64(1)))

	for _, n := range nodesOld {
		// go func(node *PssTestNode2) {

		initMsg := *InitMsgs[n.details.GetNodeDetailsID()]

		pssMsgData, err := bijson.Marshal(initMsg)
		assert.Nil(t, err)

		InitPssMessage := common.PSSMessage{
			PSSRoundDetails: initMsg.PSSRoundDetails,
			Type:            initMsg.Kind,
			Data:            pssMsgData,
		}
		n.ReceiveMessage(n.Details(), InitPssMessage)
		// }(n)
	}
}

// taken from the dacss init handler test
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

// taken from the dacss init handler test
// Creates an init message for testing with a given ammount of old shares.
func createTestMsg(testDealer *PssTestNode2, nSecrets, n, k int, ephemeralKeypair common.KeyPair) (*dacss.InitMessage, error) {
	id := big.NewInt(1)
	roundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: testDealer.Details(),
	}

	shares, err := generateOldShares(nSecrets, n, k, common.SECP256K1)
	if err != nil {
		return nil, err
	}

	msg := &dacss.InitMessage{
		PSSRoundDetails:    roundDetails,
		OldShares:          shares,
		EphemeralSecretKey: ephemeralKeypair.PrivateKey.Bytes(),
		EphemeralPublicKey: ephemeralKeypair.PublicKey.ToAffineCompressed(),
		Kind:               dacss.InitMessageType,
		CurveName:          &common.SECP256K1,
	}
	return msg, nil
}
