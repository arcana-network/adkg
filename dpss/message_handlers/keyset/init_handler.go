package keyset

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

var InitMessageType common.DPSSMessageType = "keyset_init"

type InitMessage struct {
	RoundID common.DPSSRoundID
	kind    common.DPSSMessageType
	data    []byte
	curve   *curves.Curve
}

func NewInitMessage(roundID common.DPSSRoundID, d []byte, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := InitMessage{
		roundID,
		InitMessageType,
		d,
		curve,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m *InitMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	if sender.Index != p.ID() {
		return
	}

	msg, err := NewProposeMessage(m.RoundID, m.data, m.curve)
	if err != nil {
		return
	}

	p.Broadcast(false, *msg)
}
