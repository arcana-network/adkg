package dpss

import (
	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/aba"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

type PSSNode struct {
	id              int
	fromNodeNetwork common.NodeNetwork
	toNodeNetwork   common.NodeNetwork
	fromEpoch       int
	toEpoch         int
	IsOldNode       bool
	isNewNode       bool
	StateField      *dpsscommon.NodeState
	ContractField   dpsscommon.SmartContract
	Keypair         common.KeyPair

	details   common.KeygenNodeDetails
	transport DPSSTransport
}

func NewPSSNode(id int, self common.KeygenNodeDetails,
	transport DPSSTransport,
	isNewNode, isOldNode bool,
	fromNodeNetwork, toNodeNetwork common.NodeNetwork,
	fromEpoch, toEpoch int) *PSSNode {

	node := &PSSNode{
		id:              id,
		details:         self,
		transport:       transport,
		fromEpoch:       fromEpoch,
		toEpoch:         toEpoch,
		IsOldNode:       isOldNode,
		isNewNode:       isNewNode,
		fromNodeNetwork: fromNodeNetwork,
		toNodeNetwork:   toNodeNetwork,
	}

	transport.Init()
	transport.SetNode(node)

	return node
}
func (n *PSSNode) ReceiveMessage(m common.DPSSMessage) {
	n.ProcessMessage(n.details, m)
}

func (n *PSSNode) ID() int {
	return n.id
}

func (n *PSSNode) Params(newCommittee bool) (int, int, int) {
	if newCommittee {
		return n.toNodeNetwork.N, n.toNodeNetwork.T - 1, n.toNodeNetwork.T
	}
	return n.fromNodeNetwork.N, n.fromNodeNetwork.T - 1, n.fromNodeNetwork.T
}

func (self *PSSNode) Broadcast(toNewNodes bool, m common.DPSSMessage) {
	var nodesList map[common.NodeDetailsID]common.KeygenNodeDetails
	if toNewNodes {
		nodesList = self.toNodeNetwork.Nodes
	} else {
		nodesList = self.fromNodeNetwork.Nodes
	}
	for _, n := range nodesList {
		go func(receiver common.KeygenNodeDetails) {
			err := self.transport.Send(receiver, m)
			if err != nil {
				log.WithField("Error", err).Error("Node.Broadcast()")
			}
		}(n)
	}
}

func (n *PSSNode) Send(m common.DPSSMessage, p common.KeygenNodeDetails) {
	n.transport.Send(p, m)
}

func (n *PSSNode) Nodes(newCommittee bool) map[common.NodeDetailsID]common.KeygenNodeDetails {
	if newCommittee {
		return n.toNodeNetwork.Nodes
	}
	return n.fromNodeNetwork.Nodes
}
func (n *PSSNode) State() *dpsscommon.NodeState {
	return n.StateField
}

// func (n *PSSNode) NewCommittee(committee *[]DPSSParticipant) {
// 	n.transport.NewCommittee(committee)
// }

func (n *PSSNode) PublicKey(i int) curves.Point {
	return n.ContractField.PublicKey(i)
}
func (n *PSSNode) CurveParams(c *curves.Curve) (curves.Point, curves.Point) {
	return c.NewGeneratorPoint(), c.NewGeneratorPoint()
}

func (n *PSSNode) SelfPrivateKey() curves.Scalar {
	return n.Keypair.PrivateKey
}

func (n *PSSNode) Contract() dpsscommon.SmartContract {
	return n.ContractField
}

func (n *PSSNode) ProcessMessage(sender common.KeygenNodeDetails, msg common.DPSSMessage) error {
	switch msg.Method {
	case "aux1_aba":
		log.Debugf("Got %s", aba.Aux1MessageType)
		var aux1Msg aba.Aux1Message
		err := bijson.Unmarshal(msg.Data, &aux1Msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", msg.Method)
			return err
		}
		aux1Msg.Process(sender, n)
	default:
		log.Errorf("Unhandled message:")
	}
	return nil
}

func (n *PSSNode) ProcessBroadcastMessage(msg common.DPSSMessage) error {
	return nil
}
