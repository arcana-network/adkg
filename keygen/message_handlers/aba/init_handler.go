package aba

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
)

var InitMessageType string = "aba_init"

type InitMessage struct {
	RoundID common.RoundID
	Kind    string
	Curve   common.CurveName
	V       int
	R       int
}

func NewInitMessage(id common.RoundID, v, r int, curve common.CurveName) (*common.DKGMessage, error) {
	m := InitMessage{
		id,
		InitMessageType,
		curve,
		v,
		r,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m *InitMessage) Process(sender common.NodeDetails, self common.DkgParticipant) {
	v, r := m.V, m.R

	if sender.Index != self.ID() {
		return
	}

	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.RoundID, common.DefaultABAStore())
	if complete {
		log.Infof("Keygen already complete: %s", m.RoundID)
		return
	}

	store.Lock()
	defer store.Unlock()

	if store.GetStarted(r) {
		return
	}
	store.SetStarted(r)

	log.Debugf("ABA::Self: %d, Round: %d", self.ID(), r)

	if !store.Sent("est", r, v) {
		store.SetSent("est", r, v)
		msg, err := NewEst1Message(m.RoundID, v, r, m.Curve)
		if err != nil {
			return
		}
		go self.Broadcast(*msg)
	}
}
