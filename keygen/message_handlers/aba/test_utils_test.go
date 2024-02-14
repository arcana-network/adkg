package aba

import (
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	acssc "github.com/arcana-network/dkgnode/keygen/common/acss"

	"github.com/arcana-network/dkgnode/keygen/message_handlers/keyderivation"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"

	"github.com/torusresearch/bijson"
)

// TODO not hardcode this
var n int = 7
var f int = 3

// TODO cleanup

func getSingleNode() (*Node, *MockTransport) {
	nodes := []*Node{}
	keypair := acssc.GenerateKeyPair(curves.K256())
	transport := NewMockTransport(nodes)
	node := NewNode(1, n, f, keypair, transport, false)
	transport.Init([]*Node{node})
	return node, transport
}

func setupNodes(count int, faultyCount int) ([]*Node, *MockTransport) {
	nodes := []*Node{}
	nodeList := make(map[int]common.KeyPair)
	for i := 1; i <= count+faultyCount; i++ {
		keypair := acssc.GenerateKeyPair(curves.K256())
		nodeList[i] = keypair
	}
	transport := NewMockTransport(nodes)

	log.Info("Creating nodes...")
	i := 1

	// Note `NewNode` is called with k=f+1
	for j := 0; j < count; j++ {
		log.Infof("Creating node %d", i)
		node := NewNode(i, n, f+1, nodeList[i], transport, false)
		nodes = append(nodes, node)
		i++
	}
	for j := 0; j < faultyCount; j++ {
		log.Infof("Creating faulty node %d", i)
		node := NewNode(i, n, f+1, nodeList[i], transport, true)
		nodes = append(nodes, node)
		i++
	}

	transport.Init(nodes)
	return nodes, transport
}

type MockTransport struct {
	nodes               []*Node
	nodeDetails         map[common.NodeDetailsID]common.KeygenNodeDetails
	output              chan string
	broadcastedMessages []common.DKGMessage // Store messages that are broadcasted
	sentMessages        []common.DKGMessage
	receivedMessages    []common.DKGMessage
}

func NewMockTransport(nodes []*Node) *MockTransport {
	return &MockTransport{nodes: nodes, output: make(chan string, 100)}
}

func (t *MockTransport) Init(nodes []*Node) {
	t.nodes = nodes
	nodeDetails := make(map[common.NodeDetailsID]common.KeygenNodeDetails)

	for _, node := range nodes {
		d := node.Details()
		nodeDetails[(&d).ToNodeDetailsID()] = node.Details()
	}
	t.nodeDetails = nodeDetails
}

func (t *MockTransport) Broadcast(sender common.KeygenNodeDetails, m common.DKGMessage) {
	t.broadcastedMessages = append(t.broadcastedMessages, m) // Save the message
	for _, p := range t.nodes {
		go func(node common.DkgParticipant) {
			node.ReceiveMessage(sender, m)
		}(p)
	}
}

// Method to retrieve broadcasted messages for assertions
func (t *MockTransport) GetBroadcastedMessages() []common.DKGMessage {
	return t.broadcastedMessages
}

// Sends message to the participant
func (t *MockTransport) Send(sender, receiver common.KeygenNodeDetails, msg common.DKGMessage) {
	// time.Sleep(500 * time.Millisecond)
	t.sentMessages = append(t.sentMessages, msg) // Save the message

	for _, n := range t.nodes {
		log.Debugf("msg=%s, sender=%d, receiver=%d, round=%s", msg.Method, n.ID(), receiver.Index, msg.RoundID)
		if n.ID() == receiver.Index {
			go n.ReceiveMessage(sender, msg)
			break
		}
	}
}

func (t *MockTransport) GetSentMessages() []common.DKGMessage {
	return t.sentMessages
}

type KeyMap struct {
	shares map[int]*big.Int
}

type Node struct {
	id           int
	n            int
	k            int
	transport    *MockTransport
	state        *common.NodeState
	keypair      common.KeyPair
	isFaulty     bool
	messageCount int
	shares       map[int64]*big.Int
}

func (node *Node) ReceiveMessage(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	node.messageCount = node.messageCount + 1
	switch {
	case strings.HasPrefix(keygenMessage.Method, "aba"):
		node.ProcessABAMessages(sender, keygenMessage)
	case strings.HasPrefix(keygenMessage.Method, "key_derivation"):
		node.ProcessKeyDerivationMessages(sender, keygenMessage)

	default:
		log.Infof("No handler found. MsgType=%s", keygenMessage.Method)
		// return fmt.Errorf("KeygenMessage method %v not found", keygenMessage.Method)
	}
}

func (node *Node) ProcessABAMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	switch keygenMessage.Method {
	case InitMessageType:
		log.Debugf("Got %s", InitMessageType)
		var msg InitMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case Est1MessageType:
		log.Debugf("Got %s", Est1MessageType)
		var msg Est1Message
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case Aux1MessageType:
		log.Debugf("Got %s", Aux1MessageType)
		var msg Aux1Message
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case AuxsetMessageType:
		log.Debugf("Got %s", AuxsetMessageType)
		var msg AuxsetMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case Est2MessageType:
		log.Debugf("Got %s", Est2MessageType)
		var msg Est2Message
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case Aux2MessageType:
		log.Debugf("Got %s", Aux2MessageType)
		var msg Aux2Message
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case CoinInitMessageType:
		log.Debugf("Got %s", CoinInitMessageType)
		var msg CoinInitMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case CoinMessageType:
		log.Debugf("Got %s", CoinMessageType)
		var msg CoinMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	}
}

func (node *Node) ProcessKeyDerivationMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
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

func (n *Node) ID() int {
	return n.id
}

func (n *Node) AdjustParamN(new_n int) {
	n.n = new_n
}

func (n *Node) Params() (int, int, int) {
	return n.n, n.k, n.k - 1
}

var c = curves.K256()
var randomScalar = c.Scalar.Random(rand.Reader)

func (n *Node) CurveParams(c string) (curves.Point, curves.Point) {
	return sharing.CurveParams(c)
}

func (n *Node) State() *common.NodeState {
	return n.state
}

func (n *Node) Cleanup(id common.ADKGID) {
	n.cleanupKeygenStore(id)
	n.cleanupABAStore(id)
	n.cleanupADKGSessionStore(id)
	// debug.FreeOSMemory()
}

func (node *Node) cleanupKeygenStore(id common.ADKGID) {
	for _, n := range node.Nodes() {
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
func (node *Node) cleanupABAStore(id common.ADKGID) {
	for _, n := range node.Nodes() {
		node.state.ABAStore.Complete((&common.RoundDetails{
			ADKGID: id,
			Dealer: n.Index,
			Kind:   "keyset",
		}).ID())
	}
}
func (node *Node) cleanupADKGSessionStore(id common.ADKGID) {
	node.state.SessionStore.Complete(id)
}

func (n *Node) StoreCompletedShare(index big.Int, si big.Int, c common.CurveName) {
	n.shares[index.Int64()] = &si
}
func (n *Node) StoreCommitment(index big.Int, metadata common.ADKGMetadata, c common.CurveName) {
	// n.shares[index.Int64()] = &si
}

func (n *Node) Broadcast(m common.DKGMessage) {
	if n.isFaulty {
		log.Debugf("Got Broadcast %s at faulty node %d", m.Method, n.id)
		return
	}
	n.transport.Broadcast(n.Details(), m)
}

func (n *Node) Send(receiver common.KeygenNodeDetails, msg common.DKGMessage) error {
	if n.isFaulty {
		log.Debugf("Got Send %s at faulty node %d", msg.Method, n.id)
		return nil
	}
	n.transport.Send(n.Details(), receiver, msg)
	return nil
}

func (n *Node) Nodes() map[common.NodeDetailsID]common.KeygenNodeDetails {
	return n.transport.nodeDetails
}

func (n *Node) Details() common.KeygenNodeDetails {
	return common.KeygenNodeDetails{
		Index:  n.id,
		PubKey: kcommon.CurvePointToPoint(n.keypair.PublicKey, common.SECP256K1),
	}
}

func (n *Node) ReceiveBFTMessage(msg common.DKGMessage) {
	if msg.Method == keyderivation.PubKeygenType {
		var m keyderivation.PubKeygenMessage
		if err := bijson.Unmarshal(msg.Data, &m); err != nil {
			log.WithError(err).Infof("ReceiveBFTMessage()")
			return
		}
		adkgid, _ := common.ADKGIDFromRoundID(m.RoundID)
		log.Debugf("ADKGID=%s", adkgid)
		res := m.PublicKey.X.Text(16) + m.PublicKey.Y.Text(16)
		go func() { n.transport.output <- res }()
	}
}

func (n *Node) PrivateKey() curves.Scalar {
	return n.keypair.PrivateKey
}

func (n *Node) PublicKey(index int) curves.Point {
	for _, n := range n.transport.nodes {
		if n.ID() == index {
			return n.keypair.PublicKey
		}
	}
	c := curves.K256()
	return c.Point.Identity()
}

func NewNode(id, n, k int, keypair common.KeyPair, transport *MockTransport, isFaulty bool) *Node {
	node := Node{
		id: id,
		n:  n,
		k:  k,
		state: &common.NodeState{
			KeygenStore:  &common.SharingStoreMap{},
			SessionStore: &common.ADKGSessionStore{},
			ABAStore:     &common.ABAStoreMap{},
		},
		transport: transport,
		keypair:   keypair,
		isFaulty:  isFaulty,
		shares:    make(map[int64]*big.Int),
	}
	return &node
}

func (node *Node) GetReceivedMessages(msgType string) []common.DKGMessage {
	receivedMessages := node.transport.receivedMessages
	filteredMessages := make([]common.DKGMessage, 0)

	for _, msg := range receivedMessages {
		if msg.Method == msgType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return filteredMessages
}
