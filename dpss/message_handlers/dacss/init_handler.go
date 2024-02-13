package dacss

import (
	"encoding/json"
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var InitMessageType string = "init_dpss"

type InitMessage struct {
	RIndex    int
	BatchSize int
	Kind      string
	Curve     *curves.Curve
}

func NewInitMessage(rIndex, batchSize int, curve curves.Curve) (*common.DKGMessage, error) {
	m := InitMessage{
		rIndex,
		batchSize,
		InitMessageType,
		&curve,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage("INIT", m.Kind, bytes)
	return &msg, nil
}

func (msg *InitMessage) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {
	if sender.Index != self.ID() {
		return
	}

	log.Debugf("InitMessageHandler: Received Init message from self(%d), starting DPSS\n\n", sender.Index)

	dpssID := dpss.GenerateDPSSID(*new(big.Int).SetInt64(int64(msg.RIndex)), *new(big.Int).SetInt64(int64(msg.BatchSize)))
	round := common.RoundDetails{
		ADKGID: dpssID,
		Dealer: self.ID(),
		Kind:   "dacss",
	}
	acssShareMsg, err := NewDacssShareMessage(round.ID(), msg.Curve)
	if err != nil {
		log.WithField("error", err).Error("NewDacssShareMessage")
		return
	}
	go self.ReceiveMessage(self.Details(), *acssShareMsg)
}
