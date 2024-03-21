package dacss

import (
	"encoding/hex"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var DacssOutputMessageType string = "dacss_output"

type DacssOutputMessage struct {
	AcssRoundDetails common.ACSSRoundDetails
	kind             string
	curveName        common.CurveName
	Data             []byte
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

//TODO: This output handler is incomplete
// But is suficient for testing the end to end flow of the DACSS

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

	// Retrieves the state.
	state, found, err := self.State().AcssStore.Get(
		m.AcssRoundDetails.ToACSSRoundID(),
	)
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error retrieving the state of the node.",
			},
		).Error("DACSSOutputMessage: Process")
		return
	}
	if !found {
		log.WithFields(
			log.Fields{
				"Message": "The state was not found",
			},
		).Error("DACSSOutputMessage: Process")
		return
	}

	// Check if the RBC state has already ended.
	if state.RBCState.Phase == common.Ended {
		log.WithFields(
			log.Fields{
				"RBCState": state.RBCState.Phase,
				"Message":  "The RBC state has already finished. Doing an early return",
			},
		).Info("DACSSOutputMessage: Process")
		return
	}

	self.State().AcssStore.Lock()
	defer self.State().AcssStore.Unlock()

	log.Debugf("acss_output: round=%v, self=%v", m.AcssRoundDetails, self.Details().Index)

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

	_, k, _ := self.Params()
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

	share, verifier, verified := sharing.Predicate(key, msgData.ShareMap[hexPubKey], msgData.Commitments, k, curve)

	if verified && !state.CommitmentSent {
		log.Debugf("acss_verified: share=%v", *share)

		// Set the state to reflect that RBC has ended.
		self.State().AcssStore.UpdateAccsState(
			m.AcssRoundDetails.ToACSSRoundID(),
			func(state *common.AccsState) {
				state.RBCState.Phase = common.Ended

				// Store the shares received at the end of the RBC.
				state.ReceivedShare = share

				// Line 203, Algorithm 4, DPS paper. Stores the commitment
				concatCommitments := sharing.CompressCommitments(verifier)
				hashCommitments := common.HashByte(concatCommitments)
				state.OwnCommitmentsHash = hex.EncodeToString(hashCommitments)
			},
		)

		commitmentMsg, err := NewDacssCommitmentMessage(
			m.AcssRoundDetails,
			m.curveName,
			verifier,
		)
		if err != nil {
			log.Errorf("Error creating the Commitment message: %v", err)
			return
		}

		self.State().AcssStore.UpdateAccsState(
			m.AcssRoundDetails.ToACSSRoundID(),
			func(state *common.AccsState) {
				state.CommitmentSent = true
			},
		)
		go self.Broadcast(!self.IsNewNode(), *commitmentMsg)
	} else if !verified {
		log.WithFields(
			log.Fields{
				"Message": "The predicate was not verified correctly",
			},
		).Error("DACSSOutputMessage: Process")
	}
}
