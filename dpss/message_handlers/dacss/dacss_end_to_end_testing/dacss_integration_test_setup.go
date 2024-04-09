package dacss

import (
	"github.com/arcana-network/dkgnode/common"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

type IntegrationTestSetup struct {
	OldCommitteeParams   common.CommitteeParams
	NrFaultyOldCommittee int
	NewCommitteeParams   common.CommitteeParams
	NrFaultyNewCommittee int

	oldCommitteeNetwork []*testutils.IntegrationTestNode
	newCommitteeNetwork []*testutils.IntegrationTestNode
}

func DacssIntegrationTestSetup() (*IntegrationTestSetup, *testutils.IntegrationMockTransport) {
	oldCommitteeParams, newCommitteeParams := testutils.StandardCommitteesParams()
	return NewTestSetup(oldCommitteeParams, newCommitteeParams, testutils.DefaultNrFaulty_old, testutils.DefaultNrFaulty_new)
}

// NewTestSetup creates a complete TestSetup with the given committee parameters.
// It generates all the nodes necessary and connects them with a standard MockTransport.
func NewTestSetup(oldCommitteeParams, newCommitteeParams common.CommitteeParams, nrFaultyOldCommittee, nrFaultyNewCommittee int) (*IntegrationTestSetup, *testutils.IntegrationMockTransport) {
	setup := &IntegrationTestSetup{
		OldCommitteeParams:   oldCommitteeParams,
		NrFaultyOldCommittee: nrFaultyOldCommittee,
		NewCommitteeParams:   newCommitteeParams,
		NrFaultyNewCommittee: nrFaultyNewCommittee,
		oldCommitteeNetwork:  make([]*testutils.IntegrationTestNode, oldCommitteeParams.N),
		newCommitteeNetwork:  make([]*testutils.IntegrationTestNode, newCommitteeParams.N),
	}

	sharedTransport := testutils.NewMockTransport(nil, nil)

	// Create the nodes of the Old Committee
	for i := 0; i < oldCommitteeParams.N; i++ {
		isFaulty := i < nrFaultyOldCommittee // Mark first 'nrFaulty' nodes as faulty
		keypair := common.GenerateKeyPair(curves.K256())
		// Make sure the index starts at 1
		node := NewDacssTestNode(i+1, keypair, sharedTransport, isFaulty, false)
		setup.oldCommitteeNetwork[i] = node
	}

	// Create the nodes of the New Committee
	for i := 0; i < newCommitteeParams.N; i++ {
		isFaulty := i < nrFaultyNewCommittee // Mark first 'nrFaulty' nodes as faulty
		keypair := common.GenerateKeyPair(curves.K256())
		// Make sure the index starts at 1
		node := NewDacssTestNode(i+1, keypair, sharedTransport, isFaulty, true)
		setup.newCommitteeNetwork[i] = node
	}

	// Add check all pub keys of the nodes are distinct
	pubKeys := make(map[curves.Point]bool)
	for _, node := range setup.oldCommitteeNetwork {
		pubKeys[node.LongtermKey.PublicKey] = true
	}
	for _, node := range setup.newCommitteeNetwork {
		pubKeys[node.LongtermKey.PublicKey] = true
	}
	if len(pubKeys) != oldCommitteeParams.N+newCommitteeParams.N {
		panic("Some nodes have the same public key")
	}

	// Update the transport with the created nodes
	sharedTransport.Init(setup.oldCommitteeNetwork, setup.newCommitteeNetwork)

	return setup, sharedTransport
}

func (setup *IntegrationTestSetup) GetSingleNode(newCommittee bool) *testutils.IntegrationTestNode {
	if newCommittee {
		return setup.newCommitteeNetwork[0]
	} else {
		return setup.oldCommitteeNetwork[0]
	}
}

func (setup *IntegrationTestSetup) GetCommittee(newCommittee bool) []*testutils.IntegrationTestNode {
	if newCommittee {
		return setup.newCommitteeNetwork
	} else {
		return setup.oldCommitteeNetwork
	}
}
