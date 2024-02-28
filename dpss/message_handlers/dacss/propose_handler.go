package dacss

import (
	"encoding/json"
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var AcssProposeMessageType common.MessageType = "Acss_propose"

type AcssProposeMessage struct {
	RoundID            common.PSSRoundID
	NewCommittee       bool
	Kind               common.MessageType
	CurveName          common.CurveName
	Data               messages.MessageData
	EphemeralPublicKey []byte // the dealer's ephemeral publicKey

}

func NewAcssProposeMessageroundID(roundID common.PSSRoundID, msgData messages.MessageData, curveName common.CurveName, isNewCommittee bool, ephemeralPublicKey []byte) (*common.PSSMessage, error) {
	m := AcssProposeMessage{
		RoundID:            roundID,
		NewCommittee:       isNewCommittee,
		Kind:               AcssProposeMessageType,
		CurveName:          curveName,
		Data:               msgData,
		EphemeralPublicKey: ephemeralPublicKey,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, string(m.Kind), bytes)
	return &msg, nil
}

func (msg *AcssProposeMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	log.Debugf("Received Propose message from %d on %d", sender.Index, self.Details().Index)
	log.Debugf("Propose: Node=%d, Value=%v", self.Details().Index, msg.Data)

	leader, err := msg.RoundID.Leader()
	if err != nil {
		log.Debugf("Cound not get leader from roundID, err=%s", err)
		return
	}
	senderId := *new(big.Int).SetInt64(int64(sender.Index))

	// If leader of the round is not sender skip
	if leader.Cmp(&senderId) != 0 {
		return
	}

	// Generated shared symmetric key
	n, k, _ := self.Params()

	curve := common.CurveFromName(msg.CurveName)
	dealerKey, err := curve.Point.FromAffineCompressed(msg.EphemeralPublicKey)
	if err != nil {
		log.Errorf("AcssProposeMessage: error constructing the EphemeralPublicKey: %v", err)
		return
	}

	priv := self.PrivateKey()
	key, err := sharing.CalculateSharedKey(dealerKey, priv)
	if err != nil {
		log.Errorf("AcssProposeMessage: error calculating shared key: %v", err)
		return
	}

	// Verify self share against commitments.
	log.Debugf("Going to verify predicate for node=%d", self.Details().Index)
	log.Debugf("IMP1: round=%s, node=%d, msg=%v", msg.RoundID, self.Details().Index, msg.Data)

	_, _, verified := sharing.Predicate(key, msg.Data.ShareMap[uint32(self.Details().Index)][:],
		msg.Data.Commitments[:], k, common.CurveFromName(msg.CurveName))

	//If verified, means the share is encrypted correctly and the commitments is also verified
	// If verified, send echo to each node
	if verified {

		// Store dealerPublicKey
		//TODO: GetSessionStoreFromRoundID is not defined
		// TODO: The variable s is not being used.

		// s, err := common.GetSessionStoreFromRoundID(msg.RoundID, self)
		// if err != nil {
		// 	log.Debugf("Could not get session store for roundID=%s, self=%d", msg.RoundID, self.ID())
		// 	return
		// }
		// s.Lock()
		// defer s.Unlock()

		// Starts the RBC protocol.
		// Create Reed-Solomon encoding. This is part of the RBC protocol.
		f, err := infectious.NewFEC(k, n)
		if err != nil {
			log.Debugf("error during creation of fec, err=%s", err)
			return
		}

		// Serialize data
		msg_bytes, err := msg.Data.Serialize()

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
			log.Debugf("Sending echo: from=%d, to=%d", self.Details().Index, n.Index)
			go func(node common.NodeDetails) {

				//This instruction corresponds to Line 10, Algorithm 4 from
				//"Asynchronous data disemination and applications."
				echoMsg, err := NewDacssEchoMessage(msg.RoundID, shares[node.Index-1], msg_hash, common.CurveFromName(msg.CurveName), self.Details().Index, msg.NewCommittee)
				if err != nil {
					log.WithField("error", err).Error("NewDacssEchoMessage")
					return
				}
				self.Send(node, *echoMsg)
			}(n)
		}
	} else {

		//If verified is false, that means either an error occured while decrypting share or shares not verified.
		//In that case send implicate with the ephemeral public key of the dealer
		//TODO: IMPLICATE

		log.Debugf("Predicate failed on %d for propose message by %d", self.Details().Index, sender.Index)
	}
}
