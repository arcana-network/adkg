package keyset

import (
	"encoding/binary"
	"math"

	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/aba"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
)

var OutputMessageType string = "keyset_output"

type OutputMessage struct {
	RoundID common.PSSRoundDetails
	Kind    string
	Curve   common.CurveName
	M       []byte
}

func NewOutputMessage(id common.PSSRoundDetails, data []byte, curve common.CurveName) (*common.PSSMessage, error) {
	m := OutputMessage{
		id,
		OutputMessageType,
		curve,
		data,
	}
	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m OutputMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	// Ignore if not received by self
	if sender.Index != self.Details().Index {
		return
	}

	pssID := m.RoundID.PssID
	leader := m.RoundID.Dealer.Index

	// create default session to use below
	pssState, complete := self.State().PSSStore.GetOrSetIfNotComplete(pssID)
	if complete {
		// if keygen is complete, ignore and return
		log.Infof("keygen already complete: %v", m.RoundID)
		return
	}

	pssState.Lock()
	defer pssState.Unlock()

	log.WithFields(log.Fields{
		"pssID":          pssID,
		"self":           self.Details().Index,
		"alreadyStarted": pssState.ABAStarted,
	}).Info("aba_predicate")

	if kcommon.Contains(pssState.ABAStarted, leader) {
		return
	}
	n, _, t := self.Params()

	numShares := len(self.State().ShareStore.OldShares)

	alpha := int(math.Ceil(float64(numShares) / float64((n - 2*t))))
	TSet, _ := pssState.CheckForThresholdCompletion(alpha, n-t)
	b := uint64(TSet)
	a := binary.BigEndian.Uint64(m.M)

	vote := 0
	if b&a == a {
		if !pssState.ABAComplete {
			vote = 1
		}
	}

	pssState.ABAStarted = append(pssState.ABAStarted, leader)
	msg, err := aba.NewInitMessage(m.RoundID, vote, 0, m.Curve)
	if err != nil {
		log.WithError(err).Error("Could not create init message")
		return
	}
	go self.ReceiveMessage(self.Details(), *msg)
}
