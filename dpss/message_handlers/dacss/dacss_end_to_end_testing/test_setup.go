package dacss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

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

	oldCommitteeNetwork []*PssTestNode2
	newCommitteeNetwork []*PssTestNode2
}

func DefaultTestSetup() (*TestSetup, *MockTransport) {
	oldCommitteeParams, newCommitteeParams := StandardCommitteesParams()
	return NewTestSetup(oldCommitteeParams, newCommitteeParams, DefaultNrFaulty_old, DefaultNrFaulty_new)
}

// build upon the noSend test setup from test_utils

// NewTestSetup creates a complete TestSetup with the given committee parameters.
// It generates all the nodes necessary and connects them with a standard MockTransport.
func NewTestSetup(oldCommitteeParams, newCommitteeParams common.CommitteeParams, nrFaultyOldCommittee, nrFaultyNewCommittee int) (*TestSetup, *MockTransport) {
	setup := &TestSetup{
		OldCommitteeParams:   oldCommitteeParams,
		NrFaultyOldCommittee: nrFaultyOldCommittee,
		NewCommitteeParams:   newCommitteeParams,
		NrFaultyNewCommittee: nrFaultyNewCommittee,
		oldCommitteeNetwork:  make([]*PssTestNode2, oldCommitteeParams.N),
		newCommitteeNetwork:  make([]*PssTestNode2, newCommitteeParams.N),
	}

	sharedTransport := NewMockTransport(nil, nil)

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

	return setup, sharedTransport
}

func (setup *TestSetup) GetSingleNode(newCommittee bool) *PssTestNode2 {
	if newCommittee {
		return setup.newCommitteeNetwork[0]
	} else {
		return setup.oldCommitteeNetwork[0]
	}
}

func (setup *TestSetup) GetCommittee(newCommittee bool) []*PssTestNode2 {
	if newCommittee {
		return setup.newCommitteeNetwork
	} else {
		return setup.oldCommitteeNetwork
	}
}
