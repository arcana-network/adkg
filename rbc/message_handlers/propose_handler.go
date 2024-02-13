package rbc

import (
	"encoding/json"
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var RbcProposeMessageType common.MessageType = "dacss_propose"

// RbcProposeMessage represents the PROPOSE message in the RBC protocol
type RbcProposeMessage struct {
	RoundID     common.RoundID       // The current round ID.
	Kind        common.MessageType   // The kind of the message inside the RBC protocol.
	Curve       *curves.Curve        // The curve that we are using.
	Data        messages.MessageData // The data of the message.
	ProtoOrigin string               // The protocol that is calling RBC.
	// TODO: is this the place for this?
	NewCommittee bool // Defines if the message comes from an old or a new committee.
}

// NewRbcProposeMessage creates a new PROPOSE message for the RBC protocol.
func NewRbcProposeMessage(roundID common.RoundID, msgData messages.MessageData, msgCurve *curves.Curve, id int, newCommittee bool, protoOrigin string) (*common.DKGMessage, error) {
	m := RbcProposeMessage{
		RoundID:      roundID,
		Kind:         RbcProposeMessageType,
		Curve:        msgCurve,
		Data:         msgData,
		NewCommittee: newCommittee,
		ProtoOrigin:  protoOrigin,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (msg *RbcProposeMessage) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {
	log.Debugf("Received Propose message from %d on %d", sender.Index, self.ID())
	log.Debugf("Propose: Node=%d, Value=%v", self.ID(), msg.Data)

	leader, err := msg.RoundID.Leader()
	if err != nil {
		log.Debugf("Cound not get leader from roundID, err=%s", err)
		return
	}
	sender_big_int := *new(big.Int).SetInt64(int64(sender.Index))

	// If leader of the round is not sender skip
	if leader.Cmp(&sender_big_int) != 0 {
		return
	}

	// Generated shared symmetric key
	n, k, _ := self.Params(msg.NewCommittee)

	// Keys to verify the predicate.
	dealerKey := self.PublicKey(int(leader.Int64()), msg.NewCommittee)
	priv := self.PrivateKey()
	key := acss.SharedKey(priv, dealerKey)

	// Verify self share against commitments.
	log.Debugf("Going to verify predicate for node=%d", self.ID())
	log.Debugf("IMP1: round=%s, node=%d, msg=%v", msg.RoundID, self.ID(), msg.Data)

	_, _, verified := acss.Predicate(key[:], msg.Data.ShareMap[uint32(self.ID())][:],
		msg.Data.Commitments[:], k, msg.Curve)

	// If verified, send echo to each node
	if verified {

		// Store dealerPublicKey
		// TODO: The variable s is not being used.
		s, err := common.GetSessionStoreFromRoundID(msg.RoundID, self)
		if err != nil {
			log.Debugf("Could not get session store for roundID=%s, self=%d", msg.RoundID, self.ID())
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
		msg_bytes, err := msg.Data.Serialize()
		// msg, err := m.data.Serialize()
		if err != nil {
			log.Debugf("error during data serialization of MsgData, err=%s", err)
			return
		}

		// This corresponds to Line 8, Algorithm 4 of "Asynchronous data disemination and applications."
		msg_hash := common.HashByte(msg_bytes)

		// Obtain Reed-Solomon shards.
		shares, err := acss.Encode(f, msg_hash)
		if err != nil {
			log.Debugf("error during fec encoding, err=%s", err)
			return
		}

		for _, n := range self.Nodes(msg.NewCommittee) {
			log.Debugf("Sending echo: from=%d, to=%d", self.ID(), n.Index)
			go func(node common.KeygenNodeDetails) {

				//This instruction corresponds to Line 10, Algorithm 4 from
				//"Asynchronous data disemination and applications."
				echoMsg, err := NewRbcEchoMessage(msg.RoundID, shares[node.Index-1], msg_hash, msg.Curve, self.ID(), msg.NewCommittee, msg.ProtoOrigin)
				if err != nil {
					log.WithField("error", err).Error("NewDacssEchoMessage")
					return
				}
				self.Send(node, *echoMsg)
			}(n)
		}
	} else {
		log.Debugf("Predicate failed on %d for propose message by %d", self.ID(), sender.Index)
	}
}