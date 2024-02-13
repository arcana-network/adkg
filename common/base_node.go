package common

import (
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

// Node interface represents all the nodes participating in all the protocols,
// namely DPSS and ADKG. This is the base interface.
type Node interface {
	// Returns the ID of the node.
	ID() int
	// Returns the private key of the node.
	PrivateKey() curves.Scalar
	// Returns all the details of the node which are its index and public key.
	Details() KeygenNodeDetails
	// Returns the params of the curve that is bein used by the node in the
	// Respective protocol.
	CurveParams(curveName string) (curves.Point, curves.Point)
	// Sends a message msg to a node n.
	Send(n KeygenNodeDetails, msg DKGMessage) error
	// Receives a message msg from a node sender.
	ReceiveMessage(sender KeygenNodeDetails, msg DKGMessage)
}

// BaseNode has all the attributes that are shared by the nodes in the DPSS and
// ADKG protocol. Both types of nodes will ember the base node to avoid code
// duplication.
type BaseNode struct {
	broker     *MessageBroker    // Broker to communicate the services that the node requires.
	details    KeygenNodeDetails // Details of the node, namely, its index and public key.
	privateKey curves.Scalar     // The private key of the node.
	publicKey  curves.Point      // The public key of the node.
	Transport  *NodeTransport    // Transport layer used by the node to send and receive messages.
}

func NewBaseNode(broker *MessageBroker, details KeygenNodeDetails, privateKey curves.Scalar,
	publicKey curves.Point, transport *NodeTransport) BaseNode {
	newBaseNode := BaseNode{
		broker:     broker,
		details:    details,
		privateKey: privateKey,
		publicKey:  publicKey,
		Transport:  transport,
	}
	return newBaseNode
}

func (n *BaseNode) Details() KeygenNodeDetails {
	return n.details
}

func (node *BaseNode) PrivateKey() curves.Scalar {
	return node.privateKey
}

func (node *BaseNode) CurveParams(curveName string) (curves.Point, curves.Point) {
	return sharing.CurveParams(curveName)
}

func (node *BaseNode) ID() int {
	return node.details.Index
}

func (node *BaseNode) ReceiveMessage(sender KeygenNodeDetails, msg DKGMessage) {
	err := node.Transport.Receive(sender, msg)
	if err != nil {
		log.WithError(err).Error("Node:ReceiveMessage")
	}
}

func (node *BaseNode) Broker() *MessageBroker {
	return node.broker
}
