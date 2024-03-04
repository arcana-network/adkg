package dacss

import (
	"encoding/hex"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/torusresearch/bijson"

	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var AcssProposeMessageType common.MessageType = "Acss_propose"

type AcssProposeMessage struct {
	ACSSRoundDetails   common.ACSSRoundDetails
	NewCommittee       bool
	Kind               common.MessageType
	CurveName          common.CurveName
	Data               common.DacssData // Encrypted shares, commitments & dealer's ephemeral public key for this ACSS round
	NewCommitteeParams common.CommitteeParams
}

func NewAcssProposeMessageroundID(acssRoundDetails common.ACSSRoundDetails, msgData common.DacssData, curveName common.CurveName, isNewCommittee bool, NewCommitteeParams common.CommitteeParams) (*common.PSSMessage, error) {
	m := AcssProposeMessage{
		ACSSRoundDetails:   acssRoundDetails,
		NewCommittee:       isNewCommittee,
		Kind:               AcssProposeMessageType,
		CurveName:          curveName,
		Data:               msgData,
		NewCommitteeParams: NewCommitteeParams,
	}
	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, string(m.Kind), bytes)
	return &msg, nil
}

func (msg *AcssProposeMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	// Retrieve Dealer from PSSRoundID & verify it equals the sender
	dealerNodeDetails := msg.ACSSRoundDetails.PSSRoundDetails.Dealer
	if !dealerNodeDetails.IsEqual(sender) {
		return
	}
	//save shares in node's state
	state := common.DacssShareStoreMap{
		SharesForACSSRound: make(map[common.ACSSRoundID]common.DacssData),
	}
	state.SharesForACSSRound[msg.ACSSRoundDetails.ToACSSRoundID()] = msg.Data
	//storing into dacss state
	self.State().DacssStore = &state

	//Identified by nodeDetailsId
	log.Debugf("Received Propose message from %s on %s", sender.GetNodeDetailsID(), self.Details().GetNodeDetailsID())
	log.Debugf("Propose: Node=%s, Value=%s", self.Details().GetNodeDetailsID(), msg.Data)

	// Generated shared symmetric key
	n, k, _ := self.Params()
	if msg.NewCommittee {
		n = msg.NewCommitteeParams.N
		k = msg.NewCommitteeParams.K
	}
	curve := common.CurveFromName(msg.CurveName)

	pubkeyBytes, err := hex.DecodeString(msg.Data.DealerEphemeralPubKey)
	if err != nil {
		log.Errorf("Error decoding hex string: %v", err)
		return
	}

	dealerKey, err := curve.Point.FromAffineCompressed(pubkeyBytes)

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
	//TODO: we can identify by node index and whether in old or new committee by self.IsOldNode()
	log.Debugf("Going to verify predicate for node=%v, %v", self.Details().Index, self.IsOldNode())
	log.Debugf("IMP1: round=%s, node=%s, msg=%v", msg.ACSSRoundDetails.ToACSSRoundID(), self.Details().GetNodeDetailsID(), msg.Data)

	X := self.Details().PubKey.X
	Y := self.Details().PubKey.Y
	pubKeyPoint, err := curves.K256().NewIdentityPoint().Set(&X, &Y)

	if err != nil {
		log.Errorf("AcssProposeMessage: error calculating pubKeyPoint: %v", err)
		return
	}

	hexPubKey := hex.EncodeToString(pubKeyPoint.ToAffineCompressed())
	_, _, verified := sharing.Predicate(key, msg.Data.ShareMap[hexPubKey][:],
		msg.Data.Commitments[:], k, common.CurveFromName(msg.CurveName))

	//If verified, means the share is encrypted correctly and the commitments is also verified
	// If verified, send echo to each node
	if verified {

		// Store dealerPublicKey
		//TODO: GetSessionStoreFromRoundID is not defined
		// TODO: The variable s is not being used.

		// s, err := common.GetSessionStoreFromRoundID(msg.RoundID, self)
		// if err != nil {
		//  log.Debugf("Could not get session store for roundID=%s, self=%d", msg.RoundID, self.ID())
		//  return
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
		msg_bytes, err := bijson.Marshal(msg.Data)

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

			//TODO: running this go-routine result into error in few cases
			// Therefore, as of now we are directly sending the the msg

			// go func(node common.NodeDetails) {

			//This instruction corresponds to Line 10, Algorithm 4 from
			//"Asynchronous data disemination and applications."
			echoMsg, err := NewDacssEchoMessage(msg.ACSSRoundDetails, shares[n.Index-1], msg_hash, common.CurveFromName(msg.CurveName), self.Details().Index, msg.NewCommittee)
			if err != nil {
				log.WithField("error", err).Error("NewDacssEchoMessage")
				return
			}
			self.Send(n, *echoMsg)
			// }(n)
		}
	} else {

		//If verified is false, that means either an error occured while decrypting share or shares not verified.
		//In that case send implicate with the ephemeral public key of the dealer
		//TODO: IMPLICATE

		log.Debugf("Predicate failed on %d for propose message by %d", self.Details().Index, sender.Index)
	}
}
