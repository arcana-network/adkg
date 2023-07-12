package keyset

import (
	"encoding/binary"
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/aba"

	"github.com/arcana-network/groot/logger"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var OutputMessageType common.DPSSMessageType = "keyset_output"

type OutputMessage struct {
	RoundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	m       []byte
}

func NewOutputMessage(id common.DPSSRoundID, data []byte, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := &OutputMessage{
		RoundID: id,
		kind:    OutputMessageType,
		curve:   curve,
		m:       data,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m *OutputMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	log.Debug("message=keyset-output", logger.Field{
		"M":       m.m[:],
		"PSSNode": sender.Index,
		"round":   m.RoundID,
	})
	// log.SetOutput(nil)

	store, err := dpsscommon.GetSessionStoreFromRoundID(m.RoundID, p)
	if err != nil {
		return
	}
	store.Lock()
	defer store.Unlock()
	a := binary.BigEndian.Uint64(m.m)
	b := uint64(store.TPrime)
	log.Debug("predicate=keyset-output", logger.Field{
		"a":     a,
		"b":     b,
		"b&a":   a,
		"node":  p.ID(),
		"round": m.RoundID,
	})
	if b&a == a {
		vote := 1
		if store.ABAComplete {
			vote = 0
			log.Debugf("Voting 0 for %s", m.RoundID)
		}
		msg, err := aba.NewInitMessage(m.RoundID, vote, 0, m.curve)
		if err != nil {
			return
		}
		go p.ReceiveMessage(*msg)
	} else {
		msg, err := aba.NewInitMessage(m.RoundID, 0, 0, m.curve)
		if err != nil {
			return
		}
		go p.ReceiveMessage(*msg)
		log.Debug("Keysetoutput didnt pass predicate")
	}
}
