package keyset

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"

	log "github.com/sirupsen/logrus"
)

var InitMessageType string = "keyset_init"

type InitMessage struct {
	RoundID common.RoundID
	Kind    string
	Data    []byte
	Curve   common.CurveName
}

func NewInitMessage(roundID common.RoundID, data []byte, curve common.CurveName) (*common.DKGMessage, error) {
	m := InitMessage{
		roundID,
		InitMessageType,
		data,
		curve,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m InitMessage) Process(sender common.NodeDetails, self common.DkgParticipant) {
	if sender.Index != self.ID() {
		log.WithFields(log.Fields{
			"Sender":  sender.Index,
			"Self":    self.ID,
			"Message": "Not equal, expected to be equal",
		}).Error("KeysetInitMessage:Process")
		return
	}

	proposeMsg, err := NewProposeMessage(m.RoundID, m.Data, m.Curve)
	if err != nil {
		log.WithField("error", err).Error("NewKeysetProposeMessage")
		return
	}

	go self.Broadcast(*proposeMsg)
}
