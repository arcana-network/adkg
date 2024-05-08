package keyset

import (
	"math"

	"github.com/arcana-network/dkgnode/common"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/torusresearch/bijson"

	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var ProposeMessageType string = "keyset_propose"

type ProposeMessage struct {
	RoundID common.PSSRoundDetails
	Kind    string
	Curve   common.CurveName
	Data    []byte
}

func NewProposeMessage(id common.PSSRoundDetails, d []byte, curve common.CurveName) (*common.PSSMessage, error) {
	m := ProposeMessage{
		id,
		ProposeMessageType,
		curve,
		d,
	}
	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m ProposeMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.Debugf("Received keyset Propose message from %d on %d", sender.Index, self.Details().Index)

	leader := m.RoundID.Dealer.Index

	// If leader of the round is not sender skip
	if leader != sender.Index {
		return
	}

	// Verify keyset predicate Tj and output
	log.Debugf("Verify keyset predicate for node=%d, leader=%d", self.Details().Index, leader)

	pssID := m.RoundID.PssID

	pssState, complete := self.State().PSSStore.GetOrSetIfNotComplete(pssID)
	if complete {
		log.Infof("pss already complete: %s", pssID)
		return
	}

	n, _, t := self.Params()

	pssState.Lock()
	defer pssState.Unlock()

	numShares := m.RoundID.BatchSize

	alpha := int(math.Ceil(float64(numShares) / float64((n - 2*t))))
	TSet, _ := pssState.CheckForThresholdCompletion(alpha, n-t)
	verified := Predicate(kcommon.IntToByteValue(TSet), m.Data)

	// If verified, send echo to each node
	if verified {
		OnKeysetVerified(m.RoundID, m.Curve, m.Data, pssState, leader, self)
	} else {
		pssState.TProposals[sender.Index] = kcommon.ByteToIntValue(m.Data)
	}
}

func OnKeysetVerified(roundID common.PSSRoundDetails, curve common.CurveName, keyset []byte,
	pssState *common.PSSState, leader int, self common.PSSParticipant) {
	if leader != self.Details().Index {
		data := kcommon.ByteToIntValue(keyset)
		pssState.T[int(leader)] = data
	}

	n, k, _ := self.Params()

	// Create RS encoding
	fec, err := infectious.NewFEC(k, n)
	if err != nil {
		log.Debugf("error during creation of fec, err=%s", err)
		return
	}

	hash := common.HashByte(keyset)

	shares, err := acss.Encode(fec, keyset)
	if err != nil {
		log.Debugf("error during fec encoding, err=%s", err)
		return
	}
	for _, n := range self.Nodes(false) {
		log.Debugf("Sending echo: from=%d, to=%d", self.Details().Index, n.Index)
		go func(node common.NodeDetails) {
			echoMsg, err := NewEchoMessage(roundID, shares[node.Index-1], hash, curve)
			if err != nil {
				log.WithField("error", err).Error("NewKeysetEchoMessage")
				return
			}
			err = self.Send(node, *echoMsg)
			if err != nil {
				log.WithField("error", err).Error("KeysetEchoMessage: send")
				return
			}
		}(n)
	}
}
