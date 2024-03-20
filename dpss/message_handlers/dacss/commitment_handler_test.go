package dacss

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestHappyPath(test *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	oldNodes := defaultSetup.GetAllOldNodesFromTestSetup()
	newNodes := defaultSetup.GetAllNewNodesFromTestSetup()

	// The dealer node will be the node at possition zero in the old committee.
	dealerNode := oldNodes[0]
	n, k, t := dealerNode.Params()

	// The sender committee are all old nodes that will send a message to the
	// new committee
	senderGroup := oldNodes[1 : t+1+1]
	assert.Equal(test, t+1, len(senderGroup))

	id := big.NewInt(1)
	roundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: dealerNode.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: roundDetails,
		ACSSCount:       1,
	}

	// The receiver node will be the node at possition zero in the new committe.
	receiverNode := newNodes[0]
	secret := sharing.GenerateSecret(curves.K256())
	commitments, _, err := sharing.GenerateCommitmentAndShares(
		secret,
		uint32(k),
		uint32(n),
		curves.K256(),
	)
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error constructing the commitments",
			},
		).Error("DACSSCommitmentHandler: TestHappyPath")
		test.Error("Error while constructing the commitments")
	}

	concatCommitments := sharing.ConcatenateCommitments(commitments)
	hashCommitments := common.HashByte(concatCommitments)

	receiverNode.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.OwnCommitmentsHash = hex.EncodeToString(hashCommitments)
		},
	)

	commitmentMsg := DacssCommitmentMessage{
		ACSSRoundDetails: acssRoundDetails,
		CommitmentsHash:  hashCommitments,
		Kind:             DacssCommitmentMessageType,
		CurveName:        common.CurveName(curves.K256().Name),
	}

	for _, senderNode := range senderGroup {
		commitmentMsg.Process(senderNode.Details(), receiverNode)
	}

	stateReceiver, found, err := receiverNode.State().AcssStore.Get(
		acssRoundDetails.ToACSSRoundID(),
	)
	if !found {
		log.WithFields(
			log.Fields{
				"Found":   found,
				"Message": "The state was not found",
			},
		).Error("DACSSCommitmentHandler: TestHappyPath")
		test.Error("State not found")
	}
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while retrieving the state of the node",
			},
		).Error("DACSSCommitmentHandler: TestHappyPath")
		test.Error("Error while retrieving the state of the node")
	}

	hashCommitment, found := stateReceiver.FindThresholdCommitment(t + 1)
	if !found {
		log.WithFields(
			log.Fields{
				"Threshold": t + 1,
				"Message":   "There is no record with the given threshold",
			},
		).Error("DACSSCommitmentHandler: TestHappyPath")
		test.Error("There is no record with the given threshold")
	} else {
		assert.Equal(test, t+1, stateReceiver.CommitmentCount[hashCommitment])
	}
}
