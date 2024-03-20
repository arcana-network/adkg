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

func TestCommitmentMsgHappyPath(test *testing.T) {
	commitmentMsg, receiverNode, senderGroup, err := getCommitmentMessageAndNodesSetup()
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while creating the commitment message and the nodes",
			},
		).Error("TestHappyPath")
		test.Error("Error creating the commitment message and nodes.")
	}

	for _, senderNode := range senderGroup {
		commitmentMsg.Process(senderNode.Details(), receiverNode)
	}

	receiverNode.State().AcssStore.Lock()
	defer receiverNode.State().AcssStore.Unlock()

	stateReceiver, found, err := receiverNode.State().AcssStore.Get(
		commitmentMsg.ACSSRoundDetails.ToACSSRoundID(),
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

	_, _, t := receiverNode.Params()
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

func TestCommitmentMsgRepeatedMessages(test *testing.T) {
	commitmentMsg, receiverNode, senderGroup, err := getCommitmentMessageAndNodesSetup()
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while creating the commitment message and the nodes",
			},
		).Error("TestHappyPath")
		test.Error("Error creating the commitment message and nodes.")
	}

	// Take just one node from the sender group that will send the same message
	// multiple times.
	senderNode := senderGroup[0]
	_, _, t := receiverNode.Params()

	// Send the same message t+1 times
	for range t + 1 {
		commitmentMsg.Process(senderNode.Details(), receiverNode)
	}

	receiverNode.State().AcssStore.Lock()
	defer receiverNode.State().AcssStore.Unlock()
	stateReceiver, found, err := receiverNode.State().AcssStore.Get(
		commitmentMsg.ACSSRoundDetails.ToACSSRoundID(),
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

	_, foundCommitment := stateReceiver.FindThresholdCommitment(t + 1)
	assert.False(test, foundCommitment)

	commitmentHashHex := hex.EncodeToString(commitmentMsg.CommitmentsHash)
	countMsg := stateReceiver.CommitmentCount[commitmentHashHex]
	assert.Equal(test, 1, countMsg)
}

func getCommitmentMessageAndNodesSetup() (DacssCommitmentMessage, *testutils.PssTestNode, []*testutils.PssTestNode, error) {
	defaultSetup := testutils.DefaultTestSetup()
	oldNodes := defaultSetup.GetAllOldNodesFromTestSetup()
	newNodes := defaultSetup.GetAllNewNodesFromTestSetup()

	// The dealer node will be the node at possition zero in the old committee.
	dealerNode := oldNodes[0]
	n, k, t := dealerNode.Params()

	// The sender committee are all old nodes that will send a message to the
	// new committee
	senderGroup := oldNodes[1 : t+1+1]

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
		return DacssCommitmentMessage{}, nil, []*testutils.PssTestNode{}, err
	}

	concatCommitments := sharing.CompressCommitments(commitments)
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

	return commitmentMsg, receiverNode, senderGroup, nil
}
