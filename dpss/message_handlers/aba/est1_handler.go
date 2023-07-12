package aba

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var Est1MessageType common.DPSSMessageType = "aba_est1"

type Est1Message struct {
	roundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	v       int
	r       int
}

func NewEst1Message(id common.DPSSRoundID, v, r int, curve *curves.Curve) (*common.DPSSMessage, error) {
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

	msg := common.CreateDPSSMessage(m.roundID, m.kind, bytes)
	return &msg, nil
}

func (m *Est1Message) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	v, r := m.v, m.r

	store, complete := p.State().ABAStore.GetOrSetIfNotComplete(m.roundID, common.DefaultABAStore())
	if complete {
		log.Debugf("Keygen already complete: %s", m.roundID)
		return
	}
	store.Lock()
	defer store.Unlock()

	if store.Round() != r {
		return
	}
	// Check if already present
	if Contains(store.Values("est", r, v), sender.Index) {
		log.Debugf("Got redundant EST message from %d", sender.Index)
		return
	}
	//Otherwise, add sender
	store.SetValues("est", r, v, sender.Index)

	_, _, f := p.Params(false)
	estLength := len(store.Values("est", r, v))
	if estLength > f+1 && !store.Sent("est", r, v) {
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

	if estLength > (2*f)+1 {
		store.SetBin("bin", r, v)
		w := store.Bin("bin", r)[0]
		for _, n := range p.Nodes(false) {
			go func(node common.KeygenNodeDetails) {
				msg, err := NewAux1Message(m.roundID, w, r, m.curve)
				if err != nil {
					return
				}
				p.Send(*msg, node)
			}(n)
		}
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
