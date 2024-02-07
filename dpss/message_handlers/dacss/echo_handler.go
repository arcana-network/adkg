package dacss

import (
	"encoding/hex"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"

	"github.com/arcana-network/adkg-proto/common"
)

var AcssEchoMessageType common.MessageType = "dacss_echo"

// Stores the information related to the ECHO message in the RBC protocol.
type AcssEchoMessage struct {
	roundID       common.RoundID
	sender        int
	committeeType int
	kind          common.MessageType
	curve         *curves.Curve
	share         infectious.Share
	hash          []byte // Hash of the shares.
	newCommittee  bool
}

func NewAcssEchoMessage(id common.RoundID, s infectious.Share, hash []byte, curve *curves.Curve, sender int, newCommittee bool) common.DKGMessage {
	m := AcssEchoMessage{
		roundID:      id,
		sender:       sender,
		newCommittee: newCommittee,
		kind:         AcssEchoMessageType,
		curve:        curve,
		share:        s,
		hash:         hash,
	}
	if newCommittee {
		m.committeeType = 1
	} else {
		m.committeeType = 0
	}

	return &m
}

func (m *AcssEchoMessage) Sender() int {
	return m.sender
}

func (m *AcssEchoMessage) Kind() common.MessageType {
	return m.kind
}

func (m *AcssEchoMessage) Fingerprint() string {
	var bytes []byte
	delimiter := common.Delimiter2
	bytes = append(bytes, m.hash...)
	bytes = append(bytes, delimiter...)

	bytes = append(bytes, m.share.Data...)
	bytes = append(bytes, delimiter...)

	bytes = append(bytes, byte(m.share.Number))
	bytes = append(bytes, delimiter...)
	hash := hex.EncodeToString(common.Keccak256(bytes))
	return hash
}

func (m *AcssEchoMessage) Process(p common.DkgParticipant) {
	log.Debugf("Echo received: Sender=%d, Receiver=%d", m.Sender(), p.ID())
	// Get state from node
	state := p.State().KeygenStore

	// Create empty keygen state
	defaultKeygen := &common.SharingStore{
		RoundID: m.roundID,
		State: common.RBCState{
			Phase:         common.Initial,
			ReceivedReady: make(map[int]bool),
			ReceivedEcho:  make(map[int]bool),
		},
		CStore: make(map[string]*common.CStore),
	}

	// Get or set if it doesn't exist
	keygen, complete := state.GetOrSetIfNotComplete(m.roundID, defaultKeygen)
	// log.Debugf("Keygen=%v, complete=%v", keygen, complete)
	if complete {
		// if keygen is complete, ignore and return
		return
	}

	keygen.Lock()
	defer keygen.Unlock()
	// Make sure the echo received from a node is set to true
	defer func() { keygen.State.ReceivedEcho[m.Sender()] = true }()

	// Check if the echo has alreay been received.
	receivedEcho, found := keygen.State.ReceivedEcho[m.Sender()]
	if receivedEcho && found {
		log.Debugf("Already received echo for %s from %d", m.roundID, m.Sender())
		return
	}

	// Get keygen store by serializing the share and hash of the message.
	cid := m.Fingerprint()
	c := common.GetCStore(keygen, cid)

	// increment the echo messages received
	c.EC = c.EC + 1

	// Broadcast ready message if echo count > 2f + 1
	_, _, f := p.Params(m.newCommittee)

	log.Debugf("echo_count=%d, required=%d", c.EC, 2*f+1)
	if c.EC >= (2*f + 1) {
		// Send Ready Message
		c.ReadySent = true
		for _, n := range p.Nodes(m.newCommittee) {
			go func(node common.DkgParticipant) {
				// This corresponds to Line 12, Algorithm 4, RBC Protocol.
				readyMsg := NewAcssReadyMessage(m.roundID, m.share, m.hash, m.curve, p.ID(), m.newCommittee)
				p.Send(readyMsg, node)
			}(n)
		}
	}
}
