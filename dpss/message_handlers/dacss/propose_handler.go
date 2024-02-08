package dacss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/messages"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

type DacssProposeMessage struct{}

func NewDacssProposeMessage(roundID common.RoundID, msgData messages.MessageData, msgCurve curves.Curve, id int, isNewCommittee bool) common.PSSMessage {
}

func (msg *DacssProposeMessage) Kind() common.MessageType {

}

func (msg *DacssProposeMessage) Process(sender common.KeygenNodeDetails, self common.PSSParticipant) {

}
