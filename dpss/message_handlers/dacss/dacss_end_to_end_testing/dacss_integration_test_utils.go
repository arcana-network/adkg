package dacss

import (
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	log "github.com/sirupsen/logrus"
)

type DacssMessageProcessor struct {
}

func NewDacssTestNode(index int, keypair common.KeyPair, transport *testutils.IntegrationMockTransport, isFaulty, isNewCommittee bool) *testutils.IntegrationTestNode {
	return testutils.NewIntegrationTestNode(index, keypair, transport, isFaulty, isNewCommittee, DacssMessageProcessor{})
}

func (processor DacssMessageProcessor) ProcessMessages(sender common.NodeDetails, PssMessage common.PSSMessage, node *testutils.IntegrationTestNode) {
	if !(strings.HasPrefix(PssMessage.Type, "dacss")) {
		return
	}

	switch PssMessage.Type {
	case dacss.InitMessageType:
		testutils.ProcessDACSSMessage[dacss.InitMessage](PssMessage.Data, sender, node, dacss.InitMessageType)
	case dacss.DacssEchoMessageType:
		testutils.ProcessDACSSMessage[dacss.DacssEchoMessage](PssMessage.Data, sender, node, dacss.DacssEchoMessageType)
	case dacss.ShareMessageType:
		testutils.ProcessDACSSMessage[dacss.DualCommitteeACSSShareMessage](PssMessage.Data, sender, node, dacss.ShareMessageType)
	case dacss.AcssProposeMessageType:
		testutils.ProcessDACSSMessage[*dacss.AcssProposeMessage](PssMessage.Data, sender, node, dacss.AcssProposeMessageType)
	case dacss.AcssReadyMessageType:
		testutils.ProcessDACSSMessage[*dacss.DacssReadyMessage](PssMessage.Data, sender, node, dacss.AcssReadyMessageType)
	case dacss.ImplicateExecuteMessageType:
		testutils.ProcessDACSSMessage[*dacss.ImplicateExecuteMessage](PssMessage.Data, sender, node, dacss.ImplicateExecuteMessageType)
	case dacss.ImplicateReceiveMessageType:
		testutils.ProcessDACSSMessage[*dacss.ImplicateReceiveMessage](PssMessage.Data, sender, node, dacss.ImplicateReceiveMessageType)
	case dacss.ShareRecoveryMessageType:
		testutils.ProcessDACSSMessage[*dacss.ShareRecoveryMessage](PssMessage.Data, sender, node, dacss.ShareRecoveryMessageType)
	case dacss.ReceiveShareRecoveryMessageType:
		testutils.ProcessDACSSMessage[*dacss.ReceiveShareRecoveryMessage](PssMessage.Data, sender, node, dacss.ReceiveShareRecoveryMessageType)
	case dacss.DacssOutputMessageType:
		testutils.ProcessDACSSMessage[*dacss.DacssOutputMessage](PssMessage.Data, sender, node, dacss.DacssOutputMessageType)
	case dacss.DacssCommitmentMessageType:
		testutils.ProcessDACSSMessage[*dacss.DacssCommitmentMessage](PssMessage.Data, sender, node, dacss.DacssCommitmentMessageType)

	default:
		log.Infof("No handler found. MsgType=%s", PssMessage.Type)
	}

}
