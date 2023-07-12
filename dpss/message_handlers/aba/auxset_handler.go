package aba

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var AuxsetMessageType common.DPSSMessageType = "aba_auxset"

type AuxsetMessage struct {
	roundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	v       int
	r       int
}

func NewAuxsetMessage(id common.DPSSRoundID, v, r int, curve *curves.Curve, sender int) (*common.DPSSMessage, error) {
	m := AuxsetMessage{
		id,
		AuxsetMessageType,
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

func (m *AuxsetMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	store, complete := p.State().ABAStore.GetOrSetIfNotComplete(m.roundID, common.DefaultABAStore())
	if complete {
		log.Debugf("Keygen already complete: %s", m.roundID)
		return
	}
	store.Lock()
	defer store.Unlock()
	n, _, f := p.Params(false)

	if Contains(store.Values("auxset", m.r, m.v), sender.Index) {
		return
	}

	store.SetValues("auxset", m.r, m.v, sender.Index)
	if store.Round() != m.r {
		return
	}
	bin := store.Bin("bin", m.r)
	auxsetLen0 := len(store.Values("auxset", m.r, 0))
	auxsetLen1 := len(store.Values("auxset", m.r, 1))
	auxsetLen2 := len(store.Values("auxset", m.r, 2))

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

	if !store.Sent("est2", m.r, est2) && shouldSendEst2 {
		log.Debugf("PSSNode=%d: IN AUXSET_HANDLER: Sending EST2", p.ID())
		store.SetSent("est2", m.r, est2)
		for _, n := range p.Nodes(false) {
			go func(node common.KeygenNodeDetails) {
				msg, err := NewEst2Message(m.roundID, est2, m.r, m.curve, p.ID())
				if err != nil {
					return
				}
				p.Send(*msg, node)
			}(n)
		}
	}
}
