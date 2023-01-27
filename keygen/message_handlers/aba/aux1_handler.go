package aba

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
)

var Aux1MessageType string = "aba_aux1"

type Aux1Message struct {
	RoundID common.RoundID
	Kind    string
	Curve   common.CurveName
	V       int
	R       int
}

func NewAux1Message(id common.RoundID, v, r int, curve common.CurveName) (*common.DKGMessage, error) {
	m := Aux1Message{
		id,
		Aux1MessageType,
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

func (m Aux1Message) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {
	v, r := m.V, m.R

	n, _, f := self.Params()
	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.RoundID, common.DefaultABAStore())
	if complete {
		log.Infof("Keygen already complete: %s", m.RoundID)
		return
	}
	store.Lock()
	defer store.Unlock()

	// Check if already present
	if Contains(store.Values("aux", r, v), sender.Index) {
		// log.Infof("Got redundant AUX message from %d", m.Sender())
		return
	}

	//Otherwise, add sender
	store.SetValues("aux", r, v, sender.Index)
	if store.Round() != r {
		return
	}
	view := []int{}
	shouldSendAuxset := false

	bin := store.Bin("bin", r)
	aux0Len := len(store.Values("aux", r, 0))
	aux1Len := len(store.Values("aux", r, 1))

	if Contains(bin, 1) && aux1Len >= n-f {
		view = append(view, 1)
		shouldSendAuxset = true
	} else if Contains(bin, 0) && aux0Len >= n-f {
		view = append(view, 0)
		shouldSendAuxset = true
	} else if aux0Len+aux1Len >= n-f && Contains(bin, 1) && Contains(bin, 0) {
		view = append(view, 0, 1)
		shouldSendAuxset = true
	}

	if !store.Sent("auxset", r, 1) && shouldSendAuxset {
		store.SetSent("auxset", r, 1)
		var auxsetVal = 0
		if len(view) == 1 && view[0] == 1 {
			auxsetVal = 1
		} else if len(view) == 2 {
			auxsetVal = 2
		}

		msg, err := NewAuxsetMessage(m.RoundID, auxsetVal, m.R, m.Curve)
		if err != nil {
			return
		}
		go self.Broadcast(*msg)
	}
}
