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

	auxsetVal := 0
	shouldSendAuxset := false

	bin := store.GetBin("bin", r)
	aux0Len := len(store.Values("aux", r, 0))
	aux1Len := len(store.Values("aux", r, 1))

	if Contains(bin, 1) && aux1Len >= n-f {
		auxsetVal = 1
		shouldSendAuxset = true
	} else if Contains(bin, 0) && aux0Len >= n-f {
		// Not required since default is 0
		auxsetVal = 0
		shouldSendAuxset = true
	} else if aux0Len+aux1Len >= n-f && Contains(bin, 1) && Contains(bin, 0) {
		auxsetVal = 2
		shouldSendAuxset = true
	}

	if !store.Sent("auxset", r, auxsetVal) && shouldSendAuxset {
		msg, err := NewAuxsetMessage(m.RoundID, auxsetVal, m.R, m.Curve)
		if err != nil {
			return
		}
		store.SetSent("auxset", r, auxsetVal)
		self.Broadcast(*msg)
	}
}
