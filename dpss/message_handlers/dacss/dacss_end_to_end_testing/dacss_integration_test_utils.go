package dacss

import (
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	log "github.com/sirupsen/logrus"
)

// Helper for DACSS Integration testing
type DacssMessageProcessor struct {
}

// Returns an IntegrationTestNode that has the DACSS message processor
func NewDacssTestNode(index int, keypair common.KeyPair, transport *testutils.IntegrationMockTransport, isFaulty, isNewCommittee bool) *testutils.IntegrationTestNode {
	return testutils.NewIntegrationTestNode(index, keypair, transport, isFaulty, isNewCommittee, DacssMessageProcessor{})
}

// Only messages with prefix dacss are passed on for this processor
func (processor DacssMessageProcessor) ProcessMessages(sender common.NodeDetails, PssMessage common.PSSMessage, node *testutils.IntegrationTestNode) {
	if !(strings.HasPrefix(PssMessage.Type, "dacss")) {
		return
	}

	switch PssMessage.Type {
	case dacss.InitMessageType:
		testutils.ProcessMessageForType[dacss.InitMessage](PssMessage.Data, sender, node, dacss.InitMessageType)
	case dacss.DacssEchoMessageType:
		testutils.ProcessMessageForType[dacss.DacssEchoMessage](PssMessage.Data, sender, node, dacss.DacssEchoMessageType)
	case dacss.ShareMessageType:
		testutils.ProcessMessageForType[dacss.DualCommitteeACSSShareMessage](PssMessage.Data, sender, node, dacss.ShareMessageType)
	case dacss.AcssProposeMessageType:
		testutils.ProcessMessageForType[*dacss.AcssProposeMessage](PssMessage.Data, sender, node, dacss.AcssProposeMessageType)
	case dacss.AcssReadyMessageType:
		testutils.ProcessMessageForType[*dacss.DacssReadyMessage](PssMessage.Data, sender, node, dacss.AcssReadyMessageType)
	case dacss.ImplicateExecuteMessageType:
		testutils.ProcessMessageForType[*dacss.ImplicateExecuteMessage](PssMessage.Data, sender, node, dacss.ImplicateExecuteMessageType)
	case dacss.ImplicateReceiveMessageType:
		testutils.ProcessMessageForType[*dacss.ImplicateReceiveMessage](PssMessage.Data, sender, node, dacss.ImplicateReceiveMessageType)
	case dacss.ShareRecoveryMessageType:
		testutils.ProcessMessageForType[*dacss.ShareRecoveryMessage](PssMessage.Data, sender, node, dacss.ShareRecoveryMessageType)
	case dacss.ReceiveShareRecoveryMessageType:
		testutils.ProcessMessageForType[*dacss.ReceiveShareRecoveryMessage](PssMessage.Data, sender, node, dacss.ReceiveShareRecoveryMessageType)
	case dacss.DacssOutputMessageType:
		testutils.ProcessMessageForType[*dacss.DacssOutputMessage](PssMessage.Data, sender, node, dacss.DacssOutputMessageType)
	case dacss.DacssCommitmentMessageType:
		testutils.ProcessMessageForType[*dacss.DacssCommitmentMessage](PssMessage.Data, sender, node, dacss.DacssCommitmentMessageType)

	default:
		log.Infof("No handler found. MsgType=%s", PssMessage.Type)
	}

}
