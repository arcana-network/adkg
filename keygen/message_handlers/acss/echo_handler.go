package acss

import (
	"encoding/hex"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"

	"github.com/arcana-network/dkgnode/common"
)

var EchoMessageType string = "acss_echo"

type EchoMessage struct {
	RoundID common.RoundID
	Kind    string
	Curve   common.CurveName
	Share   infectious.Share
	Hash    []byte
}

func NewAcssEchoMessage(id common.RoundID, s infectious.Share, hash []byte, curve common.CurveName) (*common.DKGMessage, error) {
	m := EchoMessage{
		id,
		EchoMessageType,
		curve,
		s,
		hash,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
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

func (m EchoMessage) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {
	log.Debugf("Echo received: Sender=%s, Receiver=%d", sender.ToNodeDetailsID(), self.ID())
	// Get state from node
	state := self.State().KeygenStore

	// Create empty keygen state
	defaultKeygen := &common.SharingStore{
		RoundID: m.RoundID,
		State: common.RBCState{
			Phase:         common.Initial,
			ReceivedReady: make(map[int]bool),
			ReceivedEcho:  make(map[int]bool),
		},
		EchoStore: make(map[string]*common.EchoStore),
	}

	// Get or set if it doesn't exist
	keygen, complete := state.GetOrSetIfNotComplete(m.RoundID, defaultKeygen)
	if complete {
		log.Debugf("keygen is complete")
		// if keygen is complete, ignore and return
		return
	}

	keygen.Lock()
	defer keygen.Unlock()

	log.WithFields(log.Fields{
		"sender":   sender.Index,
		"self":     self.ID(),
		"roundID":  m.RoundID,
		"received": keygen.State.ReceivedEcho,
	}).Debug("acss_echo_before")

	// Check if the echo has already been received
	receivedEcho, found := keygen.State.ReceivedEcho[sender.Index]
	if receivedEcho && found {
		log.Debugf("Already received echo for %s from %d", m.RoundID, sender.Index)
		return
	}

	// Make sure the echo received from a node is set to true
	keygen.State.ReceivedEcho[sender.Index] = true

	// Get keygen store by serializing the share and hash of the message
	cid := m.Fingerprint()
	echoStore := keygen.GetEchoStore(cid, m.Share, m.Hash)
	echoStore.Count = echoStore.Count + 1

	// increment the echo messages received
	log.WithFields(log.Fields{
		"echoCount": echoStore.Count,
		"sender":    sender.Index,
		"self":      self.ID(),
		"roundID":   m.RoundID,
	}).Debug("acss_echo_after")

	_, _, f := self.Params()

	log.Debugf("node=%d, echo_count=%d, required=%d", self.ID(), echoStore.Count, (2*f + 1))

	// Broadcast ready message if echo count > 2f + 1
	if echoStore.Count >= ((2*f)+1) && !keygen.State.ReadySent {
		// Broadcast Ready Message
		readyMsg, err := NewReadyMessage(m.RoundID, m.Share, m.Hash, m.Curve)
		if err != nil {
			log.Errorf("Could not create ready message: err=%s", err)
			return
		}
		keygen.State.ReadySent = true
		go self.Broadcast(*readyMsg)
	}

	if len(keygen.ReadyStore) >= f+1 && !keygen.State.ReadySent {
		// Find a share with f+1 echo counts, ideally it should have only 1 (key,value) pair
		echoStore := keygen.FindThresholdEchoStore(f + 1)
		if echoStore != nil {
			// Broadcast ready message
			readyMsg, err := NewReadyMessage(m.RoundID, echoStore.Share, echoStore.Hash, m.Curve)
			if err != nil {
				log.Errorf("Could not create ready message: err=%s", err)
				return
			}
			keygen.State.ReadySent = true
			go self.Broadcast(*readyMsg)
		}

	}
}
