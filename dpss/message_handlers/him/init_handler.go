package him

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
)

var InitMessageType string = "him_init"

// This is the sample implementation
type InitMessage struct {
	RoundID common.PSSRoundDetails
	Kind    string
	Curve   common.CurveName
}

func NewInitMessage(id common.PSSRoundDetails, curve common.CurveName) (*common.PSSMessage, error) {
	m := InitMessage{
		id,
		InitMessageType,
		curve,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}
