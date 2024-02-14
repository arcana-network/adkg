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

// This just has to adhere to the PssParticipant interface
type PssTestNode struct {
	details common.NodeDetails

	n int
	k int

	transport *MockTransport
	state     *common.PSSNodeState
	keypair   common.KeyPair
	isFaulty  bool

	isNewCommittee bool

	messageCount int

	//shares of old/new committee
	//false: old, true: new
	shares map[bool]map[int64]*big.Int
}

type MockTransport struct {
	nodesOld            []*PssTestNode
	nodesNew            []*PssTestNode
	nodeDetails         map[common.NodeDetailsID]common.NodeDetails
	output              chan string
	broadcastedMessages []common.DKGMessage // Store messages that are broadcasted
	sentMessages        []common.DKGMessage
	receivedMessages    []common.DKGMessage
}

// getting single old/new node
func getSingleNode(isNewCommittee bool) (*PssTestNode, *MockTransport) {
	nodesOld := []*PssTestNode{}
	nodesNew := []*PssTestNode{}
	keypair := acssc.GenerateKeyPair(curves.K256())
	transport := NewMockTransport(nodesOld, nodesNew)

	node := NewNode(1, n, k, keypair, transport, false, isNewCommittee)

	if isNewCommittee {
		transport.Init(nodesOld, []*PssTestNode{node})
	} else {
		transport.Init([]*PssTestNode{node}, nodesNew)
	}

	return node, transport
}

func NewMockTransport(nodesOld, nodesNew []*PssTestNode) *MockTransport {
	return &MockTransport{nodesNew: nodesNew, nodesOld: nodesOld, output: make(chan string, 100)}
}

// Create standard old and nodes along with faulty which are all connected in a transport
func setupStandardNodes(countOld, countNew int, faultyCountOld, faultyCountNew int) ([]*PssTestNode, []*PssTestNode, *MockTransport) {
	nodesOld := []*PssTestNode{}
	nodesNew := []*PssTestNode{}

	nodeList := make(map[int]common.KeyPair)

	for i := 1; i <= countOld+countNew+faultyCountOld+faultyCountNew; i++ {
		keypair := acssc.GenerateKeyPair(curves.K256())
		nodeList[i] = keypair
	}
	transport := NewMockTransport(nodesOld, nodesNew)

	log.Info("Creating nodes...")
	// Index of nodes always start at 1
	index := 1

	// Note `NewNodeOld` is called with k=f+1
	for j := 0; j < countOld; j++ {
		log.Infof("Creating node %d", index)
		keypair := acssc.GenerateKeyPair(curves.K256())
		// isNewCommittee and isOldCommittee are set later
		node := NewNode(index, n, k, keypair, transport, false, false)
		nodesOld = append(nodesOld, node)
		index++
	}

	for j := 0; j < faultyCountOld; j++ {
		log.Infof("Creating faulty node %d", index)
		keypair := acssc.GenerateKeyPair(curves.K256())
		node := NewNode(index, n, k, keypair, transport, true, false)
		nodesOld = append(nodesOld, node)
		index++
	}

	// `NewNodeNew`
	for j := 0; j < countNew; j++ {
		log.Infof("Creating node %d", index)
		keypair := acssc.GenerateKeyPair(curves.K256())

		node := NewNode(index, n1, k1, keypair, transport, false, true)
		nodesNew = append(nodesNew, node)
		index++
	}

	for j := 0; j < faultyCountNew; j++ {
		log.Infof("Creating faulty node %d", index)
		keypair := acssc.GenerateKeyPair(curves.K256())
		node := NewNode(index, n1, f1, keypair, transport, true, true)
		nodesNew = append(nodesNew, node)
		index++
	}

	transport.Init(nodesOld, nodesNew)
	return nodesOld, nodesNew, transport
}

func NewNode(index, n, k int, keypair common.KeyPair, transport *MockTransport, isFaulty, isNewCommittee bool) *PssTestNode {
	node := PssTestNode{
		details: common.NodeDetails{Index: index, PubKey: kcommon.CurvePointToPoint(keypair.PublicKey, common.SECP256K1)},
		n:       n,
		k:       k,
		state: &common.PSSNodeState{
			Shares:   &common.PSSShareStoreMap{},
			RbcStore: &common.RBCStateMap{},
		},
		transport:      transport,
		keypair:        keypair,
		isFaulty:       isFaulty,
		isNewCommittee: isNewCommittee,
		shares:         make(map[bool]map[int64]*big.Int),
	}
	return &node
}

func (t *MockTransport) Init(nodesOld, nodesNew []*PssTestNode) {

	t.nodesNew = nodesOld
	t.nodesOld = nodesNew

	nodeDetails := make(map[common.NodeDetailsID]common.NodeDetails)

	for _, node := range nodesOld {
		d := node.Details(false)
		nodeDetails[(&d).ToNodeDetailsID()] = node.Details(false)
	}

	for _, node := range nodesNew {
		d := node.Details(true)
		nodeDetails[(&d).ToNodeDetailsID()] = node.Details(true)
	}

	t.nodeDetails = nodeDetails
}

func (t *MockTransport) Broadcast(sender common.NodeDetails, toNewCommittee bool, m common.DKGMessage) {

	if toNewCommittee {
		for _, p := range t.nodesNew {
			go func(node common.PSSParticipant) {
				node.ReceiveMessage(sender, m)
			}(p)
		}
	} else {
		for _, p := range t.nodesOld {
			go func(node common.PSSParticipant) {
				node.ReceiveMessage(sender, m)
			}(p)
		}
	}

}

// Sends message to the participant
func (t *MockTransport) Send(sender, receiver common.NodeDetails, msg common.DKGMessage) {

	t.sentMessages = append(t.sentMessages, msg) // Save the message

	for _, n := range t.nodesOld {
		log.Debugf("msg=%s, sender=%d, receiver=%d, round=%s", msg.Method, n.ID(), receiver.Index, msg.RoundID)
		if n.details.IsEqual(receiver) {
			go n.ReceiveMessage(sender, msg)
			break
		}
	}

	for _, n := range t.nodesNew {
		log.Debugf("msg=%s, sender=%d, receiver=%d, round=%s", msg.Method, n.ID(), receiver.Index, msg.RoundID)
		if n.details.IsEqual(receiver) {
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

func (node *PssTestNode) ReceiveMessage(sender common.NodeDetails, keygenMessage common.DKGMessage) {
	node.transport.receivedMessages = append(node.transport.receivedMessages, keygenMessage) // Save the message
	node.messageCount = node.messageCount + 1
	switch {
	case strings.HasPrefix(keygenMessage.Method, "dacss"):
		node.ProcessDACSSMessages(sender, keygenMessage)

	default:
		log.Infof("No handler found. MsgType=%s", keygenMessage.Method)
	}
}

func (node *PssTestNode) CountReceivedMessages(msgType string) int {
	receivedMessages := node.transport.receivedMessages
	filteredMessages := make([]common.DKGMessage, 0)

	for _, msg := range receivedMessages {
		if msg.Method == msgType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return len(filteredMessages)
}

func (node *PssTestNode) GetReceivedMessages(msgType string) []common.DKGMessage {
	receivedMessages := node.transport.receivedMessages
	filteredMessages := make([]common.DKGMessage, 0)

	for _, msg := range receivedMessages {
		if msg.Method == msgType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return filteredMessages
}

func (node *PssTestNode) ProcessDACSSMessages(sender common.NodeDetails, keygenMessage common.DKGMessage) {
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

// The index of the node
func (n *PssTestNode) ID() int {
	return n.details.Index
}

func (n *PssTestNode) IsOldNode() bool {
	return !n.isNewCommittee
}

func (n *PssTestNode) IsNewNode() bool {
	return n.isNewCommittee
}

func (n *PssTestNode) AdjustParamN(new_n int) {
	n.n = new_n
}

func (n *PssTestNode) Params(fromNewCommittee bool) (int, int, int) {
	// since the node is created based on new and old
	// therefore, we can directly extract the params
	return n.n, n.k, n.k - 1
}

var c = curves.K256()
var randomScalar = c.Scalar.Random(rand.Reader)

func (n *PssTestNode) CurveParams(c string) (curves.Point, curves.Point) {
	return sharing.CurveParams(c)
}

func (n *PssTestNode) State() *common.PSSNodeState {
	return n.state
}

func (n *PssTestNode) StoreCompletedShare(index big.Int, si big.Int, fromNewCommittee bool, c common.CurveName) {
	n.shares[fromNewCommittee][index.Int64()] = &si
}
func (n *PssTestNode) StoreCommitment(index big.Int, metadata common.ADKGMetadata, c common.CurveName) {
	// n.shares[index.Int64()] = &si
}

func (n *PssTestNode) Broadcast(toNewCommittee bool, m common.DKGMessage) {
	if n.isFaulty {
		log.Debugf("Got Broadcast %s at faulty node %d", m.Method, n.details.Index)
		return
	}
	n.transport.Broadcast(n.Details(toNewCommittee), toNewCommittee, m)
}

func (n *PssTestNode) Send(receiver common.NodeDetails, msg common.DKGMessage) error {

	if n.isFaulty {
		log.Debugf("Got Send %s at faulty node %d", msg.Method, n.details.Index)
		return nil
	}
	//NOTE: Details() does not use the boolean parameter in the function
	n.transport.Send(n.Details(true), receiver, msg)
	return nil
}

func (n *PssTestNode) Nodes(fromNewCommittee bool) map[common.NodeDetailsID]common.NodeDetails {

	return n.transport.nodeDetails
}

func (n *PssTestNode) Details(isNewCommittee bool) common.NodeDetails {

	return n.details
}

func (n *PssTestNode) ReceiveBFTMessage(msg common.DKGMessage) {
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

func (n *PssTestNode) PrivateKey() curves.Scalar {
	return n.keypair.PrivateKey
}

func (n *PssTestNode) PublicKey(idx int, fromNewCommittee bool) curves.Point {

	if fromNewCommittee {
		for _, n := range n.transport.nodesNew {
			if n.ID() == idx {
				return n.keypair.PublicKey
			}
		}
	} else {
		for _, n := range n.transport.nodesOld {
			if n.ID() == idx {
				return n.keypair.PublicKey
			}
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
