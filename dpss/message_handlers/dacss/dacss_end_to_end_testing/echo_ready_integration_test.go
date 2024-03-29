package dacss

import (
	"math/big"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

func TestEchoReadyInteraction(test *testing.T) {
	log.SetLevel(log.DebugLevel)
	setup, transport := DefaultTestSetup()
	completeCommittee := setup.oldCommitteeNetwork

	log.Debugf(
		"n = %d, t = %d",
		setup.OldCommitteeParams.N,
		setup.OldCommitteeParams.T,
	)

	// The receiver node will be the node at position zero. The dealer in the
	// RBC protocol will be the node at position one.
	receiverNode := completeCommittee[0]
	dealerNode := completeCommittee[1]

	// Defines the group of nodes that will send the READY and ECHO messages.
	t := setup.OldCommitteeParams.T
	echoSenderGroup := completeCommittee[1 : t+1+1]
	readySenderGroup := completeCommittee[t+1+1 : 2*t+2+1]
	assert.Equal(test, t+1, len(echoSenderGroup))
	assert.Equal(test, t+1, len(readySenderGroup))

	// Creates the details for the ACSS protocol.
	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: dealerNode.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: pssRoundDetails,
		ACSSCount:       1,
	}

	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

	shards, hashMsg, err := testutils.CreateShardAndHash(
		dealerNode,
		ephemeralKeypairDealer,
	)
	if err != nil {
		test.Errorf("Error computing the shards of the message: %v", err)
	}

	for _, node := range completeCommittee {
		node.State().AcssStore.UpdateAccsState(
			acssRoundDetails.ToACSSRoundID(),
			func(state *common.AccsState) {
				state.RBCState.OwnReedSolomonShard = shards[node.Details().Index-1]
				state.AcssDataHash = hashMsg
			},
		)
	}
	shardReceiver := shards[receiverNode.Details().Index-1]

	// Sends t + 1 READY messages to the receiver node.
	for _, senderNode := range readySenderGroup {
		readyMsg, err := dacss.NewDacssReadyMessage(
			acssRoundDetails,
			shardReceiver,
			hashMsg,
			common.CurveName(curves.K256().Name),
		)
		if err != nil {
			test.Errorf("error creating the ECHO message: %v", err)
		}

		go senderNode.Send(receiverNode.Details(), *readyMsg)
	}
	time.Sleep(5 * time.Second)

	// Test that the received messages so far are all ECHO messages.
	stateReceiver, found, err := receiverNode.State().AcssStore.Get(
		acssRoundDetails.ToACSSRoundID(),
	)
	if !found {
		test.Errorf("State not found")
	}
	if err != nil {
		test.Errorf("Error retrieving the state of the node: %v", err)
	}

	assert.Equal(test, t+1, len(transport.GetSentMessages()))
	for _, message := range transport.GetSentMessages() {
		assert.Equal(test, dacss.AcssReadyMessageType, message.Type)
	}
	assert.Equal(test, t+1, stateReceiver.RBCState.CountReady())

	// Sends t + 1 ECHO messages to the receiver node.
	echoMsg, err := dacss.NewDacssEchoMessage(
		acssRoundDetails,
		shardReceiver,
		hashMsg,
		common.CurveName(curves.K256().Name),
		receiverNode.IsNewNode(),
	)
	if err != nil {
		test.Errorf("error creating the ECHO message: %v", err)
	}
	for _, senderNode := range echoSenderGroup {
		go senderNode.Send(receiverNode.Details(), *echoMsg)
	}
	time.Sleep(2 * time.Second)

	// Test that we have received 2t + 2 messages, and they are either READY
	// messages or ECHO messages.
	assert.Equal(test, 2*t+2, len(transport.GetSentMessages()))
	for _, message := range transport.GetSentMessages() {
		assert.Condition(
			test,
			func() (success bool) {
				success = (message.Type == dacss.DacssEchoMessageType || message.Type == dacss.AcssReadyMessageType)
				return
			},
		)
	}

	echoMsgTemp := dacss.DacssEchoMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             dacss.DacssEchoMessageType,
		CurveName:        common.CurveName(curves.K256().Name),
		Share:            shardReceiver,
		Hash:             hashMsg,
	}
	msgInfo := stateReceiver.RBCState.GetEchoStore(
		echoMsgTemp.Fingerprint(),
		hashMsg,
		shardReceiver,
	)
	assert.Equal(test, t+1, msgInfo.Count)
	assert.Equal(test, 1, len(transport.GetBroadcastedMessages()))
	assert.Equal(
		test,
		dacss.AcssReadyMessageType,
		transport.GetBroadcastedMessages()[0].Type,
	)
}
