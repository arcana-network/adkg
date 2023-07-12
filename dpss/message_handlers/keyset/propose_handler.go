package keyset

import (
	"encoding/binary"
	"encoding/json"
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"

	"github.com/arcana-network/groot/logger"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var ProposeMessageType common.DPSSMessageType = "keyset_propose"

type ProposeMessage struct {
	RoundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	data    []byte
}

func NewProposeMessage(id common.DPSSRoundID, d []byte, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := &ProposeMessage{
		id,
		ProposeMessageType,
		curve,
		d,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m *ProposeMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	log.Debugf("Received keyset Propose message from %d on %d", sender.Index, p.ID())
	log.Debug("Starting keyset", logger.Field{
		"node":   p.ID(),
		"sender": sender.Index,
		"data":   m.data,
	})
	leader, err := m.RoundID.Leader()
	if err != nil {
		return
	}
	senderBigInt := *new(big.Int).SetInt64(int64(sender.Index))

	// If leader of the round is not sender skip
	if leader.Cmp(&senderBigInt) != 0 {
		return
	}
	// Verify keyset predicate Tj and output
	log.Debugf("Going to verify keyset predicate for node=%d", p.ID())

	store, err := dpsscommon.GetSessionStoreFromRoundID(m.RoundID, p)
	store.Lock()
	defer store.Unlock()
	if err != nil {
		log.Debugf("Could not get session store from roundID, err=%s", err)
		return
	}
	log.Debugf("keyset_propose: output=%q, data=%q", store.TPrime, m.data)
	var Ti [8]byte
	binary.BigEndian.PutUint64(Ti[:], uint64(store.TPrime))
	verified := Predicate(m.data, Ti[:])

	// If verified, send echo to each node
	if verified {
		store.T[sender.Index] = int(binary.BigEndian.Uint64(m.data))
		// Create RS encoding
		n, k, _ := p.Params(false)
		f, err := infectious.NewFEC(k, n)
		if err != nil {
			log.Debugf("error during creation of fec, err=%s", err)
			return
		}

		hash := dpsscommon.Hash(m.data)

		shares, err := acss.Encode(f, m.data)
		if err != nil {
			log.Debugf("error during fec encoding, err=%s", err)
			return
		}

		for _, n := range p.Nodes(false) {
			log.Debugf("Sending echo: from=%d, to=%d", p.ID(), n.Index)
			go func(node common.KeygenNodeDetails) {
				msg, err := NewEchoMessage(m.RoundID, shares[node.Index-1], hash, m.curve)
				if err != nil {
					return
				}
				p.Send(*msg, node)
			}(n)
		}
	} else {
		log.Debugf("Predicate failed on %d for keyset propose message by %d", p.ID(), sender.Index)
	}
}

func OnKeysetVerified(roundID common.DPSSRoundID, curve curves.Curve, keyset []byte,
	sessionStore *dpsscommon.DPSSSession, leader int, self dpsscommon.DPSSParticipant) {
	if leader != self.ID() {
		data := dpsscommon.ByteToIntValue(keyset)
		sessionStore.T[int(leader)] = data
	}

	n, k, _ := self.Params(false)

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
		log.Debugf("Sending echo: from=%d, to=%d", self.ID(), n.Index)
		go func(node common.KeygenNodeDetails) {
			echoMsg, err := NewEchoMessage(roundID, shares[node.Index-1], hash, &curve)
			if err != nil {
				log.WithField("error", err).Error("NewKeysetEchoMessage")
				return
			}
			self.Send(*echoMsg, node)
		}(n)
	}
}
