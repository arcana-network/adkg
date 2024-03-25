package DpssEndToEndTesting

import (
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/old_committee"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	log "github.com/sirupsen/logrus"
)

// Helper for DPSS end-to-end testing
type DpssEndToEndMessageProcessor struct {
}

// Returns an IntegrationTestNode that has the DpssEndToEndMessageProcessor
func NewTestNode(index int, keypair common.KeyPair, transport *testutils.IntegrationMockTransport, isFaulty, isNewCommittee bool) *testutils.IntegrationTestNode {
	return testutils.NewIntegrationTestNode(index, keypair, transport, isFaulty, isNewCommittee, DpssEndToEndMessageProcessor{})
}

// Exact definition of what messages are to be passed on for this processor
func (processor DpssEndToEndMessageProcessor) ProcessMessages(sender common.NodeDetails, PssMessage common.PSSMessage, node *testutils.IntegrationTestNode) {

	switch {

	case strings.HasPrefix(PssMessage.Type, "dacss"):
		processor.ProcessDacssMessages(sender, PssMessage, node)

	case strings.HasPrefix(PssMessage.Type, "dpss"):
		processor.ProcessDpssMessages(sender, PssMessage, node)

	// // PlaceHolder for MBVA
	// case strings.HasPrefix(PssMessage.Type, "aba"):
	// 	processor.ProcessMvbaMessages(sender, PssMessage, node)

	default:
		log.Infof("No handler found. MsgType=%s", PssMessage.Type)

	}

}

// message processor for DACSS
func (processor DpssEndToEndMessageProcessor) ProcessDacssMessages(sender common.NodeDetails, PssMessage common.PSSMessage, node *testutils.IntegrationTestNode) {

	switch PssMessage.Type {
	case dacss.InitMessageType:
		dpss.ProcessMessageForType[dacss.InitMessage](PssMessage.Data, sender, node, dacss.InitMessageType)
	case dacss.DacssEchoMessageType:
		dpss.ProcessMessageForType[dacss.DacssEchoMessage](PssMessage.Data, sender, node, dacss.DacssEchoMessageType)
	case dacss.ShareMessageType:
		dpss.ProcessMessageForType[dacss.DualCommitteeACSSShareMessage](PssMessage.Data, sender, node, dacss.ShareMessageType)
	case dacss.AcssProposeMessageType:
		dpss.ProcessMessageForType[*dacss.AcssProposeMessage](PssMessage.Data, sender, node, dacss.AcssProposeMessageType)
	case dacss.AcssReadyMessageType:
		dpss.ProcessMessageForType[*dacss.DacssReadyMessage](PssMessage.Data, sender, node, dacss.AcssReadyMessageType)
	case dacss.ImplicateExecuteMessageType:
		dpss.ProcessMessageForType[*dacss.ImplicateExecuteMessage](PssMessage.Data, sender, node, dacss.ImplicateExecuteMessageType)
	case dacss.ImplicateReceiveMessageType:
		dpss.ProcessMessageForType[*dacss.ImplicateReceiveMessage](PssMessage.Data, sender, node, dacss.ImplicateReceiveMessageType)
	case dacss.ShareRecoveryMessageType:
		dpss.ProcessMessageForType[*dacss.ShareRecoveryMessage](PssMessage.Data, sender, node, dacss.ShareRecoveryMessageType)
	case dacss.ReceiveShareRecoveryMessageType:
		dpss.ProcessMessageForType[*dacss.ReceiveShareRecoveryMessage](PssMessage.Data, sender, node, dacss.ReceiveShareRecoveryMessageType)
	case dacss.DacssOutputMessageType:
		dpss.ProcessMessageForType[*dacss.DacssOutputMessage](PssMessage.Data, sender, node, dacss.DacssOutputMessageType)
	case dacss.DacssCommitmentMessageType:
		dpss.ProcessMessageForType[*dacss.DacssCommitmentMessage](PssMessage.Data, sender, node, dacss.DacssCommitmentMessageType)

	default:
		log.Infof("No handler found. MsgType=%s", PssMessage.Type)

	}
}

// message processor for DPSS
func (processor DpssEndToEndMessageProcessor) ProcessDpssMessages(sender common.NodeDetails, PssMessage common.PSSMessage, node *testutils.IntegrationTestNode) {
	switch PssMessage.Type {
	case old_committee.DpssHimHandlerType:
		dpss.ProcessMessageForType[*old_committee.DpssHimMessage](PssMessage.Data, sender, node, old_committee.DpssHimHandlerType)
	case old_committee.InitRecHandlerType:
		dpss.ProcessMessageForType[*old_committee.InitRecMessage](PssMessage.Data, sender, node, old_committee.InitRecHandlerType)
	case old_committee.PreprocessBatchRecMessageType:
		dpss.ProcessMessageForType[*old_committee.PreprocessBatchRecMessage](PssMessage.Data, sender, node, old_committee.PreprocessBatchRecMessageType)
	case old_committee.PrivateRecMessageType:
		dpss.ProcessMessageForType[*old_committee.PrivateRecMsg](PssMessage.Data, sender, node, old_committee.PrivateRecMessageType)
	case old_committee.PublicRecMessageType:
		dpss.ProcessMessageForType[*old_committee.PublicRecMsg](PssMessage.Data, sender, node, old_committee.PublicRecMessageType)
	default:
		log.Infof("No handler found. MsgType=%s", PssMessage.Type)

	}
}

// TODO: MBVA handlers needed
// message processor for MVBA
// func (processor DpssEndToEndMessageProcessor) ProcessMvbaMessages(sender common.NodeDetails, PssMessage common.PSSMessage, node *testutils.IntegrationTestNode) {
// 	//PLaceholder for processing MVBA messge handlers
// }