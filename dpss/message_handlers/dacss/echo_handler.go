package dacss

import (
	"encoding/hex"
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var EchoMessageType common.DPSSMessageType = "dacss_echo"

type EchoMessage struct {
	RoundID       common.DPSSRoundID
	committeeType int
	kind          common.DPSSMessageType
	curve         *curves.Curve
	share         infectious.Share
	hash          []byte
	newCommittee  bool
}

func NewEchoMessage(id common.DPSSRoundID, s infectious.Share, hash []byte, curve *curves.Curve, newCommittee bool) (*common.DPSSMessage, error) {
	m := EchoMessage{
		RoundID:      id,
		newCommittee: newCommittee,
		kind:         EchoMessageType,
		curve:        curve,
		share:        s,
		hash:         hash,
	}
	if newCommittee {
		m.committeeType = 1
	} else {
		m.committeeType = 0
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m *EchoMessage) Fingerprint() string {
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

func (m *EchoMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	log.Debugf("Echo received: Sender=%d, Receiver=%d", sender.Index, p.ID())
	// Get state from node
	state := p.State().KeygenStore

	// Create empty keygen state
	defaultKeygen := &dpsscommon.SharingStore{
		RoundID: m.RoundID,
		State: common.RBCState{
			Phase:         common.Initial,
			ReceivedReady: make(map[int]bool),
			ReceivedEcho:  make(map[int]bool),
		},
		CStore: make(map[string]*common.CStore),
	}

	// Get or set if it doesn't exist
	keygen, complete := state.GetOrSetIfNotComplete(m.RoundID, defaultKeygen)
	if complete {
		// if keygen is complete, ignore and return
		return
	}

	keygen.Lock()
	defer keygen.Unlock()

	// Check if the echo has alreay been received
	receivedEcho, found := keygen.State.ReceivedEcho[sender.Index]
	if receivedEcho && found {
		log.Debugf("Already received echo for %s from %d", m.RoundID, sender.Index)
		return
	}

	// Make sure the echo received from a node is set to true
	defer func() { keygen.State.ReceivedEcho[sender.Index] = true }()

	// Get keygen store by serializing the share and hash of the message
	cid := m.Fingerprint()
	c := dpsscommon.GetCStore(keygen, cid)

	// increment the echo messages received
	c.EC = c.EC + 1

	// Broadcast ready message if echo count > 2f + 1

	_, _, f := p.Params(m.newCommittee)

	log.Debugf("echo_count=%d, required=%d", c.EC, 2*f+1)
	if c.EC >= (2*f + 1) {
		// Send Ready Message
		c.ReadySent = true
		for _, n := range p.Nodes(m.newCommittee) {
			go func(node common.KeygenNodeDetails) {
				msg, err := NewReadyMessage(m.RoundID, m.share, m.hash, m.curve, m.newCommittee)
				if err != nil {
					return
				}
				p.Send(*msg, node)
			}(n)
		}
	}
}
