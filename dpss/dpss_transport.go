package dpss

import (
	"errors"
	"math/big"
	"strconv"

	"github.com/arcana-network/dkgnode/common/signature"
	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/avast/retry-go"
	"github.com/libp2p/go-libp2p/core/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/secp256k1"
)

type DPSSTransport struct {
	bus    eventbus.Bus
	Prefix PSSProtocolPrefix
	broker *common.MessageBroker
	node   *PSSNode
}

func (t *DPSSTransport) Receive(sender common.KeygenNodeDetails, msg common.DPSSMessage) error {
	log.WithFields(log.Fields{
		"method": stringify(msg.Method),
	}).Debug("pss_transport")
	return t.node.ProcessMessage(sender, msg)
}

func (t *DPSSTransport) ReceiveBroadcast(msg common.DPSSMessage) error {
	log.Debug("Received pss_broadcast")
	return t.node.ProcessBroadcastMessage(msg)
}

func (t *DPSSTransport) SetNode(node *PSSNode) {
	t.node = node
}

type PSSProtocolPrefix string

func GetPSSProtocolPrefix(oldEpoch int, newEpoch int) PSSProtocolPrefix {
	return PSSProtocolPrefix("dpss" + "-" + strconv.Itoa(oldEpoch) + "-" + strconv.Itoa(newEpoch) + "/")
}

func NewDPSSTransport(bus eventbus.Bus, prefix PSSProtocolPrefix) *DPSSTransport {
	transport := DPSSTransport{
		bus:    bus,
		Prefix: prefix,
		broker: common.NewServiceBroker(bus, "pss_transport"),
	}
	return &transport
}

func (t *DPSSTransport) SendBroadcast(msg common.DPSSMessage) error {
	_, err := t.broker.TendermintMethods().Broadcast(msg)
	return err
}
func (t *DPSSTransport) CheckIfNIZKPProcessed(keyIndex big.Int) bool {
	return t.broker.DBMethods().IndexToPublicKeyExists(keyIndex)
}

func (t *DPSSTransport) Init() {
	err := t.broker.P2PMethods().SetStreamHandler(string(t.Prefix), t.streamHandler)
	if err != nil {
		log.WithField("protocol prefix", t.Prefix).WithError(err).Error("could not set stream handler")
	}
}

func (t *DPSSTransport) streamHandler(streamMessage common.StreamMessage) {
	log.Debug("streamHandler receiving message")

	p2pBasicMsg := streamMessage.Message

	var message common.DPSSMessage
	err := bijson.Unmarshal(p2pBasicMsg.Payload, &message)
	if err != nil {
		log.WithError(err).Error("could not unmarshal payload to keyMessage")
		return
	}
	var pubKey common.Point
	err = bijson.Unmarshal(p2pBasicMsg.GetNodePubKey(), &pubKey)
	if err != nil {
		log.WithError(err).Error("could not unmarshal pubkey")
		return
	}
	nodeReference := t.broker.ChainMethods().GetNodeDetailsByAddress(common.PointToEthAddress(common.Point(pubKey)))
	index := int(nodeReference.Index.Int64())
	go func(ind int, pubK common.Point, msg common.DPSSMessage) {
		err := t.Receive(common.KeygenNodeDetails{
			Index:  ind,
			PubKey: pubK,
		}, msg)
		if err != nil {
			log.WithError(err).Error("error when received message")
			return
		}
	}(index, pubKey, message)
}

func (t *DPSSTransport) Send(node common.KeygenNodeDetails, msg common.DPSSMessage) error {
	log.WithFields(log.Fields{
		"to":     stringify(node),
		"method": stringify(msg.Method),
	}).Debug("KeygenTransport:Send()")

	pubKey := t.node.details.PubKey
	if node.PubKey.X.Cmp(&pubKey.X) == 0 && node.PubKey.Y.Cmp(&pubKey.Y) == 0 {
		return t.Receive(node, msg)
	}

	// get recipient details
	nodeReference := t.broker.ChainMethods().GetNodeDetailsByAddress(common.PointToEthAddress(common.Point(node.PubKey)))
	byt, err := bijson.Marshal(msg)
	if err != nil {
		return err
	}
	p2pMsg := t.broker.P2PMethods().NewP2PMessage(secp256k1.HashToString(byt), false, byt, "transportKeygenMessage")
	log.WithField("P2P connection string", nodeReference.P2PConnection).Debug()
	peerID, err := common.GetPeerIDFromP2pListenAddress(nodeReference.P2PConnection)
	if err != nil {
		return err
	}
	// sign the data
	signature, err := t.broker.P2PMethods().SignP2PMessage(&p2pMsg)
	if err != nil {
		return errors.New("failed to sign p2p Message" + err.Error())
	}
	p2pMsg.Sign = signature
	err = retry.Do(func() error {
		err := t.broker.P2PMethods().SendP2PMessage(*peerID, protocol.ID(t.Prefix), &p2pMsg)
		if err != nil {
			log.WithFields(log.Fields{
				"peerID":     peerID,
				"protocolID": protocol.ID(t.Prefix),
			}).WithError(err).Debug("error when sending p2p message")
			return err
		}
		return nil
	})
	if err != nil {
		log.Error("Could not send the p2p message, failed after retries " + err.Error())
		return err
	}

	return nil
}
func (t *DPSSTransport) Sign(s []byte) ([]byte, error) {
	k := t.broker.ChainMethods().GetSelfPrivateKey()
	return signature.ECDSASignBytes(s, &k), nil
}

func stringify(i interface{}) string {
	bytArr, ok := i.([]byte)
	if ok {
		return string(bytArr)
	}
	str, ok := i.(string)
	if ok {
		return str
	}
	byt, err := bijson.Marshal(i)
	if err != nil {
		log.WithError(err).Error("Could not fastjsonmarshal")
	}
	return string(byt)
}
