package dpss

import (
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dpss"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
)

// Helper for DPSS Integration testing
type DpssMessageProcessor struct {
}

// Returns an IntegrationTestNode that has the DPSS message processor
func NewDpssTestNode(index int, keypair common.KeyPair, transport *testutils.IntegrationMockTransport, isFaulty, isNewCommittee bool) *testutils.IntegrationTestNode {
	return testutils.NewIntegrationTestNode(index, keypair, transport, isFaulty, isNewCommittee, DpssMessageProcessor{})
}

// Exact definition of what messages are to be passed on for this processor
// only messages with prefix dpss are passed on for this processor
func (processor DpssMessageProcessor) ProcessMessages(sender common.NodeDetails, PssMessage common.PSSMessage, node *testutils.IntegrationTestNode) {
	if !(strings.HasPrefix(PssMessage.Type, "dpss")) {
		return
	}

	switch PssMessage.Type {
	case dpss.DpssHimHandlerType:
		testutils.ProcessMessageForType[*dpss.DpssHimMessage](PssMessage.Data, sender, node, dpss.DpssHimHandlerType)
	case dpss.InitRecHandlerType:
		testutils.ProcessMessageForType[*dpss.InitRecMessage](PssMessage.Data, sender, node, dpss.InitRecHandlerType)
	case dpss.PreprocessBatchRecMessageType:
		testutils.ProcessMessageForType[*dpss.PreprocessBatchRecMessage](PssMessage.Data, sender, node, dpss.PreprocessBatchRecMessageType)
	case dpss.PrivateRecMessageType:
		testutils.ProcessMessageForType[*dpss.PrivateRecMsg](PssMessage.Data, sender, node, dpss.PrivateRecMessageType)
	case dpss.PublicRecMessageType:
		testutils.ProcessMessageForType[*dpss.PublicRecMsg](PssMessage.Data, sender, node, dpss.PublicRecMessageType)
	}
}
