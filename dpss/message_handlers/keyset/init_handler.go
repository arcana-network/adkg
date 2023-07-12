package keyset

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

var KeysetInitMessageType common.DPSSMessageType = "keyset_init"

type KeysetInitMessage struct {
	RoundID common.DPSSRoundID
	kind    common.DPSSMessageType
	data    []byte
	curve   *curves.Curve
}

func NewKeysetInitMessage(roundID common.DPSSRoundID, d []byte, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := KeysetInitMessage{
		roundID,
		KeysetInitMessageType,
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

func (m *KeysetInitMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	if sender.Index != p.ID() {
		return
	}

	msg, err := NewKeysetProposeMessage(m.RoundID, m.data, m.curve)
	if err != nil {
		return
	}

	p.Broadcast(false, *msg)
}
