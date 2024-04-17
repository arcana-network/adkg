package keyset

import (
	"bytes"
	"encoding/hex"
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"

	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var ReadyMessageType string = "keyset_ready"

type ReadyMessage struct {
	RoundID common.RoundID
	Kind    string
	Curve   common.CurveName
	Share   infectious.Share
	Hash    []byte
}

func NewReadyMessage(id common.RoundID, s infectious.Share, hash []byte, curve common.CurveName) (*common.DKGMessage, error) {
	m := ReadyMessage{
		id,
		ReadyMessageType,
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

func (m ReadyMessage) Fingerprint() string {
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

func (m ReadyMessage) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {
	log.Debugf("Received Ready message from %d on %d", sender.Index, self.ID())
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
		// if keygen is complete, ignore and return
		log.Infof("keygen already complete: %s", m.RoundID)
		return
	}
	keygen.Lock()
	defer keygen.Unlock()

	receivedReady, found := keygen.State.ReceivedReady[sender.Index]
	if found && receivedReady {
		log.Debugf("Already received ready for %s from %d on %d", m.RoundID, sender.Index, self.ID())
		return
	}

	// Make sure the ready received from a node is set to true
	keygen.State.ReceivedReady[sender.Index] = true
	keygen.ReadyStore = append(keygen.ReadyStore, m.Share)

	n, _, f := self.Params()
	log.Debugf("ready_count=%d, threshold=%d, node=%d", len(keygen.ReadyStore), f+1, self.ID())

	if len(keygen.ReadyStore) >= f+1 && !keygen.State.ReadySent {
		echoStore := keygen.FindThresholdEchoStore(f + 1)
		if echoStore != nil {
			// Broadcast ready message
			readyMsg, err := NewReadyMessage(m.RoundID, echoStore.Share, echoStore.Hash, m.Curve)
			if err != nil {
				log.WithField("error", err).Error("NewKeysetProposeMessage")
				return
			}
			keygen.State.ReadySent = true
			go self.Broadcast(*readyMsg)
		}
	}

	if keygen.State.Phase == common.Ended {
		return
	}

	for i := 0; i <= f; i += 1 {
		log.Debugf("len(readstore)=%d, threshold=%d", len(keygen.ReadyStore), (2*f + 1 + i))
		if len(keygen.ReadyStore) >= ((2 * f) + 1 + i) {
			// Create RS encoding
			fec, err := infectious.NewFEC(f+1, n)
			if err != nil {
				log.Debugf("error during creation of fec, err=%s", err)
				return
			}

			M, err := acss.Decode(fec, keygen.ReadyStore)
			if err != nil {
				log.Infof("Decode faced an error, err=%s", err)
				return
			}

			hash := common.HashByte(M)
			log.Debugf("HashCompare, hash=%v, mHash=%v", hash, m.Hash)

			if bytes.Equal(hash, m.Hash) {
				keygen.State.Phase = common.Ended
				outputMsg, err := NewOutputMessage(m.RoundID, M, m.Curve)
				if err != nil {
					log.WithField("error", err).Error("NewKeysetProposeMessage")
					return
				}
				go self.ReceiveMessage(self.Details(), *outputMsg)
				break
			}
		}
	}
}
