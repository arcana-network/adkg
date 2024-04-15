package keyset

import (
	"bytes"
	"encoding/hex"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/torusresearch/bijson"

	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var ReadyMessageType string = "keyset_ready"

type ReadyMessage struct {
	RoundID common.PSSRoundDetails
	Kind    string
	Curve   common.CurveName
	Share   infectious.Share
	Hash    []byte
}

func NewReadyMessage(id common.PSSRoundDetails, s infectious.Share, hash []byte, curve common.CurveName) (*common.PSSMessage, error) {
	m := ReadyMessage{
		id,
		ReadyMessageType,
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

func (m ReadyMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.Debugf("Received Ready message from %d on %d", sender.Index, self.Details().Index)
	// Get state from node
	defaultKeygen := &common.KeysetState{
		RoundID: m.RoundID.ToRoundID(),
		RBCState: common.NewRBCState{
			Phase:         common.Initial,
			ReceivedReady: make(map[int]bool),
			ReceivedEcho:  make(map[int]bool),
			EchoStore:     make(map[string]*common.EchoStore),
			ReadyStore:    []infectious.Share{},
		},
	}
	state, complete := self.State().KeysetStore.GetOrSetIfNotComplete(m.RoundID.ToRoundID(), defaultKeygen)
	if complete {
		// if keygen is complete, ignore and return
		log.Infof("keygen already complete: %v", m.RoundID)
		return
	}

	state.Lock()
	defer state.Unlock()

	receivedReady, found := state.RBCState.ReceivedReady[sender.Index]
	if found && receivedReady {
		log.Debugf("Already received ready for %v from %d on %d", m.RoundID, sender.Index, self.Details().Index)
		return
	}

	// Make sure the ready received from a node is set to true
	state.RBCState.ReceivedReady[sender.Index] = true
	state.RBCState.ReadyStore = append(state.RBCState.ReadyStore, m.Share)

	n, _, f := self.Params()
	log.Debugf("ready_count=%d, threshold=%d, node=%d", len(state.RBCState.ReadyStore), f+1, self.Details().Index)

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

	if state.RBCState.Phase == common.Ended {
		return
	}

	for i := range f + 1 {
		log.Infof("len(readstore)=%d, threshold=%d", len(state.RBCState.ReadyStore), (2*f + 1 + i))
		if len(state.RBCState.ReadyStore) >= ((2 * f) + 1 + i) {
			// Create RS encoding
			fec, err := infectious.NewFEC(f+1, n)
			if err != nil {
				log.Debugf("error during creation of fec, err=%s", err)
				return
			}

			M, err := acss.Decode(fec, state.RBCState.ReadyStore)
			if err != nil {
				log.Infof("Decode faced an error, err=%s", err)
				return
			}

			hash := common.HashByte(M)
			log.Debugf("HashCompare, hash=%v, mHash=%v", hash, m.Hash)

			if bytes.Equal(hash, m.Hash) {
				state.RBCState.Phase = common.Ended
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
