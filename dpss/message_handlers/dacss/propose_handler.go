package dacss

import (
	"math/big"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"

	"github.com/arcana-network/adkg-proto/common"
	"github.com/arcana-network/adkg-proto/common/acss"
	"github.com/arcana-network/adkg-proto/messages"
)

var AcssProposeMessageType common.MessageType = "dacss_propose"

// Stores the information needed in the proposal phase of the DACSS.
type AcssProposeMessage struct {
	id           common.RoundID
	sender       int
	newCommittee bool
	kind         common.MessageType
	curve        *curves.Curve
	data         messages.MessageData
}

func NewAcssProposeMessage(id common.RoundID, d messages.MessageData, curve *curves.Curve, sender int, newCommittee bool) common.DKGMessage {

	m := AcssProposeMessage{
		id:           id,
		sender:       sender,
		newCommittee: newCommittee,
		kind:         AcssProposeMessageType,
		curve:        curve,
		data:         d,
	}
	return &m
}

func (m *AcssProposeMessage) Sender() int {
	return m.sender
}

func (m *AcssProposeMessage) Kind() common.MessageType {
	return m.kind
}

func (m *AcssProposeMessage) Process(p common.DkgParticipant) {
	log.Debugf("Received Propose message from %d on %d", m.Sender(), p.ID())
	log.Debugf("Propose: Node=%d, Value=%v", p.ID(), m.data)

	leader, err := m.id.Leader()
	if err != nil {
		log.Debugf("Cound not get leader from roundID, err=%s", err)
		return
	}
	sender := *new(big.Int).SetInt64(int64(m.Sender()))

	// If leader of the round is not sender skip
	if leader.Cmp(&sender) != 0 {
		return
	}

	// Generated shared symmetric key
	n, k, _ := p.Params(m.newCommittee)

	dealerKey := p.PublicKey(int(leader.Int64()))

	priv := p.SelfPrivateKey()
	key := acss.SharedKey(priv, dealerKey)

	// Verify self share against commitments.
	log.Debugf("Going to verify predicate for node=%d", p.ID())
	log.Debugf("IMP1: round=%s, node=%d, msg=%v", m.id, p.ID(), m.data)

	_, _, verified := acss.Predicate(key[:], m.data.ShareMap[uint32(p.ID())][:],
		m.data.Commitments[:], k, m.curve)

	// If verified, send echo to each node
	if verified {

		// Store dealerPublicKey
		s, err := common.GetSessionStoreFromRoundID(m.id, p)
		if err != nil {
			log.Debugf("Could not get session store for roundID=%s, self=%d", m.id, p.ID())
			return
		}
		s.Lock()
		defer s.Unlock()

		// Starts the RBC protocol.
		// Create Reed-Solomon encoding. This is part of the RBC protocol.
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

		// This corresponds to Line 8, Algorithm 4 of "Asynchronous data disemination and applications."
		hash := common.Hash(msg)

		// Obtain Reed-Solomon shards.
		shares, err := acss.Encode(f, msg)
		if err != nil {
			log.Debugf("error during fec encoding, err=%s", err)
			return
		}

		for _, n := range p.Nodes(m.newCommittee) {
			log.Debugf("Sending echo: from=%d, to=%d", p.ID(), n.ID())
			go func(node common.DkgParticipant) {

				//This instruction corresponds to Line 10, Algorithm 4 from
				//"Asynchronous data disemination and applications."
				echoMsg := NewAcssEchoMessage(m.id, shares[node.ID()-1], hash, m.curve, p.ID(), m.newCommittee)
				p.Send(echoMsg, node)
			}(n)
		}
	} else {
		log.Debugf("Predicate failed on %d for propose message by %d", p.ID(), m.Sender())
	}
}
