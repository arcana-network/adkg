package testutils

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

// The default parameters for old & new committee are distinct on purpose,
// to make sure the correct ones are being used
const DefaultN_old = 7
const DefaultNrFaulty_old = 0
const DefaultK_old = 2

const DefaultN_new = 8
const DefaultNrFaulty_new = 0
const DefaultK_new = 2

// NewCommitteeParams creates a common.CommitteeParams for a single committee
func NewCommitteeParams(N, K int) common.CommitteeParams {
	return common.CommitteeParams{
		N: N,
		K: K,
		T: K - 1,
	}
}

// a standard old committe has 7 honest nodes and t=2
func StandardOldCommitteeParams() common.CommitteeParams {
	return NewCommitteeParams(DefaultN_old, DefaultK_old)
}

// a standard new committe has 8 honest nodes and t=2
func StandardNewCommitteeParams() common.CommitteeParams {
	return NewCommitteeParams(DefaultN_new, DefaultK_new)
}

// returns the standard params for old & new committees at once
func StandardCommitteesParams() (oldCommitteeParams, newCommitteeParams common.CommitteeParams) {
	return StandardOldCommitteeParams(), StandardNewCommitteeParams()
}

type TestSetup struct {
	OldCommitteeParams   common.CommitteeParams
	NrFaultyOldCommittee int
	NewCommitteeParams   common.CommitteeParams
	NrFaultyNewCommittee int

	oldCommitteeNetwork []*PssTestNode
	newCommitteeNetwork []*PssTestNode
}

func DefaultTestSetup() *TestSetup {
	oldCommitteeParams, newCommitteeParams := StandardCommitteesParams()
	return NewTestSetup(oldCommitteeParams, newCommitteeParams, DefaultNrFaulty_old, DefaultNrFaulty_new)
}

// NewTestSetup creates a complete TestSetup with the given committee parameters.
// It generates all the nodes necessary and connects them with a standard NoSendMockTransport.
func NewTestSetup(oldCommitteeParams, newCommitteeParams common.CommitteeParams, nrFaultyOldCommittee, nrFaultyNewCommittee int) *TestSetup {
	setup := &TestSetup{
		OldCommitteeParams:   oldCommitteeParams,
		NrFaultyOldCommittee: nrFaultyOldCommittee,
		NewCommitteeParams:   newCommitteeParams,
		NrFaultyNewCommittee: nrFaultyNewCommittee,
		oldCommitteeNetwork:  make([]*PssTestNode, oldCommitteeParams.N),
		newCommitteeNetwork:  make([]*PssTestNode, newCommitteeParams.N),
	}

	// This MockTransport doesn't pass on any msg, it only registers what has been sent/broadcasted.
	sharedTransport := NewNoSendMockTransport(nil, nil)

	// Create the nodes of the Old Committee
	for i := 0; i < oldCommitteeParams.N; i++ {
		isFaulty := i < nrFaultyOldCommittee // Mark first 'nrFaulty' nodes as faulty
		keypair := common.GenerateKeyPair(curves.K256())
		// Make sure the index starts at 1
		node := NewEmptyNode(i+1, keypair, sharedTransport, isFaulty, false)
		setup.oldCommitteeNetwork[i] = node
	}

	// Create the nodes of the New Committee
	for i := 0; i < newCommitteeParams.N; i++ {
		isFaulty := i < nrFaultyNewCommittee // Mark first 'nrFaulty' nodes as faulty
		keypair := common.GenerateKeyPair(curves.K256())
		// Make sure the index starts at 1
		node := NewEmptyNode(i+1, keypair, sharedTransport, isFaulty, true)
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

	return setup
}

// Returns just a node in the old committee from the given test setup.
func (setup *TestSetup) GetSingleOldNodeFromTestSetup() *PssTestNode {
	return setup.oldCommitteeNetwork[0]
}

// Returns two nodes in the old committee for a given test setup.
func (setup *TestSetup) GetTwoOldNodesFromTestSetup() (*PssTestNode, *PssTestNode) {
	return setup.oldCommitteeNetwork[0], setup.oldCommitteeNetwork[1]
}

// Returns three nodes in the old committee for a given test setup.
func (setup *TestSetup) GetThreeOldNodesFromTestSetup() (*PssTestNode, *PssTestNode, *PssTestNode) {
	return setup.oldCommitteeNetwork[0], setup.oldCommitteeNetwork[1], setup.oldCommitteeNetwork[2]
}

// Returns all nodes in the old committee for a given test setup.
func (setup *TestSetup) GetAllOldNodesFromTestSetup() []*PssTestNode {

	nodes := make([]*PssTestNode, 0)
	nodes = append(nodes, setup.newCommitteeNetwork...)

	return nodes
}

// Returns a node in the new committee from a given test setup.
func (setup *TestSetup) GetSingleNewNodeFromTestSetup() *PssTestNode {
	return setup.newCommitteeNetwork[0]
}

// Returns a receiver node and 2k + 1 nodes in a given committee
func (setup *TestSetup) GetDealerAnd2kPlusOneNodes(fromOldCommittee bool) (*PssTestNode, []*PssTestNode) {
	if fromOldCommittee {
		k := setup.OldCommitteeParams.K
		receiver := setup.oldCommitteeNetwork[0]
		group := setup.oldCommitteeNetwork[1 : 2*k+2]
		return receiver, group
	} else {
		k := setup.NewCommitteeParams.K
		receiver := setup.newCommitteeNetwork[0]
		group := setup.newCommitteeNetwork[1 : 2*k+2]
		return receiver, group
	}
}
