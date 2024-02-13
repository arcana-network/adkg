package dpss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/eventbus"
	log "github.com/sirupsen/logrus"
)

type ProtocolPrefix string

type PssNodeTransport struct {
	bus     eventbus.Bus
	broker  *common.MessageBroker
	Prefix  ProtocolPrefix
	PSSNode *PSSNode
}

func NewPssNodeTransport(bus eventbus.Bus, prefix ProtocolPrefix, caller string) *PssNodeTransport {
	transport := PssNodeTransport{
		bus:    bus,
		Prefix: prefix,
		broker: common.NewServiceBroker(bus, caller),
	}
	return &transport
}

func (tp *PssNodeTransport) SetPSSNode(pssNode *PSSNode) {
	tp.PSSNode = pssNode
}

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
func (tp *PssNodeTransport) Receive(senderDetails common.NodeDetails, keygenMessage common.DKGMessage) error {
	log.WithFields(log.Fields{
		"method": common.Stringify(keygenMessage.Method),
	}).Debug("keygen_transport")
	return tp.PSSNode.ProcessMessage(senderDetails, keygenMessage)
}

func (tp *PssNodeTransport) Send(nodeDetails common.NodeDetails, keygenMessage common.DKGMessage) error {
	// TODO implement correctly for DPSS
	// this will be different from Keygen
	return nil
}
