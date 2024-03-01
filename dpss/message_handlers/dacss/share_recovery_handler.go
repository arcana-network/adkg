package dacss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/torusresearch/bijson"
)

var ShareRecoveryMessageType string = "dacss_share_recovery"

type ShareRecoveryMessage struct {
	ACSSRoundDetails common.ACSSRoundDetails
	Kind             string

	// TODO add fields
}

func NewShareRecoveryMessage(acssRoundDetails common.ACSSRoundDetails) (*common.PSSMessage, error) {
	m := &ShareRecoveryMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ShareRecoveryMessageType,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, m.Kind, bytes)
	return &msg, nil
}
