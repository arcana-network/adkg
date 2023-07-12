package dpss

import (
	"fmt"
	"strings"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/aba"
	batchreconstruction "github.com/arcana-network/dkgnode/dpss/message_handlers/batch_reconstruction"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/him"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/keyset"
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
	log.WithFields(log.Fields{
		"sender":   sender.Index,
		"receiver": n.ID(),
		"Method":   msg.Method,
		"RoundID":  msg.PSSID,
	}).Debug("DPSSNode:ProcessMessage()")

	switch {
	case strings.HasPrefix(string(msg.Method), "dacss"):
		n.ProcessDACSSMessages(sender, msg)
	case strings.HasPrefix(string(msg.Method), "keyset"):
		n.ProcessKeysetMessages(sender, msg)
	case strings.HasPrefix(string(msg.Method), "aba"):
		n.ProcessABAMessages(sender, msg)
	case strings.HasPrefix(string(msg.Method), "br"):
		n.ProcessBatchReconsMessages(sender, msg)
	case strings.HasPrefix(string(msg.Method), "him"):
		n.ProcessHIMMessages(sender, msg)

	default:
		log.Infof("No handler found. MsgType=%s", msg.Method)
		return fmt.Errorf("dpssMessage method %v not found", msg.Method)
	}
	return nil
}

func (n *PSSNode) ProcessDACSSMessages(sender common.KeygenNodeDetails, rawMsg common.DPSSMessage) {
	switch rawMsg.Method {
	case dacss.ShareMessageType:
		var msg dacss.ShareMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, n)

	case dacss.ProposeMessageType:
		var msg dacss.ProposeMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, n)
	case dacss.EchoMessageType:
		var msg dacss.EchoMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, n)
	case dacss.ReadyMessageType:
		var msg dacss.ReadyMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, n)
	case dacss.CommitMessageType:
		var msg dacss.CommitMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, n)
	case dacss.OutputMessageType:
		var msg dacss.OutputMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, n)
	}
}

func (n *PSSNode) ProcessKeysetMessages(sender common.KeygenNodeDetails, rawMsg common.DPSSMessage) {
	switch rawMsg.Method {
	case keyset.InitMessageType:
		var msg keyset.InitMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, n)
	case keyset.ProposeMessageType:
		var msg keyset.ProposeMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, n)
	case keyset.EchoMessageType:
		var msg keyset.EchoMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, n)
	case keyset.ReadyMessageType:
		var msg keyset.ReadyMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, n)
	case keyset.OutputMessageType:
		var msg keyset.OutputMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, n)
	}
}

func (node *PSSNode) ProcessABAMessages(sender common.KeygenNodeDetails, rawMsg common.DPSSMessage) {
	switch rawMsg.Method {
	case aba.InitMessageType:
		var msg aba.InitMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, node)
	case aba.Est1MessageType:
		var msg aba.Est1Message
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, node)
	case aba.Aux1MessageType:
		var msg aba.Aux1Message
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, node)
	case aba.AuxsetMessageType:
		var msg aba.AuxsetMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, node)
	case aba.Est2MessageType:
		var msg aba.Est2Message
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, node)
	case aba.Aux2MessageType:
		var msg aba.Aux2Message
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, node)
	case aba.CoinInitMessageType:
		var msg aba.CoinInitMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, node)
	case aba.CoinMessageType:
		var msg aba.CoinMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, node)
	}
}

func (node *PSSNode) ProcessBatchReconsMessages(sender common.KeygenNodeDetails, rawMsg common.DPSSMessage) {
	switch rawMsg.Method {
	case batchreconstruction.InitMessageType:
		var msg batchreconstruction.InitBatchMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, node)
	case batchreconstruction.ReconsMessageType:
		var msg batchreconstruction.InitReconsMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, node)
	case batchreconstruction.DecodeMessageType:
		var msg batchreconstruction.InitDecodeMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, node)
	}
}

func (node *PSSNode) ProcessHIMMessages(sender common.KeygenNodeDetails, rawMsg common.DPSSMessage) {
	switch rawMsg.Method {
	case him.InitMessageType:
		var msg him.InitMessage
		err := bijson.Unmarshal(rawMsg.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", rawMsg.Method)
			return
		}
		msg.Process(sender, node)
	}
}

func (n *PSSNode) ProcessBroadcastMessage(msg common.DPSSMessage) error {
	return nil
}
