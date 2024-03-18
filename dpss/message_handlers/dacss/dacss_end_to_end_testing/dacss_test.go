package dacss

import (
	"crypto/rand"
	"math/big"
	"testing"
	"time"

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

	curve := curves.K256()
	log.SetLevel(log.DebugLevel)

	//default setup and mock transport
	TestSetUp, _ := DefaultTestSetup()

	nodesOld := TestSetUp.oldCommitteeNetwork
	// nodesNew := TestSetUp.newCommitteeNetwork

	nOld := TestSetUp.OldCommitteeParams.N
	kOld := TestSetUp.OldCommitteeParams.K

	InitMsgs := make(map[common.NodeDetailsID]*dacss.InitMessage, nOld)

	//testing for only one share from nodesOld[0]
	var Shares sharing.ShamirShare

	for _, n := range nodesOld {
		ephemeralKeypair := common.GenerateKeyPair(curves.K256())
		InitMsg, shares, err := createTestMsg(n, 1, nOld, kOld, ephemeralKeypair)

		//storing only one share for old node0
		if n == nodesOld[0] {
			Shares = shares[0]
		}

		assert.Nil(t, err)

		//save the init msg against the nodes
		InitMsgs[n.details.GetNodeDetailsID()] = InitMsg
	}

	for _, n := range nodesOld {
		go func(node *PssTestNode2) {

			initMsg := *InitMsgs[node.details.GetNodeDetailsID()]

			pssMsgData, err := bijson.Marshal(initMsg)
			assert.Nil(t, err)

			InitPssMessage := common.PSSMessage{
				PSSRoundDetails: initMsg.PSSRoundDetails,
				Type:            initMsg.Kind,
				Data:            pssMsgData,
			}
			node.ReceiveMessage(node.Details(), InitPssMessage)
		}(n)
	}

	time.Sleep(8 * time.Second)

	//Reconstructing the share for testing

	//shares received from oldNodes[0] to all the old nodes
	var sharesReceived []*sharing.ShamirShare
	for _, node := range nodesOld {

		roundDetail := InitMsgs[nodesOld[0].details.GetNodeDetailsID()].PSSRoundDetails

		// since only one share is considered,
		// acss count is 0
		acssRound := common.ACSSRoundDetails{
			PSSRoundDetails: roundDetail,
			ACSSCount:       0,
		}

		state, _, _ := node.State().AcssStore.Get(acssRound.ToACSSRoundID())

		pubKey := nodesOld[0].details.PubKey
		pubKeyCurvePoint, err := common.PointToCurvePoint(pubKey, "secp256k1")
		if err != nil {
			log.WithField("error constructing PointToCurvePoint", err).Error("DacssOutputMessage")
			return
		}
		pubKeyHex := common.PointToHex(pubKeyCurvePoint)

		// storing all the shares received from oldNode0
		shareFromNodeOld0 := state.ReceivedShares[pubKeyHex]

		sharesReceived = append(sharesReceived, (*sharing.ShamirShare)(shareFromNodeOld0))

	}

	shamir, err := sharing.NewShamir(uint32(TestSetUp.OldCommitteeParams.K), uint32(TestSetUp.OldCommitteeParams.N), curve)
	assert.Nil(t, err)

	reconstructedValue, err := shamir.Combine(sharesReceived...)
	assert.Nil(t, err)

	secret, err := curve.Scalar.SetBytes(Shares.Value)
	assert.Nil(t, err)

	//expected to be equal
	//TODO: failing
	assert.Equal(t, reconstructedValue, secret)
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
func createTestMsg(testDealer *PssTestNode2, nSecrets, n, k int, ephemeralKeypair common.KeyPair) (*dacss.InitMessage, []sharing.ShamirShare, error) {
	id := big.NewInt(1)
	roundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: testDealer.Details(),
	}

	shares, err := generateOldShares(nSecrets, n, k, common.SECP256K1)
	if err != nil {
		return nil, nil, err
	}

	msg := &dacss.InitMessage{
		PSSRoundDetails:    roundDetails,
		OldShares:          shares,
		EphemeralSecretKey: ephemeralKeypair.PrivateKey.Bytes(),
		EphemeralPublicKey: ephemeralKeypair.PublicKey.ToAffineCompressed(),
		Kind:               dacss.InitMessageType,
		CurveName:          &common.SECP256K1,
	}
	return msg, shares, nil
}
