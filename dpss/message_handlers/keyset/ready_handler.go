package keyset

import (
	"bytes"
	"encoding/hex"
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var KeysetReadyMessageType common.DPSSMessageType = "acss_ready"

type KeysetReadyMessage struct {
	RoundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	share   infectious.Share
	hash    []byte
}

func NewKeysetReadyMessage(id common.DPSSRoundID, s infectious.Share, hash []byte, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := &KeysetReadyMessage{
		RoundID: id,
		kind:    KeysetReadyMessageType,
		curve:   curve,
		share:   s,
		hash:    hash,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m *KeysetReadyMessage) Fingerprint() string {
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

func (m *KeysetReadyMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	log.Debugf("Received Ready message from %d on %d", sender.Index, p.ID())
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
	// Make sure the echo received from a node is set to true
	defer func() { keygen.State.ReceivedReady[sender.Index] = true }()

	if keygen.State.Phase == common.Ended {
		return
	}

	receivedReady, found := keygen.State.ReceivedReady[sender.Index]
	if found && receivedReady {
		log.Debugf("Already received ready for %s from %d on %d", m.RoundID, sender.Index, p.ID())
		return
	}

	// Get keygen store by serializing the data of message
	cid := m.Fingerprint()
	c := dpsscommon.GetCStore(keygen, cid)

	keygen.ReadyStore = append(keygen.ReadyStore, m.share)

	// increment the echo messages received
	c.RC = c.RC + 1
	n, k, f := p.Params(false)
	log.Debugf("cid=%v,ready_count=%d, threshold=%d, node=%d", cid, c.RC, k, p.ID())

	if c.RC >= k && !c.ReadySent && c.EC >= k {
		// Broadcast ready message
		msg, err := NewKeysetReadyMessage(m.RoundID, m.share, m.hash, m.curve)
		if err != nil {
			return
		}
		go p.Broadcast(false, *msg)
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

			hash := dpsscommon.Hash(M)
			log.Debugf("HashCompare, hash=%v, mHash=%v", hash, m.hash)

			if bytes.Equal(hash, m.hash) {
				defer func() { keygen.State.Phase = common.Ended }()
				msg, err := NewKeysetOutputMessage(m.RoundID, M, m.curve)
				if err != nil {
					return
				}
				go p.ReceiveMessage(*msg)
				break
			}
		}
	}
}
