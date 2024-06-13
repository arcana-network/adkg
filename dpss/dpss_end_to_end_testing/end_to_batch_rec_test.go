package DpssEndToEndTesting

import (
	"math"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/new_committee"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/old_committee"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/torusresearch/bijson"
)

func TestEndToBatchRec(t *testing.T) {
	// log.SetLevel(log.DebugLevel)

	// this is set to panicLevel so that normal info logging can be avoided during test run
	// and logging is done only if the program panics
	// can be commented out when we want info logging
	log.SetLevel(log.PanicLevel)

	//default setup and mock transport
	TestSetUp, transport := DpssEndToEndTestSetup()

	nodesOld := TestSetUp.OldCommitteeNetwork
	nodesNew := TestSetUp.NewCommitteeNetwork

	nOld := TestSetUp.OldCommitteeParams.N
	kOld := TestSetUp.OldCommitteeParams.K
	tOld := TestSetUp.OldCommitteeParams.T

	// The old committee has shares of secrets
	// 100 has passed, but 300 not
	B := 300
	_, _, shares, _ := GenerateSecretAndGetCommitmentAndShares(B, nOld, kOld)

	// number of random secret scalar to be shared
	nGenerations := int(math.Ceil(float64(B) / float64((nOld - 2*tOld))))
	//store the init msgs for testing
	var initMessages sync.Map

	// Use a channel to signal completion
	done := make(chan struct{})

	// Each node in old committee starts dACSS
	// That means that for the single share each node has,
	// ceil(nrShare/(nrOldNodes-2*recThreshold)) = 1 random values are sampled
	// and shared to both old & new committee
	pssIdInt := 0 // PssId should be the same for all the init messages.

	var wg sync.WaitGroup
	for index, n := range nodesOld {
		wg.Add(1)
		go func(index int, node *testutils.IntegrationTestNode) {
			defer wg.Done()
			ephemeralKeypair := common.GenerateKeyPair(curves.K256())

			privKeyShare := make([]common.PrivKeyShare, B)

			for i := 0; i < B; i++ {
				share := sharing.ShamirShare{Id: shares[i][index].Id, Value: shares[i][index].Value}

				privKeyShare[i] = common.PrivKeyShare{
					UserIdOwner: "DummyUserId",
					Share:       share,
				}
			}

			initMsg := getTestInitMsg(node, *big.NewInt(int64(pssIdInt)), B, &privKeyShare, ephemeralKeypair, TestSetUp.NewCommitteeParams)

			pssMsgData, err := bijson.Marshal(initMsg)
			assert.Nil(t, err)

			InitPssMessage := common.PSSMessage{
				PSSRoundDetails: initMsg.PSSRoundDetails,
				Type:            initMsg.Kind,
				Data:            pssMsgData,
			}

			// initialize empty state against each acssRound
			for i := 0; i < nGenerations; i++ {
				acssRound := common.ACSSRoundDetails{
					PSSRoundDetails: initMsg.PSSRoundDetails,
					ACSSCount:       i,
				}

				// Initialize the empty state for all the old nodes
				for _, node := range nodesOld {
					node.State().AcssStore.UpdateAccsState(
						acssRound.ToACSSRoundID(),
						func(as *common.AccsState) {},
					)
				}

				// Initialize the empty state for all the new nodes
				for _, node := range nodesNew {
					node.State().AcssStore.UpdateAccsState(
						acssRound.ToACSSRoundID(),
						func(as *common.AccsState) {},
					)
				}

			}

			//store the init msg for testing
			initMessages.Store(node.Details().GetNodeDetailsID(), initMsg)

			node.ReceiveMessage(node.Details(), InitPssMessage)

			// Signal completion
			done <- struct{}{}

		}(index, n)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait until all signals have been received
	for range nodesOld {
		<-done
	}

	time.Sleep(15 * time.Second)

	// DACSS Checks

	// round details for oldNode0 being the dealer
	//since only one random secret is shared, the ACSSCount = 0
	retrievedMsg, found := initMessages.Load(nodesOld[0].Details().GetNodeDetailsID())
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

	// getting the random secret shared

	for i := 0; i < nGenerations; i++ {
		acssRound := common.ACSSRoundDetails{
			PSSRoundDetails: message.PSSRoundDetails,
			ACSSCount:       i,
		}

		// getting the random secret shared by nodesOld[0]
		state, _, err := nodesOld[0].State().AcssStore.Get(acssRound.ToACSSRoundID())
		assert.Nil(t, err)
		randomSecretShared := state.RandomSecretShared[acssRound.ToACSSRoundID()]

		//storing shares received from the oldnode0 for old committee
		var OldCommitteReceivedShareFromNodeOld0 []*sharing.ShamirShare

		for _, n := range nodesOld {

			state, _, err := n.State().AcssStore.Get(acssRound.ToACSSRoundID())
			assert.Nil(t, err)

			// Check that the valid share output has been set.
			assert.True(t, state.ValidShareOutput)

			rbcState := state.RBCState.Phase
			assert.Equal(t, rbcState, common.Ended)
			share := state.ReceivedShare
			OldCommitteReceivedShareFromNodeOld0 = append(OldCommitteReceivedShareFromNodeOld0, (*sharing.ShamirShare)(share))

		}

		//For Old Committee

		//reconstructing the random secret
		shamir, err := sharing.NewShamir(testutils.DefaultK_old, testutils.DefaultN_old, curves.K256())
		assert.Nil(t, err)

		reconstructedSecret, err := shamir.Combine(OldCommitteReceivedShareFromNodeOld0...)
		assert.Nil(t, err)

		assert.Equal(t, reconstructedSecret, *randomSecretShared)

		//For New Committee

		//storing shares received from the oldnode0 for New committee
		var NewCommitteReceivedShareFromNodeOld0 []*sharing.ShamirShare

		for _, n := range nodesNew {

			state, _, err := n.State().AcssStore.Get(acssRound.ToACSSRoundID())
			assert.Nil(t, err)
			share := state.ReceivedShare
			NewCommitteReceivedShareFromNodeOld0 = append(NewCommitteReceivedShareFromNodeOld0, (*sharing.ShamirShare)(share))

		}

		//reconstructing the random secret
		shamir, err = sharing.NewShamir(testutils.DefaultK_new, testutils.DefaultN_new, curves.K256())
		assert.Nil(t, err)

		reconstructedSecret, err = shamir.Combine(NewCommitteReceivedShareFromNodeOld0...)
		assert.Nil(t, err)

		assert.Equal(t, reconstructedSecret, *randomSecretShared)
	}

	// --------------------------------------
	// TODO: Add MBVA Checks here(if needed)
	// --------------------------------------

	// Batch Reconstruction handler Checks
	// TODO: can be checked once MBVA is added

	// Check step 1: Each HimHandler invocation should send 1 preprocess message
	// in total we expect n PreProcessMessages
	sentMsgs := transport.GetSentMessages()
	preprocessRecMessages := make([]common.PSSMessage, 0)

	for _, msg := range sentMsgs {
		if msg.Type == old_committee.PreprocessBatchRecMessageType {
			preprocessRecMessages = append(preprocessRecMessages, msg)
		}
	}
	assert.Equal(t, nOld, len(preprocessRecMessages))

	// Check step 2: Each PreprocessBatchRecMessage should send
	// ceil(B/(n-2t)) InitRecHandlerMessages
	// 34
	nrBatches := math.Ceil(float64(B) / float64(nOld-2*tOld))

	receiveMessageMsgs := transport.GetReceivedMessages()
	assert.True(t, len(receiveMessageMsgs) > 0)
	initRecMessages := make([]common.PSSMessage, 0)

	for _, msg := range receiveMessageMsgs {
		if msg.Type == old_committee.InitRecHandlerType {
			initRecMessages = append(initRecMessages, msg)
		}
	}
	nrInitRecMessages := len(initRecMessages)
	// ceil(B/(n-2t)) * n_old
	// 34*7 = 238
	assert.Equal(t, int(nrBatches)*nOld, nrInitRecMessages)

	// Check step 3: From each InitRecHandlerMessage we expect
	// n_old PrivateRecHandlerMessages to be sent
	privateRecMessages := make([]common.PSSMessage, 0)

	for _, msg := range sentMsgs {
		if msg.Type == old_committee.PrivateRecMessageType {
			privateRecMessages = append(privateRecMessages, msg)
		}
	}
	// ceil(B/(n-2t)) * n_old InitRecHandlerMessage were sent
	// n_old messages are sent per InitRecHandlerMessage
	// 238 * 7 = 1666
	assert.Equal(t, nrInitRecMessages*nOld, len(privateRecMessages))

	// Filter broadcasted messages on the ones that are of type PublicRecMsg
	publicRecMessages := make([]common.PSSMessage, 0)
	broadcastedMsgs := transport.GetBroadcastedMessages()

	for _, msg := range broadcastedMsgs {
		if msg.Type == old_committee.PublicRecMessageType {
			publicRecMessages = append(publicRecMessages, msg)
		}
	}

	// There are ceil(B/(n-2t)) batch rounds
	// There are n_old nodes
	// Each old node should broadcast 1 PublicRecMessage per batch round
	// 34*7 = 238
	assert.Equal(t, int(nrBatches)*nOld, len(publicRecMessages))

	// Final check: the broadcasted LocalComputationMessages
	localComputationMessages := make([]common.PSSMessage, 0)
	for _, msg := range broadcastedMsgs {
		if msg.Type == new_committee.LocalComputationMessageType {
			localComputationMessages = append(localComputationMessages, msg)
		}
	}
	// Each node, per batch, broadcasts 1 LocalComputationMessage
	// n_old nodes
	// ceil(B/(n-2t)) nr of batches
	assert.Equal(t, int(nrBatches)*nOld, len(localComputationMessages))

	// // *-------------- Local Computation check ------------------*

	// //TODO: verify the below test code
	// type ScalarMap map[int]curves.Scalar

	// NewShare := make([]ScalarMap, 0)

	// for _, node := range nodesNew {
	// 	nShare := node.GetRefreshedShares()
	// 	NewShare = append(NewShare, nShare)
	// }

	// shamir, err := sharing.NewShamir(testutils.DefaultK_new, testutils.DefaultN_new, curves.K256())
	// assert.Nil(t, err)

	// NewShareSingle := make([]*sharing.ShamirShare, 0)
	// pssIndex := common.GetIndexFromPSSID(common.NewPssID(*big.NewInt(int64(pssIdInt))))

	// for i := range nodesNew {
	// 	keyIndex := (pssIndex * B) + 1
	// 	share := NewShare[i][keyIndex]
	// 	shamirShare := sharing.ShamirShare{
	// 		Id:    uint32(i + 1),
	// 		Value: share.Bytes(),
	// 	}
	// 	NewShareSingle = append(NewShareSingle, &shamirShare)
	// }
	// reconstructedSecretNew, err := shamir.Combine(NewShareSingle...)
	// assert.Nil(t, err)
	// assert.Equal(t, reconstructedSecretNew, secret[1])

}

// returns DACSS initMsg
func getTestInitMsg(testDealer *testutils.IntegrationTestNode, pssRoundIndex big.Int, batchSize int, share *[]common.PrivKeyShare, ephemeralKeypair common.KeyPair, newCommitteeParams common.CommitteeParams) *dacss.InitMessage {
	roundDetails := common.PSSRoundDetails{
		PssID:     common.NewPssID(pssRoundIndex),
		Dealer:    testDealer.Details(),
		BatchSize: batchSize,
	}

	msg := &dacss.InitMessage{
		PSSRoundDetails:    roundDetails,
		OldShares:          *share,
		EphemeralSecretKey: ephemeralKeypair.PrivateKey.Bytes(),
		EphemeralPublicKey: ephemeralKeypair.PublicKey.ToAffineCompressed(),
		Kind:               dacss.InitMessageType,
		CurveName:          &common.SECP256K1,
		NewCommitteeParams: newCommitteeParams,
	}
	return msg
}

func GenerateSecretAndGetCommitmentAndShares(B, nOld, kOld int) ([]curves.Scalar, []*sharing.FeldmanVerifier, [][]*sharing.ShamirShare, error) {

	// generate n secrets
	testSecret := make([]curves.Scalar, B)

	for i := 0; i < B; i++ {
		testSecret[i] = sharing.GenerateSecret(curves.K256())
	}

	var verifier []*sharing.FeldmanVerifier
	var shamirShare [][]*sharing.ShamirShare

	for i := 0; i < B; i++ {

		Fverifier, shares, err := sharing.GenerateCommitmentAndShares(testSecret[i], uint32(kOld), uint32(nOld), testutils.TestCurve())

		if err != nil {
			return nil, nil, nil, err
		}

		verifier = append(verifier, (*sharing.FeldmanVerifier)(Fverifier))
		shamirShare = append(shamirShare, shares)

	}

	return testSecret, verifier, shamirShare, nil

}
