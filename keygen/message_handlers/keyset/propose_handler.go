package keyset

import (
	"encoding/json"
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"

	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var ProposeMessageType string = "keyset_propose"

type ProposeMessage struct {
	RoundID common.RoundID
	Kind    string
	Curve   common.CurveName
	Data    []byte
}

func NewProposeMessage(id common.RoundID, d []byte, curve common.CurveName) (*common.DKGMessage, error) {
	m := ProposeMessage{
		id,
		ProposeMessageType,
		curve,
		d,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m ProposeMessage) Process(sender common.NodeDetails, self common.DkgParticipant) {
	log.Debugf("Received keyset Propose message from %d on %d", sender.Index, self.ID())

	leader, err := m.RoundID.Leader()
	if err != nil {
		return
	}
	// If leader of the round is not sender skip
	if leader.Cmp(new(big.Int).SetInt64(int64(sender.Index))) != 0 {
		return
	}

	// Verify keyset predicate Tj and output
	log.Debugf("Verify keyset predicate for node=%d, leader=%d", self.ID(), leader.Int64())

	adkgid, err := common.ADKGIDFromRoundID(m.RoundID)
	if err != nil {
		log.Infof("Could not get leader from roundID, err=%s", err)
		return
	}

	sessionStore, complete := self.State().SessionStore.GetOrSetIfNotComplete(adkgid, common.DefaultADKGSession())
	if complete {
		log.Infof("Keygen already complete: %s", adkgid)
		return
	}

	sessionStore.Lock()
	defer sessionStore.Unlock()

	verified := Predicate(kcommon.IntToByteValue(sessionStore.TPrime), m.Data)

	// If verified, send echo to each node
	if verified {
		OnKeysetVerified(m.RoundID, m.Curve, m.Data, sessionStore, int(leader.Int64()), self)
	} else {
		sessionStore.TProposals[sender.Index] = kcommon.ByteToIntValue(m.Data)
	}
}

func OnKeysetVerified(roundID common.RoundID, curve common.CurveName, keyset []byte,
	sessionStore *common.ADKGSession, leader int, self common.DkgParticipant) {
	if leader != self.ID() {
		data := kcommon.ByteToIntValue(keyset)
		sessionStore.T[int(leader)] = data
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
	for _, n := range self.Nodes() {
		log.Debugf("Sending echo: from=%d, to=%d", self.ID(), n.Index)
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
