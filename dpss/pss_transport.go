package dpss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/eventbus"
	log "github.com/sirupsen/logrus"
)

type PSSProtocolPrefix string

// Represents the transport used to send and receive messages.
type PssNodeTransport struct {
	bus     eventbus.Bus          // Bus associated to the transport.
	broker  *common.MessageBroker // Broker to communicate multiple services.
	Prefix  PSSProtocolPrefix     // Prefix of the protocol using the transport.
	PSSNode *PSSNode              // TODO: what is this?
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

func (tp *PssNodeTransport) streamHandler(streamMessage common.StreamMessage) {
	// TODO implement correctly for DPSS
	// this will be different from Keygen
}

// TODO check whether this impl is correct
func (tp *PssNodeTransport) Receive(senderDetails common.NodeDetails, msg common.PSSMessage) error {
	log.WithFields(log.Fields{
		"method": common.Stringify(msg.Type),
	}).Debug("pss_transport")
	return tp.PSSNode.ProcessMessage(senderDetails, msg)
}

func (tp *PssNodeTransport) Send(nodeDetails common.NodeDetails, msg common.PSSMessage) error {
	// TODO implement correctly for DPSS
	// this will be different from Keygen
	return nil
}
