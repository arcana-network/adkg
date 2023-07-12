package aba

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var InitMessageType common.DPSSMessageType = "aba_init"

type InitMessage struct {
	roundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	v       int
	r       int
}

func NewInitMessage(id common.DPSSRoundID, v, r int, curve *curves.Curve) (*common.DPSSMessage, error) {
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

	msg := common.CreateDPSSMessage(m.roundID, m.kind, bytes)
	return &msg, nil
}

func (m *InitMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	v, r := m.v, m.r

	if sender.Index != p.ID() {
		return
	}

	store, complete := p.State().ABAStore.GetOrSetIfNotComplete(m.roundID, common.DefaultABAStore())
	if complete {
		log.Debugf("Keygen already complete: %s", m.roundID)
		return
	}
	store.Lock()
	defer store.Unlock()

	if store.Started() {
		return
	}
	store.SetStarted()

	if !store.Sent("est", r, v) {
		store.SetSent("est", r, v)
		for _, n := range p.Nodes(false) {
			go func(node common.KeygenNodeDetails) {
				msg, err := NewEst1Message(m.roundID, v, r, m.curve)
				if err != nil {
					return
				}
				p.Send(*msg, node)
			}(n)
		}
	}

}
