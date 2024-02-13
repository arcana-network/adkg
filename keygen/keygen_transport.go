package keygen

import (
	"errors"
	"math/big"

	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/avast/retry-go"
	"github.com/libp2p/go-libp2p/core/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/secp256k1"
)

func (tp *KeygenTransport) Receive(senderDetails common.NodeDetails, keygenMessage common.DKGMessage) error {
	log.WithFields(log.Fields{
		"method": common.Stringify(keygenMessage.Method),
	}).Debug("keygen_transport")
	return tp.KeygenNode.ProcessMessage(senderDetails, keygenMessage)
}

func (tp *KeygenTransport) ReceiveBroadcast(keygenMessage common.DKGMessage) error {
	log.Debug("Received broadcast")
	return tp.KeygenNode.ProcessBroadcastMessage(keygenMessage)
}

func (tp *KeygenTransport) SetKeygenNode(keygenNode *KeygenNode) {
	tp.KeygenNode = keygenNode
}

type KeygenTransport struct {
	bus        eventbus.Bus
	broker     *common.MessageBroker
	Prefix     KeygenProtocolPrefix
	KeygenNode *KeygenNode
}

func NewKeygenTransport(bus eventbus.Bus, prefix KeygenProtocolPrefix) *KeygenTransport {
	transport := KeygenTransport{
		bus:    bus,
		Prefix: prefix,
		broker: common.NewServiceBroker(bus, "keygen-transport"),
	}
	return &transport
}

func (tp *KeygenTransport) SendBroadcast(msg common.DKGMessage) error {
	_, err := tp.broker.TendermintMethods().Broadcast(msg)
	return err
}
func (tp *KeygenTransport) CheckIfNIZKPProcessed(keyIndex big.Int, curve common.CurveName) bool {
	return tp.broker.DBMethods().IndexToPublicKeyExists(keyIndex, curve)
}

func (tp *KeygenTransport) Init() {
	err := tp.broker.P2PMethods().SetStreamHandler(string(tp.Prefix), tp.streamHandler)
	if err != nil {
		log.WithField("protocol prefix", tp.Prefix).WithError(err).Error("could not set stream handler")
	}
}

func (tp *KeygenTransport) streamHandler(streamMessage common.StreamMessage) {
	log.Debug("streamHandler receiving message")

	p2pBasicMsg := streamMessage.Message

	var message common.DKGMessage
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
	nodeReference := tp.broker.ChainMethods().GetNodeDetailsByAddress(common.PointToEthAddress(common.Point(pubKey)))
	index := int(nodeReference.Index.Int64())
	go func(ind int, pubK common.Point, msg common.DKGMessage) {
		err := tp.Receive(common.NodeDetails{
			Index:  ind,
			PubKey: pubK,
		}, msg)
		if err != nil {
			log.WithError(err).Error("error when received message")
			return
		}
	}(index, pubKey, message)
}

func (tp *KeygenTransport) Send(nodeDetails common.NodeDetails, keygenMessage common.DKGMessage) error {
	log.WithFields(log.Fields{
		"to":     common.Stringify(nodeDetails),
		"method": common.Stringify(keygenMessage.Method),
	}).Debug("KeygenTransport:Send()")

	pubKey := tp.KeygenNode.details.PubKey
	if nodeDetails.PubKey.X.Cmp(&pubKey.X) == 0 && nodeDetails.PubKey.Y.Cmp(&pubKey.Y) == 0 {
		return tp.Receive(nodeDetails, keygenMessage)
	}

	// get recipient details
	nodeReference := tp.broker.ChainMethods().GetNodeDetailsByAddress(common.PointToEthAddress(common.Point(nodeDetails.PubKey)))
	byt, err := bijson.Marshal(keygenMessage)
	if err != nil {
		return err
	}
	p2pMsg := tp.broker.P2PMethods().NewP2PMessage(secp256k1.HashToString(byt), false, byt, "transportKeygenMessage")
	log.WithField("P2P connection string", nodeReference.P2PConnection).Debug()
	peerID, err := common.GetPeerIDFromP2pListenAddress(nodeReference.P2PConnection)
	if err != nil {
		return err
	}
	// sign the data
	signature, err := tp.broker.P2PMethods().SignP2PMessage(&p2pMsg)
	if err != nil {
		return errors.New("failed to sign p2p Message" + err.Error())
	}
	p2pMsg.Sign = signature
	err = retry.Do(func() error {
		err := tp.broker.P2PMethods().SendP2PMessage(*peerID, protocol.ID(tp.Prefix), &p2pMsg)
		if err != nil {
			log.WithFields(log.Fields{
				"peerID":     peerID,
				"protocolID": protocol.ID(tp.Prefix),
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
func (tp *KeygenTransport) Sign(s []byte) ([]byte, error) {
	k := tp.broker.ChainMethods().GetSelfPrivateKey()
	return common.ECDSASignBytes(s, &k), nil
}
