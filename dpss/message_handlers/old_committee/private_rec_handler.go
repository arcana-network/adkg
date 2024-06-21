package old_committee

import (
	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var PrivateRecMessageType string = "dpss_private_rec"

type PrivateRecMsg struct {
	DPSSBatchRecDetails common.DPSSBatchRecDetails
	Kind                string
	curveName           common.CurveName
	UShare              []byte
}

func NewPrivateRecMsg(
	dpssBatchRecDetails common.DPSSBatchRecDetails,
	curve common.CurveName,
	uShare []byte,
) (*common.PSSMessage, error) {
	msg := PrivateRecMsg{
		DPSSBatchRecDetails: dpssBatchRecDetails,
		Kind:                PrivateRecMessageType,
		curveName:           curve,
		UShare:              uShare,
	}

	msgBytes, err := bijson.Marshal(msg)
	if err != nil {
		return nil, err
	}

	pssMessage := common.CreatePSSMessage(
		msg.DPSSBatchRecDetails.PSSRoundDetails,
		string(msg.Kind),
		msgBytes,
	)

	return &pssMessage, nil
}

func (msg *PrivateRecMsg) Process(sender common.NodeDetails, self common.PSSParticipant) {
	self.State().BatchReconStore.Lock()

	// Holding the lock until the end of the function because the Batch Rec.
	// state is being used until the end of the function.
	defer self.State().BatchReconStore.Unlock()

	// Initialize state here if it is not initialized in the InitHanlder
	_, err := self.State().BatchReconStore.UpdateBatchRecState(
		msg.DPSSBatchRecDetails.ToBatchRecID(),
		func(s *common.BatchRecState) {},
	)
	if err != nil {
		common.LogStateUpdateError("PrivateRecHandler", "Process", common.BatchRecStateType, err)
		return
	}

	// Deserialize the share.
	curve := common.CurveFromName(msg.curveName)
	share, err := curve.Scalar.SetBytes(msg.UShare)
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while converting the share from bytes to scalar",
			},
		)
		return
	}

	// Check if there are at least d + t + 1 = 2t + 1 shares received
	_, found, err := self.State().BatchReconStore.Get(
		msg.DPSSBatchRecDetails.ToBatchRecID(),
	)
	if err != nil {
		common.LogStateRetrieveError("PrivateRecHandler", "Process", err)
		return
	}
	if !found {
		common.LogStateNotFoundError("PrivateRecHandler", "Process", found)
		return
	}
	// Store the share in the local state.
	recState, err := self.State().BatchReconStore.UpdateBatchRecState(
		msg.DPSSBatchRecDetails.ToBatchRecID(),
		func(recState *common.BatchRecState) {
			recState.UStore[sender.Index] = share
		},
	)
	if err != nil {
		common.LogStateUpdateError("PrivateRecHandler", "Process", common.BatchRecStateType, err)
		return
	}

	countU := recState.CountReceivedU()
	_, _, t := self.Params()
	if countU >= 2*t+1 && !recState.SentPubMsg {

		UStore := recState.UStore
		doMatch, interpolatedPoly, err := common.CheckPointsLieInPoly(
			UStore,
			t,
			2*t+1,
			curve,
		)
		if err != nil {
			log.WithFields(
				log.Fields{
					"Message": "error during the checking and interpolation of the polynomial",
					"Error":   err,
				},
			).Error("PrivateRecMsg: Process")
			return
		}
		if !doMatch {
			log.WithFields(
				log.Fields{
					"Message": "shares does not coincide on the interpolating polynomial",
					"Error":   err,
				},
			).Error("PrivateRecMsg: Process")
			return
		}

		// If they coincide, send u_i to the all the parties party.
		reconstructedU := interpolatedPoly.Coefficients[0]
		publicReconstructMsg, err := NewPublicRecMsg(
			msg.DPSSBatchRecDetails,
			msg.curveName,
			reconstructedU.Bytes(),
		)
		if err != nil {
			common.LogErrorNewMessage("PrivateRecHandler", "Process", PublicRecMessageType, err)
			return
		}

		_, err = self.State().BatchReconStore.UpdateBatchRecState(
			msg.DPSSBatchRecDetails.ToBatchRecID(),
			func(state *common.BatchRecState) {
				state.SentPubMsg = true
			},
		)
		if err != nil {
			common.LogStateUpdateError("PrivateRecHandler", "Process", common.BatchRecStateType, err)
			return
		}

		// Broadcast to the old committee
		go self.Broadcast(false, *publicReconstructMsg)
	}
}
