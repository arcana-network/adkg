package dacss

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/vivint/infectious"
)

var AcssReadyMessageType common.MessageType = "dacss_ready"

const (
	OldCommittee = iota
	NewCommittee
)

// Stores the information for the READY message in the RBC protocol.
type DacssReadyMessage struct {
	RoundID       common.RoundID
	NewCommittee  bool
	CommitteeType int
	Kind          common.MessageType
	Curve         *curves.Curve
	Share         infectious.Share
	Hash          []byte
}

func NewDacssReadyMessage(id common.RoundID, s infectious.Share, hash []byte, curve *curves.Curve, sender int, newCommittee bool) (*common.DKGMessage, error) {
	m := DacssReadyMessage{
		RoundID:      id,
		NewCommittee: newCommittee,
		Kind:         AcssReadyMessageType,
		Curve:        curve,
		Share:        s,
		Hash:         hash,
	}
	if newCommittee {
		m.CommitteeType = 1
	} else {
		m.CommitteeType = 0
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}
