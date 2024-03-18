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
		InitMsg, err := createTestMsg(n, 1, nOld, kOld, ephemeralKeypair)
		assert.Nil(t, err)

		//save the init msg against the nodes
		InitMsgs[n.details.GetNodeDetailsID()] = InitMsg
	}

	// id := common.NewPssID(*big.NewInt(int64(1)))

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

	time.Sleep(12 * time.Second)

	//Reconstructing the share for testing

	//shares of oldNodes
	// Shares := make(map[common.NodeDetailsID]*sharing.ShamirShare, nOld)
	// for _, node := range nodesNew {

	// 	roundDetail := InitMsgs[nodesOld[0].details.GetNodeDetailsID()].PSSRoundDetails

	// 	// since only one share is considered,
	// 	// acss count is 0
	// 	acssRound := common.ACSSRoundDetails{
	// 		PSSRoundDetails: roundDetail,
	// 		ACSSCount:       0,
	// 	}

	// 	state, _, _ := node.State().AcssStore.Get(acssRound.ToACSSRoundID())

	// 	pubKey := acssRound.PSSRoundDetails.Dealer.PubKey
	// 	pubKeyCurvePoint, err := common.PointToCurvePoint(pubKey, "secp256k1")
	// 	if err != nil {
	// 		log.WithField("error constructing PointToCurvePoint", err).Error("DacssOutputMessage")
	// 		return
	// 	}
	// 	pubKeyHex := common.PointToHex(pubKeyCurvePoint)

	// 	// storing all the shares received from oldNode0
	// 	shareFromNodeOld0 := state.ReceivedShares[pubKeyHex]

	// 	Shares[node.details.GetNodeDetailsID()] = (*sharing.ShamirShare)(shareFromNodeOld0)

	// }
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
