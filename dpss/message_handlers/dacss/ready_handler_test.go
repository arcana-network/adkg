package dacss

import (
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
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

	/*
		Sets up three types of nodes:
		-	A receiver node that will be the one that receives the ECHO/READY
			messages.
		-	A group of nodes, which are nodes that will send the ECHO
			and READY messages to the receiver node.
		-	A dealer which will be the node that calls the protocol RBC(M) where
			M is the message to be broadcasted. The dealer is a node selected
			from the group of sender nodes.

	*/
	const oldCommittee = true
	defaultSetup := testutils.DefaultTestSetup()
	receiverNode, senderGroup := defaultSetup.GetDealerAnd2kPlusOneNodes(oldCommittee)
	transport := receiverNode.Transport

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
	shardReceiver, hashMsg, err := createShardAndHash(
		dealerNode,
		receiverNode,
		ephemeralKeypairDealer,
	)
	if err != nil {
		log.WithField("error", err).Error("unable to crate RS Shares ")
	}

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

	_, t, _ := receiverNode.Params()

	for i := range t + 1 {
		echoMsg := DacssEchoMessage{
			ACSSRoundDetails: acssRoundDetails,
			Kind:             DacssEchoMessageType,
			CurveName:        common.CurveName(curves.K256().Name),
			Share:            shardReceiver,
			Hash:             hashMsg,
			NewCommittee:     receiverNode.IsOldNode(),
		}
		echoMsg.Process(senderGroup[i].Details(), receiverNode)

		readyMsg := DacssReadyMessage{
			AcssRoundDetails: acssRoundDetails,
			Kind:             DacssEchoMessageType,
			CurveName:        common.CurveName(curves.K256().Name),
			Share:            shardReceiver,
			Hash:             hashMsg,
		}
		readyMsg.Process(senderGroup[i].Details(), receiverNode)
	}

	broadcastedMsg := transport.BroadcastedMessages

	// There should be just one broadcasted message.
	assert.Equal(test, 1, len(broadcastedMsg))

	// The ECHO and READY counts should be t + 1.
	rbcState, found, err := receiverNode.State().AcssStore.Get(
		acssRoundDetails.ToACSSRoundID(),
	)
	if !found {
		test.Errorf("RBC state not found.")
	}
	if err != nil {
		test.Errorf("Error retrieving the RBC state: %v", err)
	}
	assert.Equal(test, t+1, rbcState.RBCState.CountEcho())
	assert.Equal(test, t+1, rbcState.RBCState.CountReady())

}
