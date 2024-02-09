package router

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

type RbcRouter struct {
	roundID       common.RoundID
	protoOrigin   string
	rbcFinished   bool
	outputMessage []byte
}

func NewRbcRouter(protoOrigin string) RbcRouter {
	return RbcRouter{protoOrigin: protoOrigin}
}

func (r RbcRouter) StartRbc(roundID common.RoundID, msgData messages.MessageData, msgCurve *curves.Curve, id int, newCommittee bool) {

}

func (r RbcRouter) ReturnRbcResult() {
	if !r.rbcFinished {
		return
	}

	if r.protoOrigin == "dacss" {
		outputMsg := dacss.NewDacssOutputMessage(r.roundID, r.outputMessage, m.Curve, self.ID(), "ready", m.NewCommittee)
		go self.ReceiveMessage(self.Details(), outputMsg)
	} else if r.protoOrigin == "acss" {
		outputMsg, err := acss.NewOutputMessage(m.RoundID, r.outputMessage, common.CurveName(m.Curve.Name))
		if err != nil {
			return
		}
		go self.ReceiveMessage(self.Details(), *outputMsg)
	}
}
