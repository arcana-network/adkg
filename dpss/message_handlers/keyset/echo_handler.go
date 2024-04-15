package keyset

import (
	"encoding/hex"

	"github.com/arcana-network/dkgnode/common"
	"github.com/torusresearch/bijson"

	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var EchoMessageType string = "keyset_echo"

type EchoMessage struct {
	RoundID common.PSSRoundDetails
	Kind    string
	Curve   common.CurveName
	Share   infectious.Share
	Hash    []byte
}

func NewEchoMessage(id common.PSSRoundDetails, s infectious.Share, hash []byte, curve common.CurveName) (*common.PSSMessage, error) {
	m := EchoMessage{
		id,
		EchoMessageType,
		curve,
		s,
		hash,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m EchoMessage) Fingerprint() string {
	var bytes []byte
	delimiter := common.Delimiter2
	bytes = append(bytes, m.Hash...)
	bytes = append(bytes, delimiter...)

	bytes = append(bytes, m.Share.Data...)
	bytes = append(bytes, delimiter...)

	bytes = append(bytes, byte(m.Share.Number))
	bytes = append(bytes, delimiter...)
	hash := hex.EncodeToString(common.Keccak256(bytes))
	return hash
}

func (m EchoMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.Debugf("Echo received: Sender=%d, Receiver=%d", sender.Index, self.Details().Index)
	// Create empty keygen state
	defaultKeygen := &common.KeysetState{
		RoundID: m.RoundID.ToRoundID(),
		RBCState: common.NewRBCState{
			Phase:         common.Initial,
			ReceivedReady: make(map[int]bool),
			ReceivedEcho:  make(map[int]bool),
			EchoStore:     make(map[string]*common.EchoStore),
		},
	}
	// Get or set if it doesn't exist
	log.Debugf("roundID=%v, self=%v", m.RoundID, self.Details().Index)

	// Get state from node
	state, complete := self.State().KeysetStore.GetOrSetIfNotComplete(m.RoundID.ToRoundID(), defaultKeygen)
	if complete {
		// if keygen is complete, ignore and return
		log.Infof("keygen already complete: %v", m.RoundID)
		return
	}

	state.Lock()
	defer state.Unlock()

	// Check if the echo has already been received
	receivedEcho, found := state.RBCState.ReceivedEcho[sender.Index]
	if receivedEcho && found {
		log.Debugf("Already received echo for %v from %d", m.RoundID, sender.Index)
		return
	}

	// Make sure the echo received from a node is set to true
	state.RBCState.ReceivedEcho[sender.Index] = true

	// Get keygen store by serializing the share and hash of the message
	cid := m.Fingerprint()
	echoStore := state.GetEchoStore(cid, m.Share, m.Hash)
	// increment the echo messages received
	echoStore.Count = echoStore.Count + 1

	log.Debugf("Round=%v, EchoCount=%v, self=%v", m.RoundID, echoStore.Count, self.Details().Index)
	_, _, f := self.Params()

	log.Debugf("node=%d, echo_count=%d, required=%d", self.Details().Index, echoStore.Count, (2*f + 1))
	// Broadcast ready message if echo count > 2f + 1
	if echoStore.Count >= ((2*f)+1) && !state.RBCState.ReadySent {
		// Send Ready Message
		readyMsg, err := NewReadyMessage(m.RoundID, m.Share, m.Hash, m.Curve)
		if err != nil {
			log.WithField("error", err).Error("NewKeysetProposeMessage")
			return
		}
		state.RBCState.ReadySent = true
		go self.Broadcast(false, *readyMsg)
	}

	if len(state.RBCState.ReadyStore) >= f+1 && !state.RBCState.ReadySent {
		echoStore := state.FindThresholdEchoStore(f + 1)
		if echoStore != nil {
			// Broadcast ready message
			readyMsg, err := NewReadyMessage(m.RoundID, echoStore.Shard, echoStore.HashMessage, m.Curve)
			if err != nil {
				log.WithField("error", err).Error("NewKeysetProposeMessage")
				return
			}
			state.RBCState.ReadySent = true
			go self.Broadcast(false, *readyMsg)
		}
	}
}
