package testutils

import (
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

// Implements the PssParticipant interface
type PssTestNode struct {
	// Index & PubKey of this node
	details             common.NodeDetails
	isNewCommittee      bool
	committeeTestParams CommitteeTestParams

	state    *common.PSSNodeState
	Keypair  common.KeyPair
	isFaulty bool

	transport *NoSendMockTransport

	//shares of old/new committee
	//false: old, true: new
	shares map[bool]map[int64]*big.Int
}

func (n *PssTestNode) State() *common.PSSNodeState {
	return n.state
}

func (n *PssTestNode) ID() int {
	return n.details.Index
}

func (n *PssTestNode) IsOldNode() bool {
	return !n.isNewCommittee
}

func (n *PssTestNode) PublicKey(idx int, fromNewCommittee bool) curves.Point {
	return n.Keypair.PublicKey
}

func (node *PssTestNode) Params() (n int, k int, t int) {
	return node.committeeTestParams.N, node.committeeTestParams.K, node.committeeTestParams.T
}

func (node *PssTestNode) Broadcast(toNewCommittee bool, msg common.PSSMessage) {
	node.transport.Broadcast(node.Details(), msg)
}

func (node *PssTestNode) Send(n common.NodeDetails, msg common.PSSMessage) error {
	node.transport.Send(node.Details(), n, msg)
	return nil
}

func (n *PssTestNode) Details() common.NodeDetails {
	return n.details
}

func (n *PssTestNode) PrivateKey() curves.Scalar {
	return n.Keypair.PrivateKey
}

// only register a message was received, no further action
func (node *PssTestNode) ReceiveMessage(sender common.NodeDetails, pssMessage common.PSSMessage) {
	node.transport.receivedMessages = append(node.transport.receivedMessages, pssMessage)
}

func (n *PssTestNode) Nodes(fromNewCommittee bool) map[common.NodeDetailsID]common.NodeDetails {
	nodes := make(map[common.NodeDetailsID]common.NodeDetails)
	for _, node := range n.transport.nodesOld {
		nodes[node.Details().GetNodeDetailsID()] = node.details
	}
	for _, node := range n.transport.nodesNew {
		nodes[node.Details().GetNodeDetailsID()] = node.details
	}
	return nodes
}

func NewEmptyNode(index int, keypair common.KeyPair, noSendTransport *NoSendMockTransport, isFaulty, isNewCommittee bool) *PssTestNode {
	node := PssTestNode{
		details:             common.NodeDetails{Index: index, PubKey: common.CurvePointToPoint(keypair.PublicKey, common.SECP256K1)},
		isNewCommittee:      isNewCommittee,
		committeeTestParams: StandardCommitteeParams(isNewCommittee),
		state: &common.PSSNodeState{
			ShareStore: &common.PSSShareStoreMap{},
			RbcStore:   &common.RBCStateMap{},
		},
		transport: noSendTransport,
		Keypair:   keypair,
		isFaulty:  isFaulty,

		shares: make(map[bool]map[int64]*big.Int),
	}
	return &node
}

// Get a single old/new node
// This doesn't play well to combine with other nodes,
// as it has a fixed index (1) and it is not connected to other nodes
func GetSingleNode(isNewCommittee bool, isFaulty bool) (*PssTestNode, *NoSendMockTransport) {
	nodesOld := []*PssTestNode{}
	nodesNew := []*PssTestNode{}
	keypair := common.GenerateKeyPair(curves.K256())
	transport := NewNoSendMockTransport(nodesOld, nodesNew)

	node := NewEmptyNode(1, keypair, transport, isFaulty, isNewCommittee)

	if isNewCommittee {
		transport.Init(nodesOld, []*PssTestNode{node})
	} else {
		transport.Init([]*PssTestNode{node}, nodesNew)
	}

	return node, transport
}
