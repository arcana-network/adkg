package dacss

import (
	"encoding/hex"
	"reflect"

	"crypto/hmac"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
	"github.com/vivint/infectious"
)

var AcssProposeMessageType string = "dacss_propose"

type AcssProposeMessage struct {
	ACSSRoundDetails   common.ACSSRoundDetails
	NewCommittee       bool // Question: shouldn't this be redundant? Should be same as value in self?
	Kind               string
	CurveName          common.CurveName
	Data               common.AcssData // Encrypted shares, commitments & dealer's ephemeral public key for this ACSS round
	NewCommitteeParams common.CommitteeParams
}

func NewAcssProposeMessageround(acssRoundDetails common.ACSSRoundDetails, msgData common.AcssData, curveName common.CurveName, isNewCommittee bool, NewCommitteeParams common.CommitteeParams) (*common.PSSMessage, error) {
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

	self.State().AcssStore.Lock()
	defer self.State().AcssStore.Unlock()

	// Check whether the shares were already received. If so, ignore the message
	acssState, found, err := self.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error retrieving the state",
			},
		).Error("DACSSProposeMessage: Process")
	}
	if found && len(acssState.AcssDataHash) != 0 {
		log.Debugf("AcssProposeMessage: Shares already received for ACSS round %s", msg.ACSSRoundDetails.ToACSSRoundID())
		return
	}

	// Immediately: save hash of shares, commitments & dealer's ephemeral pubkey in node's state
	// in this way we can verify the shares when they arrive in the implicate flow
	acssDataHash, err := common.HashAcssData(msg.Data)
	if err != nil {
		log.Errorf("Couldn't hash acssData: %v", err)
		return
	}

	err = self.State().AcssStore.UpdateAccsState(msg.ACSSRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.AcssDataHash = acssDataHash
	})
	if err != nil {
		log.Errorf("Error updating AcssData in state: %v", err)
		return
	}

	// Check whether for this round we are already in Implicate flow, waiting for the shares that just arrived
	// If so, send ImplicateExecuteMessage for each stored ImplicateInformation
	// hbACSS Algorithm 1, line 401 (continued, upon initially receive IMPLICATE, couldn't proceed because of missing data).
	// Reference https://eprint.iacr.org/2021/159.pdf
	acssState, _, err = self.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())

	if err != nil {
		log.Errorf("Error getting the state state: %v", err)
		return
	}

	if len(acssState.ImplicateInformationSlice) > 0 {
		// It is possible to have received multiple implicate messages from different nodes
		// They should all be processed since some could be valid and some not
		for _, implicate := range acssState.ImplicateInformationSlice {
			// First verify that the received acssData equals the acssData that was received in the implicate flow

			if !reflect.DeepEqual(acssDataHash, implicate.AcssDataHash) {
				log.Errorf("Hash of acssData in implicate flow for ACSS round %s does not match the hash of the stored implicate information", msg.ACSSRoundDetails.ToACSSRoundID())
				return
			}

			implicateExecuteMessage, err := NewImplicateExecuteMessage(
				msg.ACSSRoundDetails,
				msg.CurveName,
				implicate.SymmetricKey,
				implicate.Proof,
				implicate.SenderPubkeyHex,
				msg.Data)
			if err != nil {
				log.Errorf("Error creating implicate execute msg in proposeHandler for implicate flow for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
				return
			}
			log.Debugf("Sending NewImplicateExecuteMessage: from=%d", self.Details().Index)
			go self.ReceiveMessage(self.Details(), *implicateExecuteMessage)
		}

	}

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
	log.Debugf("Going to verify predicate for node=%v, IsNewNode: %v", self.Details().Index, self.IsNewNode())
	log.Debugf("IMP1: round=%s, node=%s, msg=%v", msg.ACSSRoundDetails.ToACSSRoundID(), self.Details().GetNodeDetailsID(), msg.Data)

	pubKeyPoint, err := common.PointToCurvePoint(self.Details().PubKey, msg.CurveName)
	if err != nil {
		log.Errorf("AcssProposeMessage: error calculating pubKeyPoint: %v", err)
		return
	}

	hexPubKey := hex.EncodeToString(pubKeyPoint.ToAffineCompressed())

	encryptedShare, ExpectedHmac := sharing.Extract(msg.Data.ShareMap[hexPubKey][:])

	calculatedHMAC, err := sharing.GetHmacTag(encryptedShare, key.ToAffineCompressed())

	if err != nil {
		log.Errorf("AcssProposeMessage: error calculating HMAC: %v", err)
		return
	}

	result := hmac.Equal(calculatedHMAC, ExpectedHmac)

	// if !result {
	// 	log.Errorf("AcssProposeMessage: calculated hmac is different from the expected: %v", err)
	// 	return
	// }
	_, _, verified := sharing.Predicate(key, encryptedShare,
		msg.Data.Commitments[:], k, common.CurveFromName(msg.CurveName))

	//If verified, means the share is encrypted correctly and is valid wrt commitments

	// If verified:
	// - save in node's state that shares were validated
	// - send echo to each node
	if verified && result {

		// Starts the RBC protocol.
		// Create Reed-Solomon encoding. This is part of the RBC protocol.
		f, err := infectious.NewFEC(k, n)
		if err != nil {
			log.Debugf("error during creation of fec, err=%s", err)
			return
		}

		// Serialize data
		msg_bytes, err := bijson.Marshal(msg.Data)
		log.WithFields(
			log.Fields{
				"ACSS Data Bytes": msg_bytes,
				"Message":         "Bytes of the data created",
			},
		).Debug("DACSSProposeMessage: Process")

		if err != nil {
			log.Debugf("error during data serialization of MsgData, err=%s", err)
			return
		}

		// This corresponds to Line 8, Algorithm 4 of "Asynchronous data disemination and applications."
		msg_hash := common.HashByte(msg_bytes)

		// Obtain Reed-Solomon shards.
		shares, err := acss.Encode(f, msg_bytes)
		if err != nil {
			log.Debugf("error during fec encoding, err=%s", err)
			return
		}

		//store own share and hash
		self.State().AcssStore.UpdateAccsState(
			msg.ACSSRoundDetails.ToACSSRoundID(),
			func(state *common.AccsState) {

				// node index starts from 1, so have to correct here by subtracting 1
				state.RBCState.OwnReedSolomonShard = shares[self.Details().Index-1]
			},
		)

		for _, n := range self.Nodes(msg.NewCommittee) {
			log.Debugf("Sending echo: from=%d, to=%d", self.Details().Index, n.Index)

			//This instruction corresponds to Line 10, Algorithm 4 from
			//"Asynchronous data disemination and applications." Reference https://eprint.iacr.org/2021/777.pdf
			echoMsg, err := NewDacssEchoMessage(msg.ACSSRoundDetails, shares[n.Index-1], msg_hash, msg.CurveName, msg.NewCommittee)
			if err != nil {
				log.WithField("error", err).Error("NewDacssEchoMessage")
				return
			}
			go self.Send(n, *echoMsg)
		}
	} else {

		//If verified is false, that means either an error occured while decrypting share or shares not verified wrt commitments.
		//In that case send implicate with the ephemeral public key of the dealer to all other nodes in the same committee
		// hbACSS Algorihtm 1, line 206. Reference https://eprint.iacr.org/2021/159.pdf

		log.Debugf("Predicate failed on %d for propose message by %d", self.Details().Index, sender.Index)

		symmetricKey := key
		POKsymmetricKey := sharing.GenerateNIZKProof(curve, priv, pubKeyPoint, dealerKey, symmetricKey, curve.NewGeneratorPoint())

		implicateMsg, err := NewImplicateReceiveMessage(msg.ACSSRoundDetails, msg.CurveName, symmetricKey.ToAffineCompressed(), POKsymmetricKey, msg.Data)

		if err != nil {
			log.WithField("error constructing ImplicateMsg", err).Error("ImplicateReceiveMessage")
			return
		}

		for _, node := range self.Nodes(msg.NewCommittee) {
			go self.Send(node, *implicateMsg)
		}
	}
}
