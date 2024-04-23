package old_committee

import (
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/old_committee"
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
	}
}
