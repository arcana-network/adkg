package dacss

import (
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

/*
Function: Process

Testcase: happy path. Test that if the node have received t + 1 READY messages
and t + 1 ECHO messages that match the correct own shard, then the receiver node
broadcasts a READY message.

Expect:
  - The receiver node sends a broadcast at the end of the test.
  - The ECHO counter is t + 1.
  - The READY counter is t + 1.
*/
func TestProcessReadyMessage(test *testing.T) {
	receiverNode, senderGroup, acssRoundDetails, err := setupDealerAndGroup()
	if err != nil {
		test.Errorf("Error while setting up the nodes: %v", err)
	}

	stateReceiver, found, err := receiverNode.State().AcssStore.Get(
		acssRoundDetails.ToACSSRoundID(),
	)
	if !found {
		test.Errorf("State not found")
	}
	if err != nil {
		test.Errorf("Error retrieving the state of the node: %v", err)
	}

	_, _, t := receiverNode.Params()

	echoMsg := DacssEchoMessage{
		ACSSRoundDetails: *acssRoundDetails,
		Kind:             DacssEchoMessageType,
		CurveName:        common.CurveName(curves.K256().Name),
		Share:            stateReceiver.RBCState.OwnReedSolomonShard,
		Hash:             stateReceiver.AcssDataHash,
		NewCommittee:     receiverNode.IsNewNode(),
	}
	readyMsg := DacssReadyMessage{
		AcssRoundDetails: *acssRoundDetails,
		Kind:             DacssEchoMessageType,
		CurveName:        common.CurveName(curves.K256().Name),
		Share:            stateReceiver.RBCState.OwnReedSolomonShard,
		Hash:             stateReceiver.AcssDataHash,
	}
	for i := range t + 1 {
		echoMsg.Process(senderGroup[i].Details(), receiverNode)
		readyMsg.Process(senderGroup[i].Details(), receiverNode)
	}

	broadcastedMsg := receiverNode.Transport.BroadcastedMessages

	// There should be just one broadcasted message.
	assert.Equal(test, 1, len(broadcastedMsg))

	// The ECHO and READY counts should be t + 1.
	state, found, err := receiverNode.State().AcssStore.Get(
		acssRoundDetails.ToACSSRoundID(),
	)
	if !found {
		test.Errorf("RBC state not found.")
	}
	if err != nil {
		test.Errorf("Error retrieving the RBC state: %v", err)
	}
	msgInfo := state.RBCState.GetEchoStore(
		echoMsg.Fingerprint(),
		stateReceiver.AcssDataHash,
		stateReceiver.RBCState.OwnReedSolomonShard,
	)
	assert.Equal(test, t+1, msgInfo.Count)
	assert.Equal(test, t+1, state.RBCState.CountReady())
}

/*
Function: Process

Testcase: semi-happy path. Test that if the node have received t + 1 READY
messages and after certain time it receives t + 1 ECHO messages that match the
correct own shard, then the receiver node broadcasts a READY message.

Expect:
  - The receiver node sends a broadcast at the end of the test.
  - The ECHO counter is t + 1 at the end of the test.
  - The READY counter is t + 1 at the end of the test.
  - After receiving t + 1 READY messages and no ECHO messages, the node does
    not broadcast any message.
  - After receiving t + 1 ECHO messages, the node does broadcast one READY
    message.
*/
func TestLateEchoAfterReady(test *testing.T) {
	receiverNode, senderGroup, acssRoundDetails, err := setupDealerAndGroup()
	if err != nil {
		test.Errorf("Error while setting up the nodes: %v", err)
	}

	stateReceiver, found, err := receiverNode.State().AcssStore.Get(
		acssRoundDetails.ToACSSRoundID(),
	)
	if !found {
		test.Errorf("State not found")
	}
	if err != nil {
		test.Errorf("Error retrieving the state of the node: %v", err)
	}

	_, _, t := receiverNode.Params()

	// Simulates the reception of t + 1 READY messages
	for i := range t + 1 {
		readyMsg := DacssReadyMessage{
			AcssRoundDetails: *acssRoundDetails,
			Kind:             DacssEchoMessageType,
			CurveName:        common.CurveName(curves.K256().Name),
			Share:            stateReceiver.RBCState.OwnReedSolomonShard,
			Hash:             stateReceiver.AcssDataHash,
		}
		readyMsg.Process(senderGroup[i].Details(), receiverNode)
	}

	// Test that no broadcast has been sent because there are t + 1 ECHO
	// messages left.
	assert.Equal(test, 0, len(receiverNode.Transport.BroadcastedMessages))
	assert.Equal(test, t+1, stateReceiver.RBCState.CountReady())

	// Simulates the reception of t + 1 ECHO messages
	echoMsg := DacssEchoMessage{
		ACSSRoundDetails: *acssRoundDetails,
		Kind:             DacssEchoMessageType,
		CurveName:        common.CurveName(curves.K256().Name),
		Share:            stateReceiver.RBCState.OwnReedSolomonShard,
		Hash:             stateReceiver.AcssDataHash,
		NewCommittee:     receiverNode.IsNewNode(),
	}
	for i := range t + 1 {
		echoMsg.Process(senderGroup[i].Details(), receiverNode)
	}
	msgInfo := stateReceiver.RBCState.GetEchoStore(
		echoMsg.Fingerprint(),
		echoMsg.Hash,
		echoMsg.Share,
	)
	assert.Equal(test, 1, len(receiverNode.Transport.BroadcastedMessages))
	assert.Equal(test, t+1, msgInfo.Count)
	assert.Equal(test, t+1, stateReceiver.RBCState.CountReady())
}

/*
Function: Process

Testcase: happy path to the OUTPUT handler. This test evaluates that if the node
sends the OUTPUT message after receiving the correct ammount of shards.

Expect:
  - The OUTPUT message is sent after receiving 2t + 1 correct shards.
  - The reconstructed message is the same as the dealt message.
*/
func TestGoingToOutputHandler(test *testing.T) {
	receiverNode, senderGroup, acssRoundDetails, err := setupDealerAndGroup()
	if err != nil {
		test.Errorf("Error while setting up the nodes: %v", err)
	}

	stateReceiver, found, err := receiverNode.State().AcssStore.Get(
		acssRoundDetails.ToACSSRoundID(),
	)
	if !found {
		test.Errorf("State not found")
	}
	if err != nil {
		test.Errorf("Error retrieving the state of the node: %v", err)
	}

	// Simulates the reception of 2t + 1 READY messages
	for _, senderNode := range senderGroup {
		stateSender, found, err := senderNode.State().AcssStore.Get(
			acssRoundDetails.ToACSSRoundID(),
		)
		if !found {
			test.Errorf("State not found")
		}
		if err != nil {
			test.Errorf("Error retrieving the state of the node: %v", err)
		}

		readyMsg := DacssReadyMessage{
			AcssRoundDetails: *acssRoundDetails,
			Kind:             DacssEchoMessageType,
			CurveName:        common.CurveName(curves.K256().Name),
			Share:            stateSender.RBCState.OwnReedSolomonShard,
			Hash:             stateReceiver.AcssDataHash,
		}
		readyMsg.Process(senderNode.Details(), receiverNode)
	}

	_, _, t := receiverNode.Params()
	hashOutputRbc := common.HashByte(stateReceiver.RBCState.ReceivedMessage)
	assert.Equal(test, 2*t+1, len(stateReceiver.RBCState.ReadyMsgShards))
	assert.Equal(test, 2*t+1, stateReceiver.RBCState.CountReady())
	assert.Equal(test, stateReceiver.AcssDataHash, hashOutputRbc)
	assert.Equal(test, stateReceiver.RBCState.Phase, common.Ended)
}

/*
Function: Process

Testcase: tests the situation where the same sender node sends a READY message
multiple times.

Expect:
  - The ammount of READY shards stored is just one.
  - The counter of received READY messages is one.
*/
func TestRepeatedReadyMessages(test *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode, senderNode := defaultSetup.GetTwoOldNodesFromTestSetup()

	// Creates the details for the ACSS protocol.
	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: senderNode.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: pssRoundDetails,
		ACSSCount:       1,
	}

	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

	shards, hashMsg, err := testutils.CreateShardAndHash(
		senderNode,
		ephemeralKeypairDealer,
	)
	if err != nil {
		test.Errorf("Error computing the shards of the message: %v", err)
	}

	senderNode.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.AcssDataHash = hashMsg
			state.RBCState.OwnReedSolomonShard = shards[senderNode.Details().Index-1]
		},
	)
	receiverNode.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.AcssDataHash = hashMsg
			state.RBCState.OwnReedSolomonShard = shards[receiverNode.Details().Index-1]
		},
	)

	// Simulates the reception of 2t + 1 READY messages by the same sender.
	stateReceiver, found, err := receiverNode.State().AcssStore.Get(
		acssRoundDetails.ToACSSRoundID(),
	)
	if !found {
		test.Errorf("State not found")
	}
	if err != nil {
		test.Errorf("Error retrieving the state of the node: %v", err)
	}
	stateSender, found, err := senderNode.State().AcssStore.Get(
		acssRoundDetails.ToACSSRoundID(),
	)
	if !found {
		test.Errorf("State not found")
	}
	if err != nil {
		test.Errorf("Error retrieving the state of the node: %v", err)
	}
	_, _, t := receiverNode.Params()
	for range 2*t + 1 {
		readyMsg := DacssReadyMessage{
			AcssRoundDetails: acssRoundDetails,
			Kind:             DacssEchoMessageType,
			CurveName:        common.CurveName(curves.K256().Name),
			Share:            stateSender.RBCState.OwnReedSolomonShard,
			Hash:             stateReceiver.AcssDataHash,
		}
		readyMsg.Process(senderNode.Details(), receiverNode)
	}

	assert.Equal(test, 1, len(stateReceiver.RBCState.ReadyMsgShards))
	assert.Equal(test, 1, stateReceiver.RBCState.CountReady())
}

/*
Sets up three types of nodes:
  - A receiver node that will be the one that receives the ECHO/READY
    messages.
  - A group of nodes, which are nodes that will send the ECHO
    and READY messages to the receiver node.
  - A dealer which will be the node that calls the protocol RBC(M) where
    M is the message to be broadcasted. The dealer is a node selected
    from the group of sender nodes.
*/
func setupDealerAndGroup() (
	*testutils.PssTestNode,
	[]*testutils.PssTestNode,
	*common.ACSSRoundDetails,
	error,
) {
	const oldCommittee = true
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode, senderGroup := defaultSetup.GetDealerAnd2kPlusOneNodes(oldCommittee)

	// Defines the dealer node and its ephemeral key pair
	dealerNode := senderGroup[0]
	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

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

	// Creates the shards and hash for a random secret
	shards, hashMsg, err := testutils.CreateShardAndHash(
		dealerNode,
		ephemeralKeypairDealer,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	// Sets the sender own shards
	for _, senderNode := range senderGroup {
		senderNode.State().AcssStore.Lock()
		senderNode.State().AcssStore.UpdateAccsState(
			acssRoundDetails.ToACSSRoundID(),
			func(state *common.AccsState) {
				state.RBCState.OwnReedSolomonShard = shards[senderNode.Details().Index-1]
				state.AcssDataHash = hashMsg
			},
		)
		senderNode.State().AcssStore.Unlock()
	}

	shardReceiver := shards[receiverNode.Details().Index-1]

	// Sets up the receivers own local share and hash of the message.
	receiverNode.State().AcssStore.Lock()
	receiverNode.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.AcssDataHash = hashMsg
			state.RBCState.OwnReedSolomonShard = shardReceiver
		},
	)
	receiverNode.State().AcssStore.Unlock()
	return receiverNode, senderGroup, &acssRoundDetails, nil
}
