package dacss

import (
	"math/big"

	"github.com/arcana-network/adkg-proto/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var InitMessageType common.MessageType = "init_dpss"

type InitMessage struct {
	sender    int
	rIndex    int
	batchSize int
	kind      common.MessageType
	curve     *curves.Curve
}

func NewInitMessage(sender, rIndex, batchSize int, curve curves.Curve) common.DKGMessage {
	m := InitMessage{
		sender,
		rIndex,
		batchSize,
		InitMessageType,
		&curve,
	}
	return &m
}

func (m *InitMessage) Sender() int {
	return m.sender
}

func (m *InitMessage) Kind() common.MessageType {
	return m.kind
}

func (m *InitMessage) Process(p common.DkgParticipant) {
	if m.Sender() != p.ID() {
		return
	}

	log.Debugf("InitMessageHandler: Received Init message from self(%d), starting DPSS\n\n", m.Sender())

	dpssID := common.GenerateDPSSID(*new(big.Int).SetInt64(int64(m.rIndex)), *new(big.Int).SetInt64(int64(m.batchSize)))
	round := common.RoundDetails{
		ADKGID: dpssID,
		Dealer: p.ID(),
		Kind:   "dacss",
	}
	acssShareMsg := NewAcssShareMessage(round.ID(), m.curve, p.ID())
	go p.ReceiveMessage(acssShareMsg)
}
