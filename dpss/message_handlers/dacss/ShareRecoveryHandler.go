package dacss

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
)

var ShareRecoveryMessageType string = "dacss_share_recovery"

type ShareRecoveryMessage struct {
	RoundID     common.PSSRoundID
	AcssRoundID common.ACSSRoundID
	Kind        string

	// TODO add fields
}

func NewShareRecoveryMessage(roundID common.PSSRoundID, acssID common.ACSSRoundID) (*common.PSSMessage, error) {
	m := &ShareRecoveryMessage{
		RoundID:     roundID,
		AcssRoundID: acssID,
		Kind:        ShareRecoveryMessageType,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}
