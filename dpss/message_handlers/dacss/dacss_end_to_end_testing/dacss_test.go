package dacss

import (
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/torusresearch/bijson"
)

// FIXME: test is not yet passing
func TestDacss(t *testing.T) {
	log.SetLevel(log.InfoLevel)

	//default setup and mock transport
	TestSetUp, _ := DefaultTestSetup()

	nodesOld := TestSetUp.oldCommitteeNetwork
	// nodesNew := TestSetUp.newCommitteeNetwork

	nOld := TestSetUp.OldCommitteeParams.N
	kOld := TestSetUp.OldCommitteeParams.K

	// The old committee has shares of a single secret
	testSecret := sharing.GenerateSecret(curves.K256())
	_, shares, _ := sharing.GenerateCommitmentAndShares(testSecret, uint32(kOld), uint32(nOld), testutils.TestCurve())

	//store the init msgs for testing
	var initMessages sync.Map

	// Each node in old committee starts dACSS
	// That means that for the single share each node has,
	// ceil(nrShare/(nrOldNodes-2*recThreshold)) = 1 random values are sampled
	// and shared to both old & new committee
	for index, n := range nodesOld {
		go func(index int, node *PssTestNode2) {
			ephemeralKeypair := common.GenerateKeyPair(curves.K256())
			share := sharing.ShamirShare{Id: shares[index].Id, Value: shares[index].Value}
			initMsg := getTestInitMsgSingleShare(n, *big.NewInt(int64(index)), &share, ephemeralKeypair, TestSetUp.NewCommitteeParams)

			pssMsgData, err := bijson.Marshal(initMsg)
			assert.Nil(t, err)

			InitPssMessage := common.PSSMessage{
				PSSRoundDetails: initMsg.PSSRoundDetails,
				Type:            initMsg.Kind,
				Data:            pssMsgData,
			}

			//store the init msg for testing
			initMessages.Store(node.details.GetNodeDetailsID(), initMsg)

			node.ReceiveMessage(node.Details(), InitPssMessage)
		}(index, n)
	}

	time.Sleep(10 * time.Second)

	// round details for oldNode0 being the dealer
	//since only one secret is shared, the ACSSCount = 0
	retrievedMsg, found := initMessages.Load(nodesOld[0].details.GetNodeDetailsID())
	if !found {
		log.WithFields(
			log.Fields{
				"Found":   found,
				"Message": "Message not found in init messages",
			},
		).Error("TestDacss")
		t.Error("Error retrieving the init message")
	}
	message := retrievedMsg.(*dacss.InitMessage)

	acssRound := common.ACSSRoundDetails{
		PSSRoundDetails: message.PSSRoundDetails,
		ACSSCount:       0,
	}

	// hex pubkey of the dealer aka oldnode0
	pubKey := message.PSSRoundDetails.Dealer.PubKey
	pubKeyCurvePoint, err := common.PointToCurvePoint(pubKey, "secp256k1")
	assert.Nil(t, err)
	pubKeyHex := common.PointToHex(pubKeyCurvePoint)

	// getting the random secret shared
	// TODO uncomment when the initial checks pass
	state, _, err := nodesOld[0].State().AcssStore.Get(acssRound.ToACSSRoundID())
	assert.Nil(t, err)
	randomSecretShared := state.RandomSecretShared[acssRound.ToACSSRoundID()]

	//storing shares received from the oldnode0 for old committee
	var OldCommitteReceivedShareFromNodeOld0 []*sharing.ShamirShare

	for _, n := range nodesOld {

		state, _, err := n.State().AcssStore.Get(acssRound.ToACSSRoundID())
		assert.Nil(t, err)

		rbcState := state.RBCState.Phase
		assert.Equal(t, rbcState, common.Ended)
		share := state.ReceivedShares[pubKeyHex]
		// A share should have been stored for this ACSS round
		assert.Equal(t, 1, len(state.ReceivedShares))
		// FIXME: either the share is stored under the wrong key,
		// or we're searching it under the wrong key
		assert.True(t, found)

		OldCommitteReceivedShareFromNodeOld0 = append(OldCommitteReceivedShareFromNodeOld0, (*sharing.ShamirShare)(share))

	}

	// TODO when finding the correct share check passes, uncomment below code for further verification
	//For Old Committee

	//reconstructing the random secret
	shamir, err := sharing.NewShamir(testutils.DefaultK_old, testutils.DefaultN_old, curves.K256())
	assert.Nil(t, err)

	reconstructedSecret, err := shamir.Combine(OldCommitteReceivedShareFromNodeOld0...)
	assert.Nil(t, err)

	assert.Equal(t, reconstructedSecret, *randomSecretShared)

	// //For New Committee

	// //storing shares received from the oldnode0 for New committee
	// var NewCommitteReceivedShareFromNodeOld0 []*sharing.ShamirShare

	// for _, n := range nodesNew {

	// 	state, _, err := n.State().AcssStore.Get(acssRound.ToACSSRoundID())
	// 	assert.Nil(t, err)
	// 	share := state.ReceivedShares[pubKeyHex]
	// 	NewCommitteReceivedShareFromNodeOld0 = append(NewCommitteReceivedShareFromNodeOld0, (*sharing.ShamirShare)(share))

	// }

	// //reconstructing the random secret
	// shamir, err = sharing.NewShamir(testutils.DefaultK_new, testutils.DefaultN_new, curves.K256())
	// assert.Nil(t, err)

	// reconstructedSecret, err = shamir.Combine(NewCommitteReceivedShareFromNodeOld0...)
	// assert.Nil(t, err)

	// assert.Equal(t, reconstructedSecret, *randomSecretShared)
}

func getTestInitMsgSingleShare(testDealer *PssTestNode2, pssRoundIndex big.Int, share *sharing.ShamirShare, ephemeralKeypair common.KeyPair, newCommitteeParams common.CommitteeParams) *dacss.InitMessage {
	roundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(pssRoundIndex),
		Dealer: testDealer.Details(),
	}
	msg := &dacss.InitMessage{
		PSSRoundDetails:    roundDetails,
		OldShares:          []sharing.ShamirShare{*share},
		EphemeralSecretKey: ephemeralKeypair.PrivateKey.Bytes(),
		EphemeralPublicKey: ephemeralKeypair.PublicKey.ToAffineCompressed(),
		Kind:               dacss.InitMessageType,
		CurveName:          &common.SECP256K1,
		NewCommitteeParams: newCommitteeParams,
	}
	return msg
}
