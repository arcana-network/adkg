package dacss

import (
	"bytes"
	"encoding/hex"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"

	"github.com/arcana-network/adkg-proto/common"
	"github.com/arcana-network/adkg-proto/common/acss"
	"github.com/arcana-network/adkg-proto/messages"
)

var AcssReadyMessageType common.MessageType = "dacss_ready"

const (
	OldCommittee = iota
	NewCommittee
)

// Stores the information for the READY message in the RBC protocol.
type AcssReadyMessage struct {
	roundID       common.RoundID
	sender        int
	newCommittee  bool
	committeeType int
	kind          common.MessageType
	curve         *curves.Curve
	share         infectious.Share
	hash          []byte
}

func NewAcssReadyMessage(id common.RoundID, s infectious.Share, hash []byte, curve *curves.Curve, sender int, newCommittee bool) common.DKGMessage {
	m := AcssReadyMessage{
		roundID:      id,
		sender:       sender,
		newCommittee: newCommittee,
		kind:         AcssReadyMessageType,
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

func (m *AcssReadyMessage) Sender() int {
	return m.sender
}

func (m *AcssReadyMessage) Kind() common.MessageType {
	return m.kind
}

func (m *AcssReadyMessage) Fingerprint() string {
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

func (m *AcssReadyMessage) Process(p common.DkgParticipant) {
	log.Debugf("Received Ready message from %d on %d", m.Sender(), p.ID())
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
	if complete {
		// If keygen is complete, ignore and return
		return
	}
	keygen.Lock()
	defer keygen.Unlock()

	// Make sure the echo received from a node is set to true
	defer func() { keygen.State.ReceivedReady[m.Sender()] = true }()

	if keygen.State.Phase == common.Ended {
		return
	}

	receivedReady, found := keygen.State.ReceivedReady[m.Sender()]
	if found && receivedReady {
		log.Debugf("Already received ready for %s from %d on %d", m.roundID, m.Sender(), p.ID())
		return
	}

	// Get keygen store by serializing the data of message
	cid := m.Fingerprint()
	c := common.GetCStore(keygen, cid)

	keygen.ReadyStore = append(keygen.ReadyStore, m.share)

	// Increment the echo messages received
	c.RC = c.RC + 1
	n, k, f := p.Params(m.newCommittee)

	log.Debugf("cid=%v,ready_count=%d, threshold=%d, node=%d", cid, c.RC, k, p.ID())

	if c.RC >= k && !c.ReadySent && c.EC >= k {
		// Broadcast ready message
		readyMsg := NewAcssReadyMessage(m.roundID, m.share, m.hash, m.curve, p.ID(), m.newCommittee)
		go p.Broadcast(m.newCommittee, readyMsg)
	}

	for i := 0; i < f; i += 1 {
		log.Debugf("len(readstore)=%d, threshold=%d", len(keygen.ReadyStore), (2*f + 1 + i))
		if len(keygen.ReadyStore) >= (2*f + 1 + i) {
			// Create RS encoding
			f, err := infectious.NewFEC(k, n)
			if err != nil {
				log.Debugf("error during creation of fec, err=%s", err)
				return
			}

			M, err := acss.Decode(f, keygen.ReadyStore)
			if err != nil {
				log.Debugf("Decode faced an error, err=%s", err)
				return
			}
			hash := common.Hash(M)
			log.Debugf("HashCompare, hash=%v, mHash=%v", hash, m.hash)

			if bytes.Equal(hash, m.hash) {
				outputMsg := NewAcssOutputMessage(m.roundID, M, m.curve, p.ID(), "ready", m.newCommittee)
				go p.ReceiveMessage(outputMsg)
				defer func() { keygen.State.Phase = common.Ended }()
				// send to other committee
				msg := messages.MessageData{}
				err := msg.Deserialize(M)
				if err != nil {
					log.Debugf("Could not deserialize message data, err=%s", err)
					return
				}
				for _, n := range p.Nodes(!m.newCommittee) {
					go func(node common.DkgParticipant) {
						readyMsg := NewAcssCommitMessage(m.roundID, msg.Commitments, m.curve, p.ID(), m.newCommittee)
						p.Send(readyMsg, node)
					}(n)
				}
			}

		}
	}

}
