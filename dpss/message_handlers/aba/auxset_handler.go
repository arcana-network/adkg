package aba

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/torusresearch/bijson"

	log "github.com/sirupsen/logrus"
)

var AuxsetMessageType string = "aba_auxset"

type AuxsetMessage struct {
	RoundID common.PSSRoundDetails
	Kind    string
	Curve   common.CurveName
	V       int
	R       int
}

func NewAuxsetMessage(id common.PSSRoundDetails, v, r int, curve common.CurveName) (*common.PSSMessage, error) {
	m := AuxsetMessage{
		id,
		AuxsetMessageType,
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

func (m AuxsetMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.RoundID.ToRoundID(), common.DefaultABAStore())
	if complete {
		log.Infof("Keygen already complete: %v", m.RoundID)
		return
	}

	n, _, f := self.Params()

	store.Lock()
	defer store.Unlock()

	if Contains(store.Values("auxset", m.R, m.V), sender.Index) {
		log.Debugf("Got redundant AUXSET message from %d for %v", sender.Index, m.RoundID)
		return
	}

	store.SetValues("auxset", m.R, m.V, sender.Index)

	bin := store.GetBin("bin", m.R)
	auxsetLen0 := len(store.Values("auxset", m.R, 0))
	auxsetLen1 := len(store.Values("auxset", m.R, 1))
	auxsetLen2 := len(store.Values("auxset", m.R, 2))

	var est2 int
	shouldSendEst2 := false
	if Contains(bin, 1) && auxsetLen1 >= n-f {
		est2 = 1
		shouldSendEst2 = true
	} else if Contains(bin, 0) && auxsetLen0 >= n-f {
		est2 = 0
		shouldSendEst2 = true
	} else if auxsetLen0+auxsetLen1+auxsetLen2 >= n-f && Contains(bin, 0) && Contains(bin, 1) {
		est2 = 2
		shouldSendEst2 = true
	}

	if !store.Sent("est2", m.R, est2) && shouldSendEst2 {
		log.Debugf("Node=%d: IN AUXSET_HANDLER: Sending EST2", self.Details().Index)
		store.SetSent("est2", m.R, est2)
		msg, err := NewEst2Message(m.RoundID, est2, m.R, m.Curve)
		if err != nil {
			return
		}
		go self.Broadcast(false, *msg)
	}
}
