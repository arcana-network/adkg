package keyset

import (
	"encoding/binary"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/common"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/aba"
)

var OutputMessageType common.MessageType = "keyset_output"

type OutputMessage struct {
	RoundID common.RoundID
	Kind    common.MessageType
	Curve   common.CurveName
	M       []byte
}

func NewOutputMessage(id common.RoundID, data []byte, curve common.CurveName) (*common.DKGMessage, error) {
	m := OutputMessage{
		id,
		OutputMessageType,
		curve,
		data,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m OutputMessage) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {
	// Ignore if not received by self
	if sender.Index != self.ID() {
		return
	}

	adkgid, err := common.ADKGIDFromRoundID(m.RoundID)
	if err != nil {
		return
	}
	leader, err := m.RoundID.Leader()
	if err != nil {
		return
	}

	// create default session to use below
	sessionStore, complete := self.State().SessionStore.GetOrSetIfNotComplete(adkgid, common.DefaultADKGSession())
	if complete {
		// if keygen is complete, ignore and return
		log.Infof("keygen already complete: %s", m.RoundID)
		return
	}

	sessionStore.Lock()
	defer sessionStore.Unlock()

	log.WithFields(log.Fields{
		"adkgid":         adkgid,
		"self":           self.ID(),
		"alreadyStarted": sessionStore.ABAStarted,
	}).Debug("aba_predicate")

	if kcommon.Contains(sessionStore.ABAStarted, int(leader.Int64())) {
		return
	}

	a := binary.BigEndian.Uint64(m.M)
	b := uint64(sessionStore.TPrime)

	vote := 0
	if b&a == a {
		if !sessionStore.ABAComplete {
			vote = 1
		}
	}

	sessionStore.ABAStarted = append(sessionStore.ABAStarted, int(leader.Int64()))
	msg, err := aba.NewInitMessage(m.RoundID, vote, 0, m.Curve)
	if err != nil {
		log.WithError(err).Error("Could not create init message")
		return
	}
	go self.ReceiveMessage(self.Details(), *msg)
}
