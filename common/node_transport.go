package common

import (
	"errors"
	"math/big"

	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/arcana-network/dkgnode/secp256k1"
	"github.com/avast/retry-go"
	"github.com/libp2p/go-libp2p/core/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

type ProtocolPrefix string

type NodeProcessor interface {
	NodeDetails() KeygenNodeDetails
	ProcessMessage(senderDetails KeygenNodeDetails, message DKGMessage) error
	ProcessBroadcastMessage(message DKGMessage) error
}

type NodeTransport struct {
	bus        		eventbus.Bus
	broker     		*MessageBroker
	Prefix     		ProtocolPrefix
	NodeProcessor NodeProcessor
}

func NewNodeTransport(bus eventbus.Bus, prefix ProtocolPrefix, caller string) *NodeTransport {
	transport := NodeTransport{
		bus:    bus,
		Prefix: prefix,
		broker: NewServiceBroker(bus, caller),
	}
	return &transport
}

func (tp *NodeTransport) SetNode(node NodeProcessor) {
	tp.NodeProcessor = node
}

// Receive handles incoming "sent" message
func (tp *NodeTransport) Receive(senderDetails KeygenNodeDetails, nodeMessage DKGMessage) error {
	log.WithFields(log.Fields{
		"method": stringify(nodeMessage.Method),
		"prefix": tp.Prefix,
	}).Debug("node_transport")
	return tp.NodeProcessor.ProcessMessage(senderDetails, nodeMessage)
}

// ReceiveBroadcast handles incoming "broadcasted" message
func (tp *NodeTransport) ReceiveBroadcast(nodeMessage DKGMessage) error {
	log.Debug("Received broadcast")
	return tp.NodeProcessor.ProcessBroadcastMessage(nodeMessage)
}

// SendBroadcast sends a message to all nodes
func (tp *NodeTransport) SendBroadcast(msg DKGMessage) error {
	_, err := tp.broker.TendermintMethods().Broadcast(msg)
	return err
}

// TODO what does CheckIfNIZKPProcessed do?
func (tp *NodeTransport) CheckIfNIZKPProcessed(keyIndex big.Int, curve CurveName) bool {
	return tp.broker.DBMethods().IndexToPublicKeyExists(keyIndex, curve)
}

// Init sets up stream for p2p communication, with `streamHandler` as the message handler
func (tp *NodeTransport) Init() {
	err := tp.broker.P2PMethods().SetStreamHandler(string(tp.Prefix), tp.streamHandler)
	if err != nil {
		log.WithField("protocol prefix", tp.Prefix).WithError(err).Error("could not set stream handler")
	}
}

// streamHandler fetches the nodeReference for msg sender and then passes on the message with sender nodeDetails to the node
func (tp *NodeTransport) streamHandler(streamMessage StreamMessage) {
	log.Debug("streamHandler receiving message")

	p2pBasicMsg := streamMessage.Message

	var message DKGMessage
	err := bijson.Unmarshal(p2pBasicMsg.Payload, &message)
	if err != nil {
		log.WithError(err).Error("could not unmarshal payload to keyMessage")
		return
	}
	var pubKey Point
	err = bijson.Unmarshal(p2pBasicMsg.GetNodePubKey(), &pubKey)
	if err != nil {
		log.WithError(err).Error("could not unmarshal pubkey")
		return
	}
	nodeReference := tp.broker.ChainMethods().GetNodeDetailsByAddress(PointToEthAddress(Point(pubKey)))
	index := int(nodeReference.Index.Int64())
	go func(ind int, pubK Point, msg DKGMessage) {
		err := tp.Receive(KeygenNodeDetails{
			Index:  ind,
			PubKey: pubK,
		}, msg)
		if err != nil {
			log.WithError(err).Error("error when received message")
			return
		}
	}(index, pubKey, message)
}

// Send sends a keygenMessage to the specified node
// If message is for node itself, `Receive` will be called.
// Otherwise, it retrieves the recipient details from the broker and creates a P2P message with the keygenMessage.
// The P2P message is then signed and sent to the recipient using the P2PMethods
func (tp *NodeTransport) Send(nodeDetails KeygenNodeDetails, nodeMessage DKGMessage) error {
	log.WithFields(log.Fields{
		"to":     stringify(nodeDetails),
		"method": stringify(nodeMessage.Method),
	}).Debug("KeygenTransport:Send()")

	pubKey := tp.NodeProcessor.NodeDetails().PubKey
	if nodeDetails.PubKey.X.Cmp(&pubKey.X) == 0 && nodeDetails.PubKey.Y.Cmp(&pubKey.Y) == 0 {
		return tp.Receive(nodeDetails, nodeMessage)
	}

	// get recipient details
	nodeReference := tp.broker.ChainMethods().GetNodeDetailsByAddress(PointToEthAddress(Point(nodeDetails.PubKey)))
	byt, err := bijson.Marshal(nodeMessage)
	if err != nil {
		return err
	}

	// TODO FIXME this has a hardcoded msgType "transportKeygenMessage"
	p2pMsg := tp.broker.P2PMethods().NewP2PMessage(secp256k1.HashToString(byt), false, byt, "transportKeygenMessage")
	log.WithField("P2P connection string", nodeReference.P2PConnection).Debug()
	peerID, err := GetPeerIDFromP2pListenAddress(nodeReference.P2PConnection)
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

func (tp *NodeTransport) Sign(s []byte) ([]byte, error) {
	k := tp.broker.ChainMethods().GetSelfPrivateKey()
	return ECDSASignBytes(s, &k), nil
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