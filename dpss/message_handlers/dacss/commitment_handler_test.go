package dacss

import (
	"crypto/rand"
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

/*
Function: Process

Test case: the receiver node receives t + 1 commitment messages that match with
the commitment in its internal state.

Expectations:
  - There is a commitment message whose counter is greater or equal than t + 1
  - The internal commitment hash and the hash of the commitment with count greater
    or equal than t + 1 should be equal.
*/
func TestCommitmentMsgHappyPath(test *testing.T) {
	commitmentMsg, receiverNode, senderGroup, _, err := getCommitmentMessageAndNodesSetup()
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
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while retrieving the state of the node",
			},
		).Error("DACSSCommitmentHandler: TestHappyPath")
		test.Error("Error while retrieving the state of the node")
	}
	if !found {
		log.WithFields(
			log.Fields{
				"Found":   found,
				"Message": "The state was not found",
			},
		).Error("DACSSCommitmentHandler: TestHappyPath")
		test.Error("State not found")
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

	assert.Equal(
		test,
		stateReceiver.OwnCommitmentsHash,
		hashCommitment,
	)
}

/*
Function: Process

Testcase: repeated send of the same commitment. The receiver node receives t + 1
messages from the same sender node with the same commitment.

Expectations:
  - There should not be a message with a counter higher than t + 1.
  - The message received should not count repetitions. Therefore, the count of the
    hash of the received message should be 1.
*/
func TestCommitmentMsgRepeatedMessages(test *testing.T) {
	commitmentMsg, receiverNode, senderGroup, _, err := getCommitmentMessageAndNodesSetup()
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
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while retrieving the state of the node",
			},
		).Error("DACSSCommitmentHandler: TestHappyPath")
		test.Error("Error while retrieving the state of the node")
	}
	if !found {
		log.WithFields(
			log.Fields{
				"Found":   found,
				"Message": "The state was not found",
			},
		).Error("DACSSCommitmentHandler: TestHappyPath")
		test.Error("State not found")
	}

	_, foundCommitment := stateReceiver.FindThresholdCommitment(t + 1)
	assert.False(test, foundCommitment)

	commitmentHashHex := hex.EncodeToString(commitmentMsg.CommitmentSecretHash)
	countMsg := stateReceiver.CommitmentCount[commitmentHashHex]
	assert.Equal(test, 1, countMsg)
}

/*
Function: Process

Testcase: The user receives t + 1 messages with a commitment that does not match the
hash of the commitments stored in its internal state. The messages are sent by
different nodes.

Expectations:
  - The count of the fake hash should increase to t + 1
  - The share obtained at the end of the protocol is not valid because the commitments
    do not match with the commitments stored in the internal state.
*/
func TestCommitmentModifiedCommitment(test *testing.T) {
	commitmentMsg, receiverNode, senderGroup, commitments, err := getCommitmentMessageAndNodesSetup()
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while creating the commitment message and the nodes",
			},
		).Error("TestHappyPath")
		test.Error("Error creating the commitment message and nodes.")
	}

	// Change the commitment of the secret on purpose.

	// Replaces the commitment in a random position for another different point
	// in the curve. Then, change the commitments of the created message by this faked
	// commitment.
	newPoint := curves.K256().Point.Random(rand.Reader)
	for newPoint.Equal(commitments.Commitments[0]) {
		newPoint = curves.K256().Point.Random(rand.Reader)
	}
	commitments.Commitments[0] = newPoint
	hashFakeCommitments := common.HashByte(commitments.Commitments[0].ToAffineCompressed())
	commitmentMsg.CommitmentSecretHash = hashFakeCommitments

	// Send the fake message.
	for _, senderNode := range senderGroup {
		commitmentMsg.Process(senderNode.Details(), receiverNode)
	}

	stateReceiver, found, err := receiverNode.State().AcssStore.Get(
		commitmentMsg.ACSSRoundDetails.ToACSSRoundID(),
	)
	assert.Nil(test, err)
	assert.True(test, found)

	_, _, t := receiverNode.Params()
	assert.Equal(
		test,
		t+1,
		stateReceiver.CommitmentCount[hex.EncodeToString(hashFakeCommitments)],
	)
	assert.False(
		test,
		stateReceiver.ValidShareOutput,
	)
}

func getCommitmentMessageAndNodesSetup() (DacssCommitmentMessage, *testutils.PssTestNode, []*testutils.PssTestNode, *sharing.FeldmanVerifier, error) {
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
		return DacssCommitmentMessage{}, nil, []*testutils.PssTestNode{}, nil, err
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
		ACSSRoundDetails:     acssRoundDetails,
		CommitmentSecretHash: hashCommitments,
		Kind:                 DacssCommitmentMessageType,
		CurveName:            common.CurveName(curves.K256().Name),
	}

	return commitmentMsg, receiverNode, senderGroup, commitments, nil
}
