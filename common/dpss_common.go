package common

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

// PSSNode represents a node participating in the DPSS protocol.
type PSSNode struct {
	NodeTransport
	BaseNode
	oldCommitteeNodes NodeNetwork // Set of nodes belonging to the old committee.
	newCommitteeNodes NodeNetwork // Set of nodes belonging to the new committee.
}

func NewPSSNode(broker MessageBroker, nodeDetails KeygenNodeDetails, oldCommittee []KeygenNodeDetails, newCommittee []KeygenNodeDetails, bus eventbus.Bus, tOldCommittee int, kOldCommittee int, tNewCommittee int, kNewCommittee int, privateKey curves.Scalar) (*PSSNode, error) {
	transport := NewNodeTransport(bus, getPSSProtocolPrefix(1), "dpss-transport")

	// Creates the committees
	oldCommitteeNetwork := NodeNetwork{
		N:     len(oldCommittee),
		T:     tOldCommittee,
		K:     kOldCommittee,
		Nodes: MapFromNodeList(oldCommittee),
	}
	newCommiteeNetwork := NodeNetwork{
		N:     len(newCommittee),
		T:     tNewCommittee,
		K:     kNewCommittee,
		Nodes: MapFromNodeList(newCommittee),
	}

	// Defines public key.
	g := curves.K256().NewGeneratorPoint()
	publicKey := g.Mul(privateKey)

	// Creates the new node.
	newPSSNode := &PSSNode{
		BaseNode: BaseNode{
			broker:     &broker,
			details:    nodeDetails,
			privateKey: privateKey,
			publicKey:  publicKey,
			transport:  *transport,
		},
		oldCommitteeNodes: oldCommitteeNetwork,
		newCommitteeNodes: newCommiteeNetwork,
	}

	transport.Init()
	transport.SetNode(newPSSNode)

	return newPSSNode, nil
}

func getPSSProtocolPrefix(epoch int) ProtocolPrefix {
	return ProtocolPrefix("dpss" + "-" + strconv.Itoa(epoch) + "/")
}

// IsOldNode determines if the current node belongs to the old committee.
func (node *PSSNode) IsOldNode() bool {
	_, found := node.oldCommitteeNodes.Nodes[node.details.ToNodeDetailsID()]
	return found
}

// IsOldNode determines if the current node belongs to the new committe.
func (node *PSSNode) IsNewNode() bool {
	_, found := node.newCommitteeNodes.Nodes[node.details.ToNodeDetailsID()]
	return found
}

// PublicKey returns the public key of the node with index idx that belongs to
// the old or new commitee according to the fromNewCommittee flag.
func (node *PSSNode) PublicKey(idx int, fromNewCommittee bool) curves.Point {
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
func (node *PSSNode) Params(fromNewCommittee bool) (n, k, t int) {
	if fromNewCommittee {
		n = node.newCommitteeNodes.N
		k = node.newCommitteeNodes.K
		t = node.newCommitteeNodes.T
	} else {
		n = node.oldCommitteeNodes.N
		k = node.newCommitteeNodes.K
		t = node.newCommitteeNodes.T
	}
	return
}

// Broadcast broadcasts a message to the given committee determined by the flag
// toNewCommittee.
func (node *PSSNode) Broadcast(toNewCommittee bool, msg DKGMessage) {
	nodesToBroadcast := node.Nodes(toNewCommittee)
	for _, n := range nodesToBroadcast {
		go func(receiver KeygenNodeDetails) {
			err := node.transport.Send(receiver, msg)
			if err != nil {
				log.WithField("Error", err).Error("Node.Broadcast()")
			}
		}(n)
	}
}

// Nodes returns the set of nodes of the old or new committee according to the flag
// fromNewCommitte.
func (node *PSSNode) Nodes(fromNewCommittee bool) map[NodeDetailsID]KeygenNodeDetails {
	if fromNewCommittee {
		return node.newCommitteeNodes.Nodes
	} else {
		return node.oldCommitteeNodes.Nodes
	}
}

// TODO: Implement this as long as we implement the DPSS protocol.
func (node *PSSNode) ProcessMessage(senderDetails KeygenNodeDetails, message DKGMessage) error {
	return nil
}

// TODO: Implement this as long as we implement the DPSS protocol.
func (node *PSSNode) ProcessBroadcastMessage(message DKGMessage) error {
	return nil
}

func (node *PSSNode) NodeDetails() KeygenNodeDetails {
	return node.details
}

func GenerateDPSSID(rindex, noOfRandoms big.Int) ADKGID {
	index := strings.Join([]string{rindex.Text(16), noOfRandoms.Text(16)}, Delimiter2)
	return ADKGID(strings.Join([]string{"DPSS", index}, Delimiter3))
}
