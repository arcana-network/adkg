package aba

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/common"
)

var Est2MessageType common.MessageType = "aba_est2"

type Est2Message struct {
	RoundID common.RoundID
	Kind    common.MessageType
	Curve   common.CurveName
	V       int
	R       int
}

func NewEst2Message(id common.RoundID, v, r int, curve common.CurveName) (*common.DKGMessage, error) {
	m := Est2Message{
		id,
		Est2MessageType,
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

func (m Est2Message) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {
	v, r := m.V, m.R

	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.RoundID, common.DefaultABAStore())
	if complete {
		log.Infof("Keygen already complete: %s", m.RoundID)
		return
	}

	store.Lock()
	defer store.Unlock()

	_, _, f := self.Params(false)

	if Contains(store.Values("est2", r, v), sender.Index) {
		log.Debugf("Got redundant EST2 message from %d", sender.Index)
		return
	}

	store.SetValues("est2", r, v, sender.Index)
	est2Len := len(store.Values("est2", r, v))
	if est2Len >= f+1 && !store.Sent("est2", r, v) {
		store.SetSent("est2", r, v)
		msg, err := NewEst2Message(m.RoundID, v, r, m.Curve)
		if err != nil {
			return
		}
		go self.Broadcast(*msg)
	}

	if est2Len == (2*f)+1 {
		store.SetBin("bin2", r, v)
		bin2 := store.GetBin("bin2", r)
		w := bin2[0]
		msg, err := NewAux2Message(m.RoundID, w, r, m.Curve)
		if err != nil {
			return
		}
		go self.Broadcast(*msg)
	}
}
