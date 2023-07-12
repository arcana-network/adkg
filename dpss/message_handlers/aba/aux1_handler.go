package aba

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/groot/logger"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var Aux1MessageType common.DPSSMessageType = "aba_aux1"

type Aux1Message struct {
	roundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	v       int
	r       int
	data    []byte
}

func (m *Aux1Message) Data() []byte {
	return m.data
}

func NewAux1Message(id common.DPSSRoundID, v, r int, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := Aux1Message{
		roundID: id,
		kind:    Aux1MessageType,
		curve:   curve,
		v:       v,
		r:       r,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.roundID, m.kind, bytes)
	return &msg, nil
}

func (m *Aux1Message) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	v, r := m.v, m.r

	n, _, f := p.Params(false)
	store, complete := p.State().ABAStore.GetOrSetIfNotComplete(m.roundID, common.DefaultABAStore())
	if complete {
		log.Debugf("Keygen already complete: %s", m.roundID)
		return
	}
	store.Lock()
	defer store.Unlock()
	// Check if already present
	if Contains(store.Values("aux", r, v), sender.Index) {
		// log.Debugf("Got redundant AUX message from %d", m.Sender())
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
	log.Debug("Aux1MessageHandler", logger.Field{
		"r":       r,
		"v":       v,
		"node":    p.ID(),
		"bin":     bin,
		"aux0Len": aux0Len,
		"aux1Len": aux1Len,
	})
	if Contains(bin, 1) && aux1Len > n-f {
		view = append(view, 1)
		shouldSendAuxset = true
	} else if Contains(bin, 0) && aux0Len > n-f {
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

		for _, n := range p.Nodes(false) {
			go func(nodes common.KeygenNodeDetails) {
				msg, err := NewAuxsetMessage(m.roundID, auxsetVal, m.r, m.curve, p.ID())
				if err != nil {
					log.Error(err)
					return
				}
				p.Send(*msg, nodes)
			}(n)
		}
	}

}
