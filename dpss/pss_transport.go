package dpss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/libp2p/go-libp2p/core/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

type PSSProtocolPrefix string

// Represents the transport used to send and receive messages.
type PssNodeTransport struct {
	bus     eventbus.Bus          // Bus associated to the transport.
	broker  *common.MessageBroker // Broker to communicate multiple services.
	Prefix  PSSProtocolPrefix     // Prefix of the protocol using the transport.
	PSSNode *PSSNode              // PSS node associated to the transport.
}

// Creates a new PSS transport.
func NewPssNodeTransport(bus eventbus.Bus, prefix PSSProtocolPrefix, caller string) *PssNodeTransport {
	transport := PssNodeTransport{
		bus:    bus,
		Prefix: prefix,
		broker: common.NewServiceBroker(bus, caller),
	}
	return &transport
}

// SetPSSNode sets the PSS node in the transport.
func (tp *PssNodeTransport) SetPSSNode(pssNode *PSSNode) {
	tp.PSSNode = pssNode
}

// Init initializes the transport.
func (tp *PssNodeTransport) Init() {
	err := tp.broker.P2PMethods().SetStreamHandler(string(tp.Prefix), tp.streamHandler)
	if err != nil {
		log.WithField("protocol prefix", tp.Prefix).WithError(err).Error("could not set stream handler")
	}
}

// incoming messages are passed onto the Receive method of the PSSNode
func (tp *PssNodeTransport) streamHandler(streamMessage common.StreamMessage) {
	log.Debug("PSS Transport streamHandler receiving message")

	p2pBasicMsg := streamMessage.Message

	var message common.PSSMessage
	err := bijson.Unmarshal(p2pBasicMsg.Payload, &message)
	if err != nil {
		log.WithError(err).Error("could not unmarshal payload to PSSMessage")
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
	go func(ind int, pubK common.Point, msg common.PSSMessage) {
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

// Receive receives a PSSMessage from a sender.
func (tp *PssNodeTransport) Receive(senderDetails common.NodeDetails, msg common.PSSMessage) error {
	log.WithFields(log.Fields{
		"method": common.Stringify(msg.Type),
	}).Debug("pss_transport")
	return tp.PSSNode.ProcessMessage(senderDetails, msg)
}

// Send sends a PSSMessage to a node over p2p network.
func (tp *PssNodeTransport) Send(nodeDetails common.NodeDetails, msg common.PSSMessage) error {
	log.WithFields(log.Fields{
		"to":     common.Stringify(nodeDetails),
		"method": common.Stringify(msg.Type),
	}).Debug("PSSTransport:Send()")

	pubKey := tp.PSSNode.Details().PubKey
	msgType := "transportPSSMessage"
	broker := tp.broker
	protocolId := protocol.ID(tp.Prefix)
	if nodeDetails.PubKey.X.Cmp(&pubKey.X) == 0 && nodeDetails.PubKey.Y.Cmp(&pubKey.Y) == 0 {
		return tp.Receive(nodeDetails, msg)
	}
	byt, err := bijson.Marshal(msg)
	if err != nil {
		return err
	}

	return common.SendForMessageType(broker, nodeDetails, byt, msgType, protocolId)
}
