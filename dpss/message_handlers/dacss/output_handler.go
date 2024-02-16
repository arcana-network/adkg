package dacss

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

var AcssOutputMessageType common.MessageType = "dacss_output"

type AcssOutputMessage struct {
	roundID      common.RoundID
	sender       int
	kind         common.MessageType
	curve        *curves.Curve
	m            []byte
	newCommittee bool
	handlerType  string
}

func NewAcssOutputMessage(id common.RoundID, data []byte, curve *curves.Curve, sender int, handlerType string, newCommittee bool) (*common.DKGMessage, error) {
	m := AcssOutputMessage{
		roundID:      id,
		sender:       sender,
		newCommittee: newCommittee,
		kind:         AcssOutputMessageType,
		curve:        curve,
		m:            data,
		handlerType:  handlerType,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.roundID, string(m.kind), bytes)
	return &msg, nil
}
