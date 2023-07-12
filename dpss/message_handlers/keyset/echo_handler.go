package keyset

import (
	"encoding/hex"
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var KeysetEchoMessageType common.DPSSMessageType = "keyset_echo"

type KeysetEchoMessage struct {
	RoundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	share   infectious.Share
	hash    []byte
}

func NewKeysetEchoMessage(id common.DPSSRoundID, s infectious.Share, hash []byte, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := &KeysetEchoMessage{
		id,
		KeysetEchoMessageType,
		curve,
		s,
		hash,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m *KeysetEchoMessage) Fingerprint() string {
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

func (m *KeysetEchoMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
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
	// log.Debugf("Keygen=%v, complete=%v", keygen, complete)
	if complete {
		// if keygen is complete, ignore and return
		return
	}

	keygen.Lock()
	defer keygen.Unlock()
	// Make sure the echo received from a node is set to true
	defer func() { keygen.State.ReceivedEcho[sender.Index] = true }()

	// Check if the echo has alreay been received
	receivedEcho, found := keygen.State.ReceivedEcho[sender.Index]
	if receivedEcho && found {
		log.Debugf("Already received echo for %s from %d", m.RoundID, sender.Index)
		return
	}

	// Get keygen store by serializing the share and hash of the message
	cid := m.Fingerprint()
	c := dpsscommon.GetCStore(keygen, cid)

	// increment the echo messages received
	c.EC = c.EC + 1

	// Broadcast ready message if echo count > 2f + 1
	_, _, f := p.Params(false)

	log.Debugf("echo_count=%d, required=%d", c.EC, (2*f + 1))
	if c.EC >= (2*f + 1) {
		// Send Ready Message
		c.ReadySent = true
		msg, err := NewKeysetReadyMessage(m.RoundID, m.share, m.hash, m.curve)
		if err != nil {
			return
		}
		p.Broadcast(false, *msg)
	}
}
