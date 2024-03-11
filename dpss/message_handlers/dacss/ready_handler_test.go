package dacss

import (
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

// TODO: WIP
// Test Happy Path
func TestProcessReadyMessage(t *testing.T) {

	// defaultSetup := testutils.DefaultTestSetup()

	// OldNodes := defaultSetup.GetAllOldNodesFromTestSetup()

	// SingleOldNode := defaultSetup.GetSingleOldNodeFromTestSetup()

	// transport := SingleOldNode.Transport

	// // 2k+1 valid ready and eco messages
	// ReadyMessages, _, senderGroup := getTestValidEcoAndReadyMsg(SingleOldNode, defaultSetup)

	// for i, senderNode := range senderGroup {
	// 	ReadyMessages[i].Process(senderNode.Details(), senderNode)

	// }

	// broadcastedMsg := transport.BroadcastedMessages

	// assert.Equal(t, len(broadcastedMsg), 2*defaultSetup.OldCommitteeParams.K+1)
}

// TODO: check correctness of the msg
// creating dummy valid ReadyMsg
// build upon EcoHandlerTest file
func getTestValidEcoAndReadyMsg(SingleNode *testutils.PssTestNode, defaultSetup *testutils.TestSetup) ([]DacssReadyMessage, []DacssEchoMessage, []*testutils.PssTestNode) {

	receiverNode, senderGroup := defaultSetup.GetDealerAnd2kPlusOneNodes(true)

	// The dealer node will be the first node in the set of 2k + 1 nodes.
	dealerNode := senderGroup[0]

	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: SingleNode.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: pssRoundDetails,
		ACSSCount:       1,
	}

	receiverNode.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(as *common.AccsState) {},
	)

	//initialise state
	for i := range senderGroup {
		senderGroup[i].State().AcssStore.UpdateAccsState(
			acssRoundDetails.ToACSSRoundID(),
			func(as *common.AccsState) {},
		)
	}

	ephemeralKeypairDealer := common.GenerateKeyPair(curves.K256())

	shardReceiver, hashMsg, err := createShardAndHash(
		dealerNode,
		receiverNode,
		ephemeralKeypairDealer,
	)
	if err != nil {
		log.WithField("error", err).Error("unable to crate RS Shares ")
	}

	EcoMessages := make([]DacssEchoMessage, 0)
	ReadyMessages := make([]DacssReadyMessage, 0)

	//creates 2k+1 eco messages
	for range senderGroup {
		echoMsg, err := getTestEchoMsg(
			dealerNode,
			receiverNode,
			acssRoundDetails,
			shardReceiver,
			hashMsg,
		)
		if err != nil {
			log.WithField("error", err).Error("unable to crate EcoMsg")
		}

		ReadyMsg := DacssReadyMessage{
			AcssRoundDetails: acssRoundDetails,
			Kind:             AcssReadyMessageType,
			CurveName:        echoMsg.CurveName,
			Share:            echoMsg.Share,
			Hash:             echoMsg.Hash,
		}

		if err != nil {
			log.WithField("error", err).Error("unable to crate EcoMsg")
		}

		EcoMessages = append(EcoMessages, echoMsg)
		ReadyMessages = append(ReadyMessages, ReadyMsg)
	}
	return ReadyMessages, EcoMessages, senderGroup
}
