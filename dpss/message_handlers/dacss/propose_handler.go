package dacss

import (
	"encoding/json"
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var AcssProposeMessageType common.DPSSMessageType = "dacss_propose"

type AcssProposeMessage struct {
	RoundID      common.DPSSRoundID
	newCommittee bool
	kind         common.DPSSMessageType
	curve        *curves.Curve
	data         messages.MessageData
}

func NewAcssProposeMessage(id common.DPSSRoundID, d messages.MessageData, curve *curves.Curve, newCommittee bool) (*common.DPSSMessage, error) {
	m := AcssProposeMessage{
		RoundID:      id,
		newCommittee: newCommittee,
		kind:         AcssProposeMessageType,
		curve:        curve,
		data:         d,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m *AcssProposeMessage) Kind() common.DPSSMessageType {
	return m.kind
}

func (m *AcssProposeMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	log.Debugf("Received Propose message from %d on %d", sender.Index, p.ID())
	log.Debugf("Propose: PSSNode=%d, Value=%v", p.ID(), m.data)

	leader, err := m.RoundID.Leader()
	if err != nil {
		log.Debugf("Cound not get leader from roundID, err=%s", err)
		return
	}
	senderID := *new(big.Int).SetInt64(int64(sender.Index))

	// If leader of the round is not sender skip
	if leader.Cmp(&senderID) != 0 {
		return
	}

	// Generated shared symmetric key
	n, k, _ := p.Params(m.newCommittee)

	dealerKey := p.PublicKey(int(leader.Int64()))

	priv := p.SelfPrivateKey()
	key := acss.SharedKey(priv, dealerKey)

	// Verify self share against commitments
	log.Debugf("Going to verify predicate for node=%d", p.ID())
	log.Debugf("IMP1: round=%s, node=%d, msg=%v", m.RoundID, p.ID(), m.data)

	_, _, verified := acss.Predicate(key[:], m.data.ShareMap[uint32(p.ID())][:],
		m.data.Commitments[:], k, m.curve)

	// If verified, send echo to each node
	if verified {

		// Store dealerPublicKey
		s, err := dpsscommon.GetSessionStoreFromRoundID(m.RoundID, p)
		if err != nil {
			log.Debugf("Could not get session store for roundID=%s, self=%d", m.RoundID, p.ID())
			return
		}
		s.Lock()
		defer s.Unlock()

		// Create RS encoding
		f, err := infectious.NewFEC(k, n)
		if err != nil {
			log.Debugf("error during creation of fec, err=%s", err)
			return
		}

		// Serialize data
		msg, err := m.data.Serialize()
		// msg, err := m.data.Serialize()
		if err != nil {
			log.Debugf("error during data serialization of MsgData, err=%s", err)
			return
		}

		hash := dpsscommon.Hash(msg)

		shares, err := acss.Encode(f, msg)
		if err != nil {
			log.Debugf("error during fec encoding, err=%s", err)
			return
		}

		for _, n := range p.Nodes(m.newCommittee) {
			log.Debugf("Sending echo: from=%d, to=%d", p.ID(), n.Index)
			go func(node common.KeygenNodeDetails) {
				msg, err := NewAcssEchoMessage(m.RoundID, shares[node.Index-1], hash, m.curve, m.newCommittee)
				if err != nil {
					return
				}
				p.Send(*msg, node)
			}(n)
		}
	} else {
		log.Debugf("Predicate failed on %d for propose message by %d", p.ID(), sender.Index)
	}
}
