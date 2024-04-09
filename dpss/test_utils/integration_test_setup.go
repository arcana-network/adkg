package testutils

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

// Setup information for Integration tests
type IntegrationTestSetup struct {
	OldCommitteeParams   common.CommitteeParams
	NrFaultyOldCommittee int
	NewCommitteeParams   common.CommitteeParams
	NrFaultyNewCommittee int

	OldCommitteeNetwork []*IntegrationTestNode
	NewCommitteeNetwork []*IntegrationTestNode
}

type NodeCreator func(index int, keypair common.KeyPair, transport *IntegrationMockTransport, isFaulty, isNewCommittee bool) *IntegrationTestNode

/*
Creates a new IntegrationTestSetup with the given parameters
note that the parameter `nodeCreator` is a function, which creates a new testNode.
this is how this function can be used by different integrations test sets
*/
func NewIntegrationTestSetup(
	oldCommitteeParams,
	newCommitteeParams common.CommitteeParams,
	nrFaultyOldCommittee,
	nrFaultyNewCommittee int,
	nodeCreator NodeCreator,
) (*IntegrationTestSetup, *IntegrationMockTransport) {
	setup := &IntegrationTestSetup{
		OldCommitteeParams:   oldCommitteeParams,
		NrFaultyOldCommittee: nrFaultyOldCommittee,
		NewCommitteeParams:   newCommitteeParams,
		NrFaultyNewCommittee: nrFaultyNewCommittee,
		OldCommitteeNetwork:  make([]*IntegrationTestNode, oldCommitteeParams.N),
		NewCommitteeNetwork:  make([]*IntegrationTestNode, newCommitteeParams.N),
	}

	sharedTransport := NewIntegrationMockTransport(nil, nil)

	// Create the nodes of the Old Committee
	for i := 0; i < oldCommitteeParams.N; i++ {
		isFaulty := i < nrFaultyOldCommittee // Mark first 'nrFaulty' nodes as faulty
		keypair := common.GenerateKeyPair(TestCurve())
		// Make sure the index starts at 1
		// invoke the function nodeCreator, which was passed as an argument
		node := nodeCreator(i+1, keypair, sharedTransport, isFaulty, false)
		setup.OldCommitteeNetwork[i] = node
	}

	// Create the nodes of the New Committee
	for i := 0; i < newCommitteeParams.N; i++ {
		isFaulty := i < nrFaultyNewCommittee // Mark first 'nrFaulty' nodes as faulty
		keypair := common.GenerateKeyPair(TestCurve())
		// Make sure the index starts at 1
		// invoke the function nodeCreator, which was passed as an argument
		node := nodeCreator(i+1, keypair, sharedTransport, isFaulty, true)
		setup.NewCommitteeNetwork[i] = node
	}

	// Add check all pub keys of the nodes are distinct
	pubKeys := make(map[curves.Point]bool)
	for _, node := range setup.OldCommitteeNetwork {
		pubKeys[node.LongtermKey.PublicKey] = true
	}
	for _, node := range setup.NewCommitteeNetwork {
		pubKeys[node.LongtermKey.PublicKey] = true
	}
	if len(pubKeys) != oldCommitteeParams.N+newCommitteeParams.N {
		panic("Some nodes have the same public key")
	}

	// Update the transport with the created nodes
	sharedTransport.Init(setup.OldCommitteeNetwork, setup.NewCommitteeNetwork)

	return setup, sharedTransport
}

func (setup *IntegrationTestSetup) GetSingleNode(newCommittee bool) *IntegrationTestNode {
	if newCommittee {
		return setup.NewCommitteeNetwork[0]
	} else {
		return setup.OldCommitteeNetwork[0]
	}
}

func (setup *IntegrationTestSetup) GetCommittee(newCommittee bool) []*IntegrationTestNode {
	if newCommittee {
		return setup.NewCommitteeNetwork
	} else {
		return setup.OldCommitteeNetwork
	}
}
