package dacss

import (
	"encoding/json"
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var InitMessageType string = "hbACSS_init"

type InitMessage struct {
	RIndex    int
	BatchSize int // number of secrets to be shared in the hbACSS
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

	//TODO: needs to confirm whether id will be "INIT"?
	msg := common.CreateMessage("INIT", m.Kind, bytes)
	return &msg, nil
}

func (msg InitMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	if !sender.IsEqual(self.Details()) {
		return
	}

	log.Debugf("InitMessageHandler: Received Init message from self(%d), starting DPSS\n\n", sender.Index)

	dpssID := dpss.GenerateDPSSID(*new(big.Int).SetInt64(int64(msg.RIndex)), *new(big.Int).SetInt64(int64(msg.BatchSize)))
	round := common.PSSRoundDetails{
		PSSID:  dpssID,
		Dealer: self.Details().Index,
		Kind:   msg.Kind,
	}
	hbacssShareMsg, err := NewHbACSSacssShareMessage(msg.BatchSize, round.ID(), msg.Curve)
	if err != nil {
		log.WithField("error", err).Error("NewDacssShareMessage")
		return
	}
	go self.ReceiveMessage(self.Details(), *hbacssShareMsg)
}
