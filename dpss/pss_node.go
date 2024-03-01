package dpss

import (
	"strconv"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

// PSSNode represents a node participating in the DPSS protocol.
type PSSNode struct {
	PssNodeTransport PssNodeTransport
	common.BaseNode
	state             common.PSSNodeState
	OldCommitteeNodes common.NodeNetwork // Set of nodes belonging to the old committee.
	NewCommitteeNodes common.NodeNetwork // Set of nodes belonging to the new committee.
	details           common.NodeDetails
}

// Creates a new PSSNode
func NewPSSNode(broker common.MessageBroker, nodeDetails common.NodeDetails, oldCommittee []common.NodeDetails,
	newCommittee []common.NodeDetails, bus eventbus.Bus, tOldCommittee int, kOldCommittee int,
	tNewCommittee int, kNewCommittee int, privateKey curves.Scalar) (*PSSNode, error) {
	transport := NewPssNodeTransport(bus, getPSSProtocolPrefix(1), "dpss-transport")

	// Creates the committees
	oldCommitteeNetwork := common.NodeNetwork{
		N:     len(oldCommittee),
		T:     tOldCommittee,
		K:     kOldCommittee,
		Nodes: common.MapFromNodeList(oldCommittee),
	}
	newCommiteeNetwork := common.NodeNetwork{
		N:     len(newCommittee),
		T:     tNewCommittee,
		K:     kNewCommittee,
		Nodes: common.MapFromNodeList(newCommittee),
	}

	// Defines public key.
	g := curves.K256().NewGeneratorPoint()
	publicKey := g.Mul(privateKey)

	// Creates the new node.
	newPSSNode := &PSSNode{
		BaseNode: common.NewBaseNode(
			&broker,
			nodeDetails,
			privateKey,
			publicKey,
		),
		OldCommitteeNodes: oldCommitteeNetwork,
		NewCommitteeNodes: newCommiteeNetwork,
	}

	transport.Init()
	transport.SetPSSNode(newPSSNode)

	return newPSSNode, nil
}

// Returns the PSS protocol prefix in the form dpss-<epoch>
func getPSSProtocolPrefix(epoch int) ProtocolPrefix {
	return ProtocolPrefix("dpss" + "-" + strconv.Itoa(epoch) + "/")
}

// IsOldNode determines if the current node belongs to the old committee.
func (node *PSSNode) IsOldNode() bool {
	nodeDetails := node.details
	_, found := node.OldCommitteeNodes.Nodes[nodeDetails.ToNodeDetailsID()]
	return found
}

// PublicKey returns the public key of the node with index idx that belongs to
// the old or new commitee according to the fromNewCommittee flag.
func (node *PSSNode) GetPublicKeyFor(idx int, fromNewCommittee bool) curves.Point {
	nodes := node.Nodes(fromNewCommittee)
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

// Params return the parameters for the network of the old or new committee
// according to the fromNewCommittee flag. It returns the parameters in a tuple
// with the following order:
//   - n: number of nodes in the committee.
//   - k: number of corrupt nodes in the committee.
//   - t: the reconstruction threshold in that committee.
func (node *PSSNode) Params() (n, k, t int) {
	if node.IsOldNode() {
		n = node.OldCommitteeNodes.N
		k = node.OldCommitteeNodes.K
		t = node.OldCommitteeNodes.T
	} else {
		n = node.NewCommitteeNodes.N
		k = node.NewCommitteeNodes.K
		t = node.NewCommitteeNodes.T
	}
	return
}

// Broadcast broadcasts a message to the given committee determined by the flag
// toNewCommittee.
func (node *PSSNode) Broadcast(toNewCommittee bool, msg common.PSSMessage) {
	nodesToBroadcast := node.Nodes(toNewCommittee)
	for _, n := range nodesToBroadcast {
		go func(receiver common.NodeDetails) {
			err := node.PssNodeTransport.Send(receiver, msg)
			if err != nil {
				log.WithField("Error", err).Error("Node.Broadcast()")
			}
		}(n)
	}
}

// Nodes returns the set of nodes of the old or new committee according to the flag
// fromNewCommitte.
func (node *PSSNode) Nodes(fromNewCommittee bool) map[common.NodeDetailsID]common.NodeDetails {
	if fromNewCommittee {
		return node.NewCommitteeNodes.Nodes
	} else {
		return node.OldCommitteeNodes.Nodes
	}
}

// TODO: Implement this as long as we implement the DPSS protocol.
func (node *PSSNode) ProcessMessage(senderDetails common.NodeDetails, message common.PSSMessage) error {
	return nil
}

// TODO: Implement this as long as we implement the DPSS protocol.
func (node *PSSNode) ProcessBroadcastMessage(message common.PSSMessage) error {
	return nil
}

// Details returns the details of the node, namely, its index and public key.
func (node *PSSNode) Details() common.NodeDetails {
	return node.details
}

//TODO: Do we need this?
// func GenerateDPSSRoundID(rindex, noOfRandoms big.Int) common.PSSRoundID {
// 	index := strings.Join([]string{rindex.Text(16), noOfRandoms.Text(16)}, common.Delimiter2)
// 	return common.PSSRoundID(strings.Join([]string{"DPSS", index}, common.Delimiter3))
// }

// Send sends a message to the node that has certain public key and index.
func (node *PSSNode) Send(n common.NodeDetails, msg common.PSSMessage) error {
	return node.PssNodeTransport.Send(n, msg)
}
