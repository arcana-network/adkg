package aba

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"

	log "github.com/sirupsen/logrus"
)

var Est1MessageType string = "aba_est1"

type Est1Message struct {
	RoundID common.PSSRoundDetails
	Kind    string
	Curve   common.CurveName
	V       int
	R       int
}

func NewEst1Message(id common.PSSRoundDetails, v, r int, curve common.CurveName) (*common.PSSMessage, error) {
	m := Est1Message{
		id,
		Est1MessageType,
		curve,
		v,
		r,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m Est1Message) Process(sender common.NodeDetails, self common.PSSParticipant) {
	v, r := m.V, m.R

	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.RoundID.ToRoundID(), common.DefaultABAStore())
	if complete {
		log.Infof("Keygen already complete: %s", m.RoundID)
		return
	}

	store.Lock()
	defer store.Unlock()

	// Check if already present
	if Contains(store.Values("est", r, v), sender.Index) {
		log.Infof("Got redundant EST message from %d, est=%v", sender.Index, store.Values("est", r, v))
		return
	}
	//Otherwise, add sender
	store.SetValues("est", r, v, sender.Index)

	_, _, f := self.Params()
	estLength := len(store.Values("est", r, v))
	log.Debugf("EstCount: %d, required: %d, round: %v", estLength, f+1, m.RoundID)
	if estLength >= f+1 && !store.Sent("est", r, v) {
		msg, err := NewEst1Message(m.RoundID, v, r, m.Curve)
		if err != nil {
			return
		}
		store.SetSent("est", r, v)
		self.Broadcast(false, *msg)
	}

	if estLength == (2*f)+1 {
		store.SetBin("bin", r, v)
		w := store.GetBin("bin", r)[0]
		msg, err := NewAux1Message(m.RoundID, w, r, m.Curve)
		if err != nil {
			return
		}
		self.Broadcast(false, *msg)
	}
}

func Contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
