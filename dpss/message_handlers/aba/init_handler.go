package aba

import (
	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var InitMessageType string = "aba_init"

type InitMessage struct {
	RoundID common.PSSRoundDetails
	Kind    string
	Curve   common.CurveName
	V       int
	R       int
}

func NewInitMessage(id common.PSSRoundDetails, v, r int, curve common.CurveName) (*common.PSSMessage, error) {
	m := InitMessage{
		id,
		InitMessageType,
		curve,
		v,
		r,
	}
	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m *InitMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	v, r := m.V, m.R

	if sender.Index != self.Details().Index {
		return
	}

	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.RoundID.ToRoundID(), common.DefaultABAStore())
	if complete {
		log.Infof("Keygen already complete: %v", m.RoundID)
		return
	}

	store.Lock()
	defer store.Unlock()

	if store.GetStarted(r) {
		return
	}
	store.SetStarted(r)

	log.Debugf("ABA::Self: %d, Round: %d", self.Details().Index, r)

	if !store.Sent("est", r, v) {
		store.SetSent("est", r, v)
		msg, err := NewEst1Message(m.RoundID, v, r, m.Curve)
		if err != nil {
			return
		}
		go self.Broadcast(false, *msg)
	}
}
