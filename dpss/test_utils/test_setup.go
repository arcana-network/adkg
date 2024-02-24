package testutils

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

type CommitteeTestParams struct {
	isNewCommittee bool
	// n: total number of nodes in this committee.
	N int
	// number of faulty nodes in this committee.
	nrFaulty int
	// k: number of corrupt nodes in this committee.
	K int
	// t: the reconstruction threshold in this committee.
	// fixed to k-1
	T int
}

// NewCommitteeParams creates a CommitteeTestParams for a single committee
func NewCommitteeParams(N, nrFaulty, K int, isNewCommittee bool) CommitteeTestParams {
	return CommitteeTestParams{
		isNewCommittee: isNewCommittee,
		N:              N,
		nrFaulty:       nrFaulty,
		K:              K,
		T:              K - 1,
	}
}

// a standard committe has 7 honest nodes and t=3
func StandardCommitteeParams(isNewCommittee bool) CommitteeTestParams {
	return NewCommitteeParams(7, 0, 3, isNewCommittee)
}

// returns the standard params for old & new committees at once
func StandardCommitteesParams() (oldCommitteeParams, newCommitteeParams CommitteeTestParams) {
	return StandardCommitteeParams(false), StandardCommitteeParams(true)
}

type TestSetup struct {
	oldCommitteeParams CommitteeTestParams
	newCommitteeParams CommitteeTestParams

	oldCommitteeNetwork []*PssTestNode
	newCommitteeNetwork []*PssTestNode
}

// NewTestSetup creates a complete TestSetup with the given committee parameters.
// It generates all the nodes necessary and connects them with a standard NoSendMockTransport.
func NewTestSetup(oldCommitteeParams, newCommitteeParams CommitteeTestParams) *TestSetup {
	setup := &TestSetup{
		oldCommitteeParams:  oldCommitteeParams,
		newCommitteeParams:  newCommitteeParams,
		oldCommitteeNetwork: make([]*PssTestNode, oldCommitteeParams.N),
		newCommitteeNetwork: make([]*PssTestNode, newCommitteeParams.N),
	}

	// This MockTransport doesn't pass on any msg, it only registers what has been sent/broadcasted.
	sharedTransport := NewNoSendMockTransport(nil, nil)

	// Create the nodes of the Old Committee
	for i := 0; i < oldCommitteeParams.N; i++ {
		isFaulty := i < oldCommitteeParams.nrFaulty // Mark first 'nrFaulty' nodes as faulty
		keypair := common.GenerateKeyPair(curves.K256())
		node := NewEmptyNode(i, keypair, sharedTransport, isFaulty, false)
		setup.oldCommitteeNetwork[i] = node
	}

	// Create the nodes of the New Committee
	for i := 0; i < newCommitteeParams.N; i++ {
		isFaulty := i < newCommitteeParams.nrFaulty // Mark first 'nrFaulty' nodes as faulty
		keypair := common.GenerateKeyPair(curves.K256())
		node := NewEmptyNode(i, keypair, sharedTransport, isFaulty, true)
		setup.newCommitteeNetwork[i] = node
	}

	// Update the transport with the created nodes
	sharedTransport.Init(setup.oldCommitteeNetwork, setup.newCommitteeNetwork)

	return setup
}
