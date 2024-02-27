package dacss

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/messages"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

var AcssProposeMessageType common.MessageType = "Acss_propose"

type AcssProposeMessage struct {
	RoundID            common.PSSRoundID
	NewCommittee       bool
	Kind               common.MessageType
	Curve              common.CurveName
	Data               messages.MessageData
	EphemeralPublicKey curves.Point // the dealer's ephemeral publicKey

}

// convert json to acssProposeMessage
func (m *AcssProposeMessage) UnmarshalJSON(data []byte) error {
	type Alias AcssProposeMessage
	aux := &struct {
		EphemeralPublicKey json.RawMessage `json:"EphemeralPublicKey"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if len(aux.EphemeralPublicKey) > 0 {
		epk, err := common.PointUnmarshalJson([]byte(aux.EphemeralPublicKey))
		if err != nil {
			return err
		}
		m.EphemeralPublicKey = epk
	}

	return nil
}

// convert acssProposeMessage to json
func (m *AcssProposeMessage) MarshalJSON() ([]byte, error) {
	// Convert EphemeralPublicKey to a suitable JSON representation
	epkJSON, err := common.PointMarshalJson(m.EphemeralPublicKey)
	if err != nil {
		return nil, err
	}

	// Marshal the rest of AcssProposeMessage as usual, but replace
	// EphemeralPublicKey with its JSON representation
	type Alias AcssProposeMessage // Prevent recursion
	return json.Marshal(&struct {
		EphemeralPublicKey json.RawMessage `json:"EphemeralPublicKey"`
		*Alias
	}{
		EphemeralPublicKey: json.RawMessage(epkJSON),
		Alias:              (*Alias)(m),
	})
}

func NewAcssProposeMessageroundID(roundID common.PSSRoundID, msgData messages.MessageData, curveName common.CurveName, isNewCommittee bool, ephemeralPublicKey curves.Point) (*common.PSSMessage, error) {
	m := AcssProposeMessage{
		RoundID:            roundID,
		NewCommittee:       isNewCommittee,
		Kind:               AcssProposeMessageType,
		Curve:              curveName,
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

// func (msg *HbAacssProposeMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

// 	log.Debugf("Received Propose message from %d on %d", sender.Index, self.Details().Index)
// 	log.Debugf("Propose: Node=%d, Value=%v", self.Details().Index, msg.Data)

// 	leader, err := msg.RoundID.Leader()
// 	if err != nil {
// 		log.Debugf("Cound not get leader from roundID, err=%s", err)
// 		return
// 	}
// 	sender_big_int := *new(big.Int).SetInt64(int64(sender.Index))

// 	// If leader of the round is not sender skip
// 	if leader.Cmp(&sender_big_int) != 0 {
// 		return
// 	}

// 	// Generated shared symmetric key
// 	n, k, _ := self.Params()

// 	dealerKey := self.PublicKey(int(leader.Int64()), msg.NewCommittee)

// 	priv := self.PrivateKey()
// 	key := acss.SharedKey(priv, dealerKey)

// 	// Verify self share against commitments.
// 	log.Debugf("Going to verify predicate for node=%d", self.Details().Index)
// 	log.Debugf("IMP1: round=%s, node=%d, msg=%v", msg.RoundID, self.Details().Index, msg.Data)

// 	_, _, verified := acss.Predicate(key[:], msg.Data.ShareMap[uint32(self.Details().Index)][:],
// 		msg.Data.Commitments[:], k, msg.Curve)

// 	// If verified, send echo to each node
// 	if verified {

// 		// Store dealerPublicKey
// 		//TODO: GetSessionStoreFromRoundID is not defined
// 		// TODO: The variable s is not being used.

// 		// s, err := common.GetSessionStoreFromRoundID(msg.RoundID, self)
// 		// if err != nil {
// 		// 	log.Debugf("Could not get session store for roundID=%s, self=%d", msg.RoundID, self.ID())
// 		// 	return
// 		// }
// 		// s.Lock()
// 		// defer s.Unlock()

// 		// Starts the RBC protocol.
// 		// Create Reed-Solomon encoding. This is part of the RBC protocol.
// 		f, err := infectious.NewFEC(k, n)
// 		if err != nil {
// 			log.Debugf("error during creation of fec, err=%s", err)
// 			return
// 		}

// 		// Serialize data
// 		msg_bytes, err := msg.Data.Serialize()
// 		// msg, err := m.data.Serialize()
// 		if err != nil {
// 			log.Debugf("error during data serialization of MsgData, err=%s", err)
// 			return
// 		}

// 		// This corresponds to Line 8, Algorithm 4 of "Asynchronous data disemination and applications."
// 		msg_hash := common.HashByte(msg_bytes)

// 		// Obtain Reed-Solomon shards.
// 		shares, err := acss.Encode(f, msg_hash)
// 		if err != nil {
// 			log.Debugf("error during fec encoding, err=%s", err)
// 			return
// 		}

// 		for _, n := range self.Nodes(msg.NewCommittee) {
// 			log.Debugf("Sending echo: from=%d, to=%d", self.Details().Index, n.Index)
// 			go func(node common.NodeDetails) {

// 				//This instruction corresponds to Line 10, Algorithm 4 from
// 				//"Asynchronous data disemination and applications."
// 				echoMsg, err := NewDacssEchoMessage(msg.RoundID, shares[node.Index-1], msg_hash, msg.Curve, self.Details().Index, msg.NewCommittee)
// 				if err != nil {
// 					log.WithField("error", err).Error("NewDacssEchoMessage")
// 					return
// 				}
// 				self.Send(node, *echoMsg)
// 			}(n)
// 		}
// 	} else {
// 		log.Debugf("Predicate failed on %d for propose message by %d", self.Details().Index, sender.Index)
// 	}
// }
