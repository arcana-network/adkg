package dacss

import (
	"crypto/hmac"
	"encoding/hex"
	"math"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/keyset"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var DacssOutputMessageType string = "dacss_output"

type DacssOutputMessage struct {
	AcssRoundDetails common.ACSSRoundDetails
	kind             string
	curveName        common.CurveName
	Data             []byte // Contains the reconstructed initial msg (that contains acssData)
}

func NewDacssOutputMessage(roundDetails common.ACSSRoundDetails, data []byte, curveName common.CurveName) (*common.PSSMessage, error) {
	m := DacssOutputMessage{
		AcssRoundDetails: roundDetails,
		kind:             DacssOutputMessageType,
		curveName:        curveName,
		Data:             data,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.AcssRoundDetails.PSSRoundDetails, string(m.kind), bytes)
	return &msg, nil
}

func (m DacssOutputMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.WithFields(
		log.Fields{
			"MsgDataInfo": m.Data,
			"Message":     "Output Message Received",
			"Receiver":    self.Details().Index,
			"Sender":      sender.Index,
		},
	).Debug("DACSSOutputMessage: Process")

	// Ignore if not received by self
	if !sender.IsEqual(self.Details()) {
		log.WithFields(
			log.Fields{
				"Sender.Index": sender.Index,
				"Self.Index":   self.Details().Index,
				"Message":      "Not equal. Expected to be equal.",
			},
		).Error("DacssOutputMessage: Process")
		return
	}

	self.State().AcssStore.Lock()

	// Using defer because the ACSS state is being used until the end.
	defer self.State().AcssStore.Unlock()

	// Retrieves the state.
	state, found, err := self.State().AcssStore.Get(
		m.AcssRoundDetails.ToACSSRoundID(),
	)
	if err != nil {
		common.LogStateRetrieveError("DacssOutputMessage", "Process", err)
		return
	}
	if !found {
		common.LogStateNotFoundError("DacssOutputMessage", "Process", found)
		return
	}

	// Check if the RBC state has already ended.
	if state.RBCState.Phase == common.Ended {
		log.WithFields(
			log.Fields{
				"Message":   "The RBC state has already finished. Doing an early return",
				"NodeIndex": self.Details().Index,
			},
		).Debug("DACSSOutputMessage: Process")
		return
	}

	log.Debugf("dacss_output: round=%v, self=%v", m.AcssRoundDetails, self.Details().Index)

	priv := self.PrivateKey()

	msgData := common.AcssData{}

	// retrive the ACSSData
	err = bijson.Unmarshal(m.Data, &msgData)

	if err != nil {
		log.WithFields(
			log.Fields{
				"Message": "Could not deserialize message data",
				"Error":   err,
				"Data":    m.Data,
			},
		).Error("DACSSOutputMessage: Process")
		return
	}

	n, k, t := self.Params()
	curve := common.CurveFromName(m.curveName)

	pubKeyPoint, err := common.PointToCurvePoint(self.Details().PubKey, m.curveName)

	if err != nil {
		log.Errorf("Error converting from point to point: %v", err)
		return
	}

	hexPubKey := hex.EncodeToString(pubKeyPoint.ToAffineCompressed())

	EphemeralPubkeyBytes, err := hex.DecodeString(msgData.DealerEphemeralPubKey)
	if err != nil {
		log.Errorf("Error decoding hex string: %v", err)
		return
	}

	dealerKey, err := curve.Point.FromAffineCompressed(EphemeralPubkeyBytes)

	if err != nil {
		log.Errorf("Error FromAffineCompressed: %v", err)
		return
	}
	key, err := sharing.CalculateSharedKey(dealerKey, priv)

	if err != nil {
		log.Errorf("Error CalculateSharedKey: %v", err)
		return
	}

	encryptedShare, ExpectedHmac := sharing.Extract(msgData.ShareMap[hexPubKey][:])

	calculatedHMAC, err := sharing.GetHmacTag(encryptedShare, key.ToAffineCompressed())

	if err != nil {
		log.Errorf("AcssProposeMessage: error calculating HMAC: %v", err)
		return
	}

	result := hmac.Equal(calculatedHMAC, ExpectedHmac)

	if !result {
		log.Errorf("AcssProposeMessage: calculated hmac is different from the expected: %v", err)
		return
	}

	share, verifier, verified := sharing.Predicate(key, encryptedShare, msgData.Commitments, k, curve)

	if !verified {
		// Start the implicate flow
		log.WithFields(
			log.Fields{
				"Message": "The predicate was not verified correctly",
			},
		).Error("DACSSOutputMessage: Process")

		symmetricKey := key
		POKsymmetricKey := sharing.GenerateNIZKProof(curve, priv, pubKeyPoint, dealerKey, symmetricKey, curve.NewGeneratorPoint())

		implicateMsg, err := NewImplicateReceiveMessage(m.AcssRoundDetails, m.curveName, symmetricKey.ToAffineCompressed(), POKsymmetricKey, msgData)

		if err != nil {
			common.LogErrorNewMessage("DACSSOutputMessage", "Process", ImplicateReceiveMessageType, err)
			return
		}

		for _, node := range self.Nodes(self.IsNewNode()) {
			go self.Send(node, *implicateMsg)
		}
		return
	}

	if verified && !state.CommitmentSent {
		log.Debugf("acss_verified: share=%v", *share)

		// Set the state to reflect that RBC has ended.
		_, err = self.State().AcssStore.UpdateAccsState(
			m.AcssRoundDetails.ToACSSRoundID(),
			func(state *common.AccsState) {
				state.RBCState.Phase = common.Ended

				// Store the share received at the end of the RBC.
				state.ReceivedShare = share

				// Line 203, Algorithm 4, DPS paper. Stores the commitment
				// We just send g^{secret} instead of the commitment of the
				// complete polynomial because the polynomials for the same
				// secret will be inherently different for the same secret and
				// between the old and new committee.
				state.OwnCommitmentsHash = hex.EncodeToString(
					common.HashByte(verifier.Commitments[0].ToAffineCompressed()),
				)
			},
		)
		if err != nil {
			common.LogStateUpdateError("OutputHandler", "Process", common.AcssStateType, err)
			return
		}

		commitmentMsg, err := NewDacssCommitmentMessage(
			m.AcssRoundDetails,
			m.curveName,
			verifier.Commitments[0],
		)
		if err != nil {
			common.LogErrorNewMessage("DacssOutputMessage", "Process", DacssCommitmentMessageType, err)
			return
		}

		_, err = self.State().AcssStore.UpdateAccsState(
			m.AcssRoundDetails.ToACSSRoundID(),
			func(state *common.AccsState) {
				state.CommitmentSent = true
			},
		)
		if err != nil {
			common.LogStateUpdateError("OutputHandler", "Process", common.AcssStateType, err)
			return
		}

		go self.Broadcast(!self.IsNewNode(), *commitmentMsg)

		// We need to check if the conditions for the commitment handler hold here
		// because this node could have received commitment messages before reaching this point.
		_, _, t := self.Params()
		commitmentHexHash, found := state.FindThresholdCommitment(t + 1)
		if found {
			// Computes the hash of the own commitment
			if commitmentHexHash == state.OwnCommitmentsHash {
				_, err = self.State().AcssStore.UpdateAccsState(
					m.AcssRoundDetails.ToACSSRoundID(),
					func(state *common.AccsState) {
						state.ValidShareOutput = true
					},
				)
				if err != nil {
					common.LogStateUpdateError("OutputHandler", "Process", common.AcssStateType, err)
					return
				}

				log.WithFields(
					log.Fields{
						"Message":   "commitment finished correctly. Start MBVA here",
						"SelfIdx":   self.Details().Index,
						"IsNewNode": self.IsNewNode,
					},
				).Debug("DacssOutputMessage: process")
			}
		} else {
			log.WithFields(
				log.Fields{
					"Threshold": t + 1,
					"Message":   "There is no commitment record surpasing the threshold",
				},
			).Info("DacssOutputMessage: Process")
		}

	} else if state.CommitmentSent {
		log.WithFields(
			log.Fields{
				"Message":   "The commitment was already sent",
				"NodeIndex": self.Details().Index,
			},
		).Error("DACSSOutputMessage: Process")
	}

	{
		// Storing this for easier fetch
		// Waiting time for 1 unlock < waiting time for (B / n-2t) * n unlocks
		dealer := m.AcssRoundDetails.PSSRoundDetails.Dealer

		pssState, complete := self.State().PSSStore.GetOrSetIfNotComplete(m.AcssRoundDetails.PSSRoundDetails.PssID)
		if complete {
			return
		}

		pssState.Lock()
		defer pssState.Unlock()

		keysetMap := pssState.GetKeysetMap(m.AcssRoundDetails.ACSSCount)
		keysetMap.TPrime = dpsscommon.SetBit(keysetMap.TPrime, dealer.Index)
		keysetMap.ShareStore[dealer.Index] = share
		keysetMap.CommitmentStore[dealer.Index] = verifier.Commitments

		// Check if all shares received
		pssState.CheckAllSharesReceivedFromT(common.CurveFromName(m.curveName))
		// Check proposals and emit
		numShares := m.AcssRoundDetails.PSSRoundDetails.BatchSize
		alpha := int(math.Ceil(float64(numShares) / float64((n - 2*t))))
		TSet, _ := pssState.CheckForThresholdCompletion(alpha, n-t)
		for key, v := range pssState.TProposals {
			if keyset.Predicate(dpsscommon.IntToByteValue(TSet), dpsscommon.IntToByteValue(v)) {
				dealer, err := self.OldNodeDetailsByID(key)
				if err != nil {
					log.Error("Could not get old node details???")
					continue
				}
				roundID := common.CreatePSSRound(pssState.PSSID, dealer, m.AcssRoundDetails.PSSRoundDetails.BatchSize)
				keyset.OnKeysetVerified(roundID, m.curveName, dpsscommon.IntToByteValue(v), pssState, key, self)
				delete(pssState.TProposals, key)
			}
		}
	}
}
