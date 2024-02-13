package dacss

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

// old
var n int = 7
var k int = 3
var f int = k - 1

// new
var n1 int = 7
var k1 int = 3
var f1 int = k1 - 1

type Node struct {
	id int

	//oldCommitte
	n int
	k int

	//newCommittee
	n1 int
	k1 int

	transport *MockTransport
	state     *common.NodeState
	keypair   common.KeyPair
	isFaulty  bool

	//the nodes can be in both old and new committee
	isOldCommittee bool
	isNewCommittee bool

	messageCount int

	//shares of old/new committee
	//false: old, true: new
	shares map[bool]map[int64]*big.Int
}

type MockTransport struct {
	nodes               []*Node
	nodeDetails         map[common.NodeDetailsID]common.KeygenNodeDetails
	output              chan string
	broadcastedMessages []common.DKGMessage // Store messages that are broadcasted
	sentMessages        []common.DKGMessage
	receivedMessages    []common.DKGMessage
}

// getting single old/new node
func getSingleNode(isOldCommittee, isNewCommittee bool) (*Node, *MockTransport) {
	nodes := []*Node{}
	keypair := acssc.GenerateKeyPair(curves.K256())
	transport := NewMockTransport(nodes)

	node := NewNode(1, n, k, n1, k1, keypair, transport, false, isOldCommittee, isNewCommittee)
	transport.Init([]*Node{node})
	return node, transport
}

func NewMockTransport(nodes []*Node) *MockTransport {
	return &MockTransport{nodes: nodes, output: make(chan string, 100)}
}

// creates nodes with both isNewCommittee and isOldCommittee as false
// later can be accordingly grouped
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
		node := NewNode(i, n, k, n1, k1, nodeList[i], transport, false, false, false)
		nodes = append(nodes, node)
		i++
	}
	for j := 0; j < faultyCount; j++ {
		log.Infof("Creating faulty node %d", i)
		node := NewNode(i, n, k, n1, k1, nodeList[i], transport, true, false, false)
		nodes = append(nodes, node)
		i++
	}

	transport.Init(nodes)
	return nodes, transport
}

func NewNode(id, n, k, n1, k1 int, keypair common.KeyPair, transport *MockTransport, isFaulty, isNewCommittee, isOldCommittee bool) *Node {
	node := Node{
		id: id,
		n:  n,
		k:  k,
		n1: n1,
		k1: k1,
		state: &common.NodeState{
			KeygenStore:  &common.SharingStoreMap{},
			SessionStore: &common.ADKGSessionStore{},
			ABAStore:     &common.ABAStoreMap{},
		},
		transport:      transport,
		keypair:        keypair,
		isFaulty:       isFaulty,
		isOldCommittee: isOldCommittee,
		isNewCommittee: isNewCommittee,
		shares:         make(map[bool]map[int64]*big.Int),
	}
	return &node
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

// TODO: should we use the DkgParticipant
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

func (node *Node) ReceiveMessage(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	node.transport.receivedMessages = append(node.transport.receivedMessages, keygenMessage) // Save the message
	node.messageCount = node.messageCount + 1
	switch {
	case strings.HasPrefix(keygenMessage.Method, "dacss"):
		node.ProcessDACSSMessages(sender, keygenMessage)
	// case strings.HasPrefix(keygenMessage.Method, "keyset"):
	// 	node.ProcessKeysetMessages(sender, keygenMessage)
	// case strings.HasPrefix(keygenMessage.Method, "aba"):
	// 	node.ProcessABAMessages(sender, keygenMessage)
	// case strings.HasPrefix(keygenMessage.Method, "key_derivation"):
	// 	node.ProcessKeyDerivationMessages(sender, keygenMessage)

	default:
		log.Infof("No handler found. MsgType=%s", keygenMessage.Method)
	}
}

func (node *Node) CountReceivedMessages(msgType string) int {
	receivedMessages := node.transport.receivedMessages
	filteredMessages := make([]common.DKGMessage, 0)

	for _, msg := range receivedMessages {
		if msg.Method == msgType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return len(filteredMessages)
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

// func (node *Node) ProcessKeysetMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
// 	switch keygenMessage.Method {
// 	case keyset.InitMessageType:
// 		log.Debugf("Got %s", keyset.InitMessageType)
// 		var msg keyset.InitMessage
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	case keyset.ProposeMessageType:
// 		log.Debugf("Got %s", keyset.ProposeMessageType)
// 		var msg keyset.ProposeMessage
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	case keyset.EchoMessageType:
// 		log.Debugf("Got %s", keyset.EchoMessageType)
// 		var msg keyset.EchoMessage
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	case keyset.ReadyMessageType:
// 		log.Debugf("Got %s", keyset.ReadyMessageType)
// 		var msg keyset.ReadyMessage
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	case keyset.OutputMessageType:
// 		log.Debugf("Got %s", keyset.OutputMessageType)
// 		var msg keyset.OutputMessage
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	}
// }

// func (node *Node) ProcessABAMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
// 	switch keygenMessage.Method {
// 	case aba.InitMessageType:
// 		log.Debugf("Got %s", aba.InitMessageType)
// 		var msg aba.InitMessage
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	case aba.Est1MessageType:
// 		log.Debugf("Got %s", aba.Est1MessageType)
// 		var msg aba.Est1Message
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	case aba.Aux1MessageType:
// 		log.Debugf("Got %s", aba.Aux1MessageType)
// 		var msg aba.Aux1Message
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	case aba.AuxsetMessageType:
// 		log.Debugf("Got %s", aba.AuxsetMessageType)
// 		var msg aba.AuxsetMessage
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	case aba.Est2MessageType:
// 		log.Debugf("Got %s", aba.Est2MessageType)
// 		var msg aba.Est2Message
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	case aba.Aux2MessageType:
// 		log.Debugf("Got %s", aba.Aux2MessageType)
// 		var msg aba.Aux2Message
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	case aba.CoinInitMessageType:
// 		log.Debugf("Got %s", aba.CoinInitMessageType)
// 		var msg aba.CoinInitMessage
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	case aba.CoinMessageType:
// 		log.Debugf("Got %s", aba.CoinMessageType)
// 		var msg aba.CoinMessage
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	}
// }

// func (node *Node) ProcessKeyDerivationMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
// 	switch keygenMessage.Method {
// 	case keyderivation.InitMessageType:
// 		log.Debugf("Got %s", keyderivation.InitMessageType)
// 		var msg keyderivation.InitMessage
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	case keyderivation.ShareMessageType:
// 		log.Debugf("Got %s", keyderivation.ShareMessageType)
// 		var msg keyderivation.ShareMessage
// 		err := bijson.Unmarshal(keygenMessage.Data, &msg)
// 		if err != nil {
// 			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
// 			return
// 		}
// 		msg.Process(sender, node)
// 	}
// }

func (node *Node) ProcessDACSSMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
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

	case string(AcssOutputMessageType):
		log.Debugf("Got %s", AcssOutputMessageType)
		var msg AcssOutputMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		//TODO: Process not implemented
		// msg.Process(sender, node)
	case ShareMessageType:
		log.Debugf("Got %s", ShareMessageType)
		var msg DacssShareMessage
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

// TODO: needs to be for any committee
func (n *Node) StoreCompletedShare(index big.Int, si big.Int, c common.CurveName) {
	n.shares[true][index.Int64()] = &si
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

func NewDefaultKeygen(round common.RoundDetails) *common.SharingStore {
	return &common.SharingStore{
		RoundID: round.ID(),
		State: common.RBCState{
			Phase:         common.Initial,
			ReceivedReady: make(map[int]bool),
			ReceivedEcho:  make(map[int]bool),
		},
		// EchoStore: make(map[string]*common.EchoStore),
	}
}
