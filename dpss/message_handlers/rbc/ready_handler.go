package rbc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var AcssReadyMessageType common.MessageType = "dacss_ready"

const (
	OldCommittee = iota
	NewCommittee
)

// Stores the information for the READY message in the RBC protocol.
type RbcReadyMessage struct {
	RoundID       common.RoundID
	NewCommittee  bool
	CommitteeType int
	Kind          common.MessageType
	Curve         *curves.Curve
	Share         infectious.Share
	Hash          []byte
	ProtoOrigin   string
}

func NewRbcReadyMessage(id common.RoundID, s infectious.Share, hash []byte, curve *curves.Curve, sender int, newCommittee bool, protoOrigin string) (*common.DKGMessage, error) {
	m := RbcReadyMessage{
		RoundID:      id,
		NewCommittee: newCommittee,
		Kind:         AcssReadyMessageType,
		Curve:        curve,
		Share:        s,
		Hash:         hash,
		ProtoOrigin:  protoOrigin,
	}
	if newCommittee {
		m.CommitteeType = 1
	} else {
		m.CommitteeType = 0
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m *RbcReadyMessage) Fingerprint() string {
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

func (m RbcReadyMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.Debugf("Received Ready message from %d on %d", sender, self.ID())
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
		CStore: make(map[string]*common.CStore),
	}

	// Get or set if it doesn't exist
	keygen, complete := state.GetOrSetIfNotComplete(m.RoundID, defaultKeygen)
	if complete {
		// If keygen is complete, ignore and return
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
		log.Debugf("Already received ready for %s from %d on %d", m.RoundID, sender, self.ID())
		return
	}

	// Get keygen store by serializing the data of message
	cid := m.Fingerprint()
	c := common.GetCStore(keygen, cid)

	keygen.ReadyStore = append(keygen.ReadyStore, m.Share)

	// Increment the echo messages received
	c.RC = c.RC + 1
	n, k, f := self.Params()

	log.Debugf("cid=%v,ready_count=%d, threshold=%d, node=%d", cid, c.RC, k, self.ID())

	if c.RC >= k && !c.ReadySent && c.EC >= k {
		// Broadcast ready message
		readyMsg, err := NewRbcReadyMessage(m.RoundID, m.Share, m.Hash, m.Curve, self.ID(), m.NewCommittee, m.ProtoOrigin)
		if err != nil {
			return
		}
		go self.Broadcast(m.NewCommittee, readyMsg)
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
			hash := common.HashByte(M)
			log.Debugf("HashCompare, hash=%v, mHash=%v", hash, m.Hash)

			if bytes.Equal(hash, m.Hash) {
				// The output of the RBC occurs here.
				// Store the message in the RBCState.
				// TODO: Is this making conflict with other Keygen processes
				// that are using also RBC?
				store, _ := self.State().KeygenStore.Get(m.RoundID)
				store.State.ReceivedMessage = M
				store.State.OutputReceived = true

				// TODO: Check this possible solution: =============
				defer func() { keygen.State.Phase = common.Ended }()

				// Send RBC Output message to NodeTransport
				// On NodeTransport it has to be picked up by the correct msg handler (probably xxxOutputHandler)

				// if m.ProtoOrigin == "dacss" {
				// 	outputMsg := dacss.NewDacssOutputMessage(m.RoundID, M, m.Curve, self.ID(), "ready", m.NewCommittee)
				// 	go self.ReceiveMessage(self.Details(), outputMsg)
				// } else if m.ProtoOrigin == "acss" {
				// 	outputMsg, err := commonacss.NewOutputMessage(m.RoundID, M, common.CurveName(m.Curve.Name))
				// 	if err != nil {
				// 		return
				// 	}
				// 	go self.ReceiveMessage(self.Details(), *outputMsg)
				// }
				// =================================================

				// TODO: we have the message stored, what should we do with the
				// code below?

				// send to other committee
				msg := messages.MessageData{}
				err := msg.Deserialize(M)
				if err != nil {
					log.Debugf("Could not deserialize message data, err=%s", err)
					return
				}
				for _, n := range self.Nodes(!m.NewCommittee) {
					go func(node common.NodeDetails) {
						readyMsg := dacss.NewDacssCommitMessage(m.RoundID, msg.Commitments, m.Curve, self.ID(), m.NewCommittee)
						self.Send(node, readyMsg)
					}(n)
				}
			}

		}
	}
}
