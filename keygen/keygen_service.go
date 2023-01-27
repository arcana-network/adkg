package keygen

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/coinbase/kryptology/pkg/core/curves"

	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"

	"github.com/arcana-network/dkgnode/common"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/aba"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/acss"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/keyderivation"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/keyset"
)

type KeygenService struct {
	sync.Mutex
	bus        eventbus.Bus
	broker     *common.MessageBroker
	KeygenNode *KeygenNode
}

func New(bus eventbus.Bus) *KeygenService {
	keygenService := &KeygenService{
		bus:    bus,
		broker: common.NewServiceBroker(bus, common.KEYGEN_SERVICE_NAME),
	}
	return keygenService
}

func (*KeygenService) ID() string {
	return common.KEYGEN_SERVICE_NAME
}

func (service *KeygenService) Start() error {
	ChainMethods := service.broker.ChainMethods()
	selfIndex := ChainMethods.GetSelfIndex()
	selfPubKey := ChainMethods.GetSelfPublicKey()
	currEpoch := ChainMethods.GetCurrentEpoch()
	currNodeList := ChainMethods.AwaitCompleteNodeList(currEpoch)
	currEpochInfo, err := ChainMethods.GetEpochInfo(currEpoch, true)
	if err != nil {
		return err
	}

	selfDetails := common.KeygenNodeDetails{
		Index:  selfIndex,
		PubKey: common.Point{X: selfPubKey.X, Y: selfPubKey.Y},
	}

	k := service.broker.ChainMethods().GetSelfPrivateKey()
	priv, err := curves.K256().NewScalar().SetBigInt(&k)
	if err != nil {
		return err
	}
	keygenNode, err := NewKeygenNode(
		service.broker,
		selfDetails,
		getCommonNodesFromNodeRefArray(currNodeList),
		service.bus,
		int(currEpochInfo.T.Int64()),
		int(currEpochInfo.K.Int64()),
		priv,
	)
	if err != nil {
		return err
	}

	service.KeygenNode = keygenNode
	return nil
}

func (service *KeygenService) Call(method string, args ...interface{}) (interface{}, error) {
	// dBMethods := service.broker.DBMethods()
	switch method {

	case "receive_message":

		var args0 common.DKGMessage
		err := common.CastOrUnmarshal(args[0], &args0)
		if err != nil {
			return nil, err
		}

		log.WithField("keygen_method", args0.Method).Debug("keygen_service_call")

		pubKey := service.KeygenNode.details.PubKey
		index := service.broker.ChainMethods().GetSelfIndex()

		details := common.KeygenNodeDetails{
			PubKey: pubKey,
			Index:  index,
		}

		log.WithFields(log.Fields{
			"index": index,
			"type":  args0.Method,
		}).Info("Broker:ReceiveMessage()")
		return nil, service.KeygenNode.Transport.Receive(details, args0)

	case "cleanup":
		var adkgid common.ADKGID
		_ = common.CastOrUnmarshal(args[0], &adkgid)
		go service.KeygenNode.Cleanup(adkgid)
		return nil, nil
	}
	return nil, fmt.Errorf("keygen service method %v not found", method)
}

func (service *KeygenService) Stop() error {
	log.Info("Stopping keygen service")
	return nil
}
func (service *KeygenService) IsRunning() bool {
	return true
}

type KeygenProtocolPrefix string

type KeygenNode struct {
	broker       *common.MessageBroker
	details      common.KeygenNodeDetails
	CurrentNodes common.NodeNetwork
	Transport    *KeygenTransport

	// New stores here
	state      *common.NodeState
	privateKey curves.Scalar
	publicKey  curves.Point
}

func (node *KeygenNode) Params() (n, k, t int) {
	n = node.CurrentNodes.N
	k = node.CurrentNodes.K
	t = node.CurrentNodes.T
	return
}

func (node *KeygenNode) Cleanup(id common.ADKGID) {
	node.cleanupKeygenStore(id)
	node.cleanupABAStore(id)
	node.cleanupADKGSessionStore(id)
}

func (node *KeygenNode) cleanupKeygenStore(id common.ADKGID) {
	for _, n := range node.CurrentNodes.Nodes {
		node.state.KeygenStore.Complete((&common.RoundDetails{
			ADKGID: id,
			Dealer: n.Index,
			Kind:   "acss",
		}).ID())
		node.state.KeygenStore.Complete((&common.RoundDetails{
			ADKGID: id,
			Dealer: n.Index,
			Kind:   "keyset",
		}).ID())
	}
}
func (node *KeygenNode) cleanupABAStore(id common.ADKGID) {
	for _, n := range node.CurrentNodes.Nodes {
		node.state.ABAStore.Complete((&common.RoundDetails{
			ADKGID: id,
			Dealer: n.Index,
			Kind:   "keyset",
		}).ID())
	}
}
func (node *KeygenNode) cleanupADKGSessionStore(id common.ADKGID) {
	node.state.SessionStore.Complete(id)
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

func getFixedScalar(c *curves.Curve) (curves.Scalar, error) {
	k256Scalar := "6c47fa13c92d8b47d1579f112657c22ddd0c3a6ed1fb56c8fc80a086477bf89c"
	ed25519Scalar := "19d7725aab29dab57a2124400cb2ca69c9830f691104d1471b8cb0759cd17d1"

	if c.Name == curves.K256Name {
		b2, ok := new(big.Int).SetString(k256Scalar, 16)
		if !ok {
			return nil, fmt.Errorf("invalid hex for scalar for curve %s ", c.Name)
		}
		s2, err := c.Scalar.SetBigInt(b2)
		return s2, err
	} else if c.Name == curves.ED25519Name {
		b2, ok := new(big.Int).SetString(ed25519Scalar, 16)
		if !ok {
			return nil, fmt.Errorf("invalid hex for scalar for curve %s ", c.Name)
		}
		s2, err := c.Scalar.SetBigInt(b2)
		return s2, err
	} else {
		return nil, fmt.Errorf("Invalid curve")
	}
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

	if !node.Transport.CheckIfNIZKPProcessed(index) {
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

func (node *KeygenNode) StoreCompletedShare(keyIndex big.Int, si big.Int) {
	err := node.broker.DBMethods().StoreCompletedPSSShare(keyIndex, si, si)
	if err != nil {
		log.WithError(err).Error("Node:StoreCompletedShare")
	}
}
func (node *KeygenNode) StoreCommitment(keyIndex big.Int, metadata common.ADKGMetadata) {
	convertedMetadata := make(map[string][]common.Point)
	for k, v := range metadata.Commitments {
		key := strconv.Itoa(k)
		if _, ok := convertedMetadata[key]; !ok {
			convertedMetadata[key] = []common.Point{}
		}

		for i := 0; i < len(v); i++ {
			val := kcommon.CurvePointToPoint(v[i])
			convertedMetadata[key] = append(convertedMetadata[key], val)
		}
	}
	err := node.broker.DBMethods().StoreCommitment(keyIndex, metadata.T, convertedMetadata)
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

func NewKeygenNode(broker *common.MessageBroker, nodeDetails common.KeygenNodeDetails,
	nodeList []common.KeygenNodeDetails, bus eventbus.Bus, T int, K int,
	privateKey curves.Scalar) (*KeygenNode, error) {
	transport := NewKeygenTransport(bus, GetKeygenProtocolPrefix(1))
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
		Transport:    transport,
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
	transport.SetKeygenNode(newKeygenNode)
	return newKeygenNode, nil
}

func GetKeygenProtocolPrefix(currEpoch int) KeygenProtocolPrefix {
	return KeygenProtocolPrefix("keygen" + "-" + strconv.Itoa(currEpoch) + "/")
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

func (node *KeygenNode) ProcessKeysetMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
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

func (node *KeygenNode) ProcessABAMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
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

func (node *KeygenNode) ProcessKeyDerivationMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
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

func (node *KeygenNode) ProcessACSSMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	switch keygenMessage.Method {
	case acss.ShareMessageType:
		log.Debugf("Got %s", acss.ShareMessageType)
		var msg acss.ShareMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)

	case acss.ProposeMessageType:
		log.Debugf("Got %s", acss.ProposeMessageType)
		var msg acss.ProposeMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case acss.EchoMessageType:
		log.Debugf("Got %s", acss.EchoMessageType)
		var msg acss.EchoMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case acss.ReadyMessageType:
		log.Debugf("Got %s", acss.ReadyMessageType)
		var msg acss.ReadyMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case acss.OutputMessageType:
		log.Debugf("Got %s", acss.OutputMessageType)
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
		node.ProcessACSSMessages(sender, keygenMessage)
	case strings.HasPrefix(keygenMessage.Method, "keyset"):
		node.ProcessKeysetMessages(sender, keygenMessage)
	case strings.HasPrefix(keygenMessage.Method, "aba"):
		node.ProcessABAMessages(sender, keygenMessage)
	case strings.HasPrefix(keygenMessage.Method, "key_derivation"):
		node.ProcessKeyDerivationMessages(sender, keygenMessage)

	default:
		log.Infof("No handler found. MsgType=%s", keygenMessage.Method)
		return fmt.Errorf("KeygenMessage method %v not found", keygenMessage.Method)
	}
	return nil
}
