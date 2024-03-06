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
	committeeTestParams common.CommitteeParams

	state       *common.PSSNodeState
	LongtermKey common.KeyPair
	isFaulty    bool

	Transport *NoSendMockTransport

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

// This requires the testnode to actually have the new committee/old committee nodes
func (n *PssTestNode) GetPublicKeyFor(idx int, fromNewCommittee bool) curves.Point {
	nodes := n.Nodes(fromNewCommittee)
	for _, n := range nodes {
		if n.Index == idx {
			pk, err := curves.K256().NewIdentityPoint().Set(&n.PubKey.X, &n.PubKey.Y)
			if err != nil {
				return nil
			}
			return pk
		}
	}
	return nil
}

func (node *PssTestNode) Params() (n int, k int, t int) {
	return node.committeeTestParams.N, node.committeeTestParams.K, node.committeeTestParams.T
}

func (node *PssTestNode) Broadcast(toNewCommittee bool, msg common.PSSMessage) {
	node.Transport.Broadcast(node.Details(), msg)
}

func (node *PssTestNode) Send(n common.NodeDetails, msg common.PSSMessage) error {
	node.Transport.Send(node.Details(), n, msg)
	return nil
}

func (n *PssTestNode) Details() common.NodeDetails {
	return n.details
}

func (n *PssTestNode) PrivateKey() curves.Scalar {
	return n.LongtermKey.PrivateKey
}

// only register a message was received, no further action
func (node *PssTestNode) ReceiveMessage(sender common.NodeDetails, pssMessage common.PSSMessage) {
	node.Transport.ReceivedMessages = append(node.Transport.ReceivedMessages, pssMessage)
}

func (n *PssTestNode) Nodes(fromNewCommittee bool) map[common.NodeDetailsID]common.NodeDetails {
	var selectedNodes []*PssTestNode
	if fromNewCommittee {
		selectedNodes = n.Transport.nodesNew
	} else {
		selectedNodes = n.Transport.nodesOld
	}

	nodes := make(map[common.NodeDetailsID]common.NodeDetails, len(selectedNodes))
	for _, node := range selectedNodes {
		nodes[node.Details().GetNodeDetailsID()] = node.details
	}

	return nodes
}

// SetupRBCState sets up the default state for the RBC protocol for a given ACSS
// round ID. At the end of the execution, all the fileds are empty and the state
// reflects that the node has not sent the READY message. It returns error if
// the setup is not done correctly.
func (n *PssTestNode) SetupRBCState(acssRoundDetails common.ACSSRoundDetails) error {
	err := n.State().AcssStore.UpdateAccsState(
		acssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.RBCState = common.RBCState{
				Phase:           common.Initial,
				ReceivedEcho:    make(map[int]bool),
				ReceivedReady:   make(map[int]bool),
				ReceivedMessage: make([]byte, 0),
				IsReadyMsgSent:  false,
				HashMsg:         make([]byte, 0),
			}
		},
	)

	return err
}

func NewEmptyNode(index int, keypair common.KeyPair, noSendTransport *NoSendMockTransport, isFaulty, isNewCommittee bool) *PssTestNode {
	var params common.CommitteeParams
	if isNewCommittee {
		params = StandardNewCommitteeParams()
	} else {
		params = StandardOldCommitteeParams()
	}
	node := PssTestNode{
		details:             common.NodeDetails{Index: index, PubKey: common.CurvePointToPoint(keypair.PublicKey, common.SECP256K1)},
		isNewCommittee:      isNewCommittee,
		committeeTestParams: params,
		state: &common.PSSNodeState{
			AcssStore:       &common.AcssStateMap{},
			DualAcssStarted: false,
		},
		Transport:   noSendTransport,
		LongtermKey: keypair,
		isFaulty:    isFaulty,

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
