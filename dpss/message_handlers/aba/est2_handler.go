package aba

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var Est2MessageType common.DPSSMessageType = "est2_aba"

type Est2Message struct {
	roundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	v       int
	r       int
}

func NewEst2Message(id common.DPSSRoundID, v, r int, curve *curves.Curve, sender int) (*common.DPSSMessage, error) {
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

	msg := common.CreateDPSSMessage(m.roundID, m.kind, bytes)
	return &msg, nil
}

func (m *Est2Message) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
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
	_, _, f := p.Params(false)

	if Contains(store.Values("est2", r, v), sender.Index) {
		log.Debugf("Got redundant EST2 message from %d", sender.Index)
		return
	}

	store.SetValues("est2", r, v, sender.Index)
	est2Len := len(store.Values("est2", r, v))
	if est2Len > f && !store.Sent("est2", r, v) {
		store.SetSent("est2", r, v)
		for _, n := range p.Nodes(false) {
			go func(node common.KeygenNodeDetails) {
				msg, err := NewEst2Message(m.roundID, v, r, m.curve, p.ID())
				if err != nil {
					return
				}
				p.Send(*msg, node)
			}(n)
		}
	}

	if est2Len > (2*f)+1 && !Contains(store.Bin("bin2", r), v) {
		store.SetBin("bin2", r, v)
		bin2 := store.Bin("bin2", r)
		w := bin2[0]
		for _, n := range p.Nodes(false) {
			go func(node common.KeygenNodeDetails) {
				msg, err := NewAux2Message(m.roundID, w, r, m.curve, p.ID())
				if err != nil {
					return
				}
				p.Send(*msg, node)
			}(n)
		}
	}
}
