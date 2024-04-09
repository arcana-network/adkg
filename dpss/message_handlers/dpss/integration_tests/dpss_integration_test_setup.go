package dpss

import (
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
)

func DacssIntegrationTestSetup() (*testutils.IntegrationTestSetup, *testutils.IntegrationMockTransport) {
	oldCommitteeParams, newCommitteeParams := testutils.StandardCommitteesParams()
	return testutils.NewIntegrationTestSetup(oldCommitteeParams, newCommitteeParams, testutils.DefaultNrFaulty_old, testutils.DefaultNrFaulty_new, NewDpssTestNode)
}
