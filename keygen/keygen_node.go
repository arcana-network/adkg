package keygen

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/eventbus"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/aba"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/acss"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/keyderivation"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/keyset"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

type KeygenNode struct {
	broker  *common.MessageBroker
	details common.KeygenNodeDetails
	CurrentNodes common.NodeNetwork
	Transport         common.NodeTransport
	state             *common.NodeState
	privateKey        curves.Scalar
	publicKey         curves.Point
	tracker           *KeygenTracker
}

// NodeDetails implements common.NodeProcessor.
func (node *KeygenNode) NodeDetails() common.KeygenNodeDetails {
	return node.details
}

func NewKeygenNode(broker *common.MessageBroker, nodeDetails common.KeygenNodeDetails,
	nodeList []common.KeygenNodeDetails, bus eventbus.Bus, T int, K int,
	privateKey curves.Scalar) (*KeygenNode, error) {
	transport := common.NewNodeTransport(bus, GetKeygenProtocolPrefix(1), "keygen-transport")
	nodeNetwork := common.NodeNetwork{
		N:     len(nodeList),
		K:     K,
		T:     T,
		Nodes: mapFromNodeList(nodeList),
	}

	g := curves.K256().NewGeneratorPoint()
	publicKey := g.Mul(privateKey)

	newKeygenNode := &KeygenNode{
		broker:       broker,
		details:      nodeDetails,
		Transport:    *transport,
		CurrentNodes: nodeNetwork,
		state: &common.NodeState{
			KeygenStore:  &common.SharingStoreMap{},
			SessionStore: &common.ADKGSessionStore{},
			ABAStore:     &common.ABAStoreMap{},
		},
		privateKey: privateKey,
		publicKey:  publicKey,
	}

	log.Info("Keygen service starting...")
	transport.Init()
	transport.SetNode(newKeygenNode)
	newKeygenNode.tracker = NewKeygenTracker(newKeygenNode.remove)
	return newKeygenNode, nil
}

func (node *KeygenNode) Params() (n, k, t int) {
	n = node.CurrentNodes.N
	k = node.CurrentNodes.K
	t = node.CurrentNodes.T
	return
}

func (node *KeygenNode) cleanup(id common.ADKGID) {
	node.cleanupKeygenStore(id)
	node.cleanupSessionStore(id)
}

func (node *KeygenNode) remove(id common.ADKGID) {
	for _, n := range node.CurrentNodes.Nodes {
		node.state.KeygenStore.Delete((&common.RoundDetails{
			ADKGID: id,
			Dealer: n.Index,
			Kind:   "acss",
		}).ID())
		keysetID := (&common.RoundDetails{
			ADKGID: id,
			Dealer: n.Index,
			Kind:   "keyset",
		}).ID()
		node.state.KeygenStore.Delete(keysetID)
		node.state.ABAStore.Delete(keysetID)
	}
	node.state.SessionStore.Delete(id)
}

func (node *KeygenNode) Cleanup(id common.ADKGID) {
	node.cleanup(id)
}

func (node *KeygenNode) cleanupKeygenStore(id common.ADKGID) {
	for _, n := range node.CurrentNodes.Nodes {
		node.state.KeygenStore.Complete((&common.RoundDetails{
			ADKGID: id,
			Dealer: n.Index,
			Kind:   "acss",
		}).ID())
		keysetID := (&common.RoundDetails{
			ADKGID: id,
			Dealer: n.Index,
			Kind:   "keyset",
		}).ID()
		node.state.KeygenStore.Complete(keysetID)
		node.state.ABAStore.Complete(keysetID)
	}
}

func (node *KeygenNode) cleanupSessionStore(id common.ADKGID) {
	node.state.SessionStore.Complete(id)
}

func (node *KeygenNode) BFTDecided(id common.ADKGID) {
	store, complete := node.state.SessionStore.GetOrSetIfNotComplete(id, common.DefaultADKGSession())
	if complete {
		return
	}
	store.Lock()
	defer store.Unlock()
	if store.Over {
		keyIndex, err := id.GetIndex()
		if err != nil {
			return
		}
		curve, err := id.GetCurve()
		if err != nil {
			return
		}
		node.StoreCompletedShare(keyIndex, *store.Share, curve)
		node.StoreCommitment(keyIndex, store.Commitments, curve)
		node.cleanup(id)
	} else {
		store.BFTDecided = true
	}
}

func (node *KeygenNode) DetailsID() common.NodeDetailsID {
	return node.details.ToNodeDetailsID()
}

func (node *KeygenNode) Details() common.KeygenNodeDetails {
	return node.details
}

func (node *KeygenNode) PrivateKey() curves.Scalar {
	return node.privateKey
}

func (node *KeygenNode) CurveParams(curveName string) (curves.Point, curves.Point) {
	return sharing.CurveParams(curveName)
}

func (node *KeygenNode) ID() int {
	return node.details.Index
}

func (node *KeygenNode) Nodes() map[common.NodeDetailsID]common.KeygenNodeDetails {
	return node.CurrentNodes.Nodes
}

func (node *KeygenNode) Send(n common.KeygenNodeDetails, msg common.DKGMessage) error {
	return node.Transport.Send(n, msg)
}

func (node *KeygenNode) Broadcast(msg common.DKGMessage) {
	for _, n := range node.CurrentNodes.Nodes {
		go func(receiver common.KeygenNodeDetails) {
			err := node.Transport.Send(receiver, msg)
			if err != nil {
				log.WithField("Error", err).Error("Node.Broadcast()")
			}
		}(n)
	}
}

func (node *KeygenNode) ReceiveBFTMessage(msg common.DKGMessage) {
	adkgid, err := common.ADKGIDFromRoundID(msg.RoundID)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("DeliverTx:ADKGIDFromRoundID()")
		return
	}

	index, err := adkgid.GetIndex()
	if err != nil {
		return
	}

	c, _ := adkgid.GetCurve()

	if !node.Transport.CheckIfNIZKPProcessed(index, c) {
		err := node.Transport.SendBroadcast(msg)
		if err != nil {
			log.WithError(err).Info("node.ReceiveBFTMessage()")
		}
	}
}

func (node *KeygenNode) PublicKey(index int) curves.Point {
	for _, n := range node.Nodes() {
		if n.Index == index {
			pk, err := curves.K256().NewIdentityPoint().Set(&n.PubKey.X, &n.PubKey.Y)
			if err != nil {
				// log.WithError(err).Error("Node:ReceiveMessage")
				return nil
			}

			return pk
		}
	}
	return nil
}

func (node *KeygenNode) ReceiveMessage(sender common.KeygenNodeDetails, msg common.DKGMessage) {
	err := node.Transport.Receive(sender, msg)
	if err != nil {
		log.WithError(err).Error("Node:ReceiveMessage")
	}
}

func (node *KeygenNode) StoreCompletedShare(keyIndex, si big.Int, c common.CurveName) {
	err := node.broker.DBMethods().StoreCompletedPSSShare(keyIndex, si, si, c)
	if err != nil {
		log.WithError(err).Error("Node:StoreCompletedShare")
	}
}
func (node *KeygenNode) StoreCommitment(keyIndex big.Int, metadata common.ADKGMetadata, c common.CurveName) {
	convertedMetadata := make(map[string][]common.Point)
	for k, v := range metadata.Commitments {
		key := strconv.Itoa(k)
		if _, ok := convertedMetadata[key]; !ok {
			convertedMetadata[key] = []common.Point{}
		}

		for i := 0; i < len(v); i++ {
			val := kcommon.CurvePointToPoint(v[i], c)
			convertedMetadata[key] = append(convertedMetadata[key], val)
		}
	}
	err := node.broker.DBMethods().StoreCommitment(keyIndex, metadata.T, convertedMetadata, c)
	if err != nil {
		log.WithError(err).Error("Node:StoreCommitment")
	}
}

func mapFromNodeList(nodeList []common.KeygenNodeDetails) (res map[common.NodeDetailsID]common.KeygenNodeDetails) {
	res = make(map[common.NodeDetailsID]common.KeygenNodeDetails)
	for _, node := range nodeList {
		res[node.ToNodeDetailsID()] = node
	}
	return
}

func getCommonNodesFromNodeRefArray(nodeRefs []common.NodeReference) (commonNodes []common.KeygenNodeDetails) {
	for _, nodeRef := range nodeRefs {
		commonNodes = append(commonNodes, common.KeygenNodeDetails{
			Index: int(nodeRef.Index.Int64()),
			PubKey: common.Point{
				X: *nodeRef.PublicKey.X,
				Y: *nodeRef.PublicKey.Y,
			},
		})
	}
	return
}

func GetKeygenProtocolPrefix(currEpoch int) common.ProtocolPrefix {
	return common.ProtocolPrefix("keygen" + "-" + strconv.Itoa(currEpoch) + "/")
}

func (node *KeygenNode) State() *common.NodeState {
	return node.state
}

// func (node *KeygenNode) cleanUp(dkgID common.DKGID) error {
// 	node.ShareStore.Complete(dkgID)
// 	return nil
// }

func (node *KeygenNode) ProcessBroadcastMessage(keygenMessage common.DKGMessage) error {
	return nil
}

func (node *KeygenNode) processKeysetMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	switch keygenMessage.Method {
	case keyset.InitMessageType:
		log.Debugf("Got %s", keyset.InitMessageType)
		var msg keyset.InitMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case keyset.ProposeMessageType:
		log.Debugf("Got %s", keyset.ProposeMessageType)
		var msg keyset.ProposeMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case keyset.EchoMessageType:
		log.Debugf("Got %s", keyset.EchoMessageType)
		var msg keyset.EchoMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case keyset.ReadyMessageType:
		log.Debugf("Got %s", keyset.ReadyMessageType)
		var msg keyset.ReadyMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case keyset.OutputMessageType:
		log.Debugf("Got %s", keyset.OutputMessageType)
		var msg keyset.OutputMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	}
}

func (node *KeygenNode) processABAMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	switch keygenMessage.Method {
	case aba.InitMessageType:
		log.Debugf("Got %s", aba.InitMessageType)
		var msg aba.InitMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case aba.Est1MessageType:
		log.Debugf("Got %s", aba.Est1MessageType)
		var msg aba.Est1Message
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case aba.Aux1MessageType:
		log.Debugf("Got %s", aba.Aux1MessageType)
		var msg aba.Aux1Message
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case aba.AuxsetMessageType:
		log.Debugf("Got %s", aba.AuxsetMessageType)
		var msg aba.AuxsetMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case aba.Est2MessageType:
		log.Debugf("Got %s", aba.Est2MessageType)
		var msg aba.Est2Message
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case aba.Aux2MessageType:
		log.Debugf("Got %s", aba.Aux2MessageType)
		var msg aba.Aux2Message
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case aba.CoinInitMessageType:
		log.Infof("Got %s", aba.CoinInitMessageType)
		var msg aba.CoinInitMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case aba.CoinMessageType:
		log.Infof("Got %s", aba.CoinMessageType)
		var msg aba.CoinMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	}
}

func (node *KeygenNode) processKeyDerivationMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	switch keygenMessage.Method {
	case keyderivation.InitMessageType:
		log.Debugf("Got %s", keyderivation.InitMessageType)
		var msg keyderivation.InitMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case keyderivation.ShareMessageType:
		log.Debugf("Got %s", keyderivation.ShareMessageType)
		var msg keyderivation.ShareMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	}
}

func (node *KeygenNode) processACSSMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	log.Debugf("Got %s, round=%s", keygenMessage.Method, keygenMessage.RoundID)
	switch keygenMessage.Method {
	case acss.ShareMessageType:
		var msg acss.ShareMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case acss.ProposeMessageType:
		var msg acss.ProposeMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case acss.EchoMessageType:
		var msg acss.EchoMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case acss.ReadyMessageType:
		var msg acss.ReadyMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case acss.OutputMessageType:
		var msg acss.OutputMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	}
}

func (node *KeygenNode) ProcessMessage(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) error {
	log.WithFields(log.Fields{
		"sender":   sender.Index,
		"receiver": node.ID(),
		"Method":   keygenMessage.Method,
		"RoundID":  keygenMessage.RoundID,
	}).Debug("KeygenNode:ProcessMessage()")

	switch {
	case strings.HasPrefix(keygenMessage.Method, "acss"):
		node.processACSSMessages(sender, keygenMessage)
	case strings.HasPrefix(keygenMessage.Method, "keyset"):
		node.processKeysetMessages(sender, keygenMessage)
	case strings.HasPrefix(keygenMessage.Method, "aba"):
		node.processABAMessages(sender, keygenMessage)
	case strings.HasPrefix(keygenMessage.Method, "key_derivation"):
		node.processKeyDerivationMessages(sender, keygenMessage)
	default:
		log.Infof("No handler found. MsgType=%s", keygenMessage.Method)
		return fmt.Errorf("KeygenMessage method %v not found", keygenMessage.Method)
	}
	return nil
}
