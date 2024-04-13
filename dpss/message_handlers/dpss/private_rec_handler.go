package dpss

import (
	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var PrivateRecHandlerType string = "dpss_private_rec"

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
		Kind:                PrivateRecHandlerType,
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
	defer self.State().BatchReconStore.Unlock()

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

	// Store the share in the local state.
	self.State().BatchReconStore.UpdateBatchRecState(
		msg.DPSSBatchRecDetails.ToBatchRecID(),
		func(recState *common.BatchRecState) {
			recState.UStore[sender.Index] = share
		},
	)

	// Check if there are at least d + t + 1 = 2t + 1 shares received
	recState, found, err := self.State().BatchReconStore.Get(
		msg.DPSSBatchRecDetails.ToBatchRecID(),
	)
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error trying to retrieve the batch reconstruction state",
			},
		).Error("PrivateRecMsg: Process")
		return
	}
	if !found {
		log.WithFields(
			log.Fields{
				"Found":   found,
				"Message": "There is no state associated with the provided ID",
			},
		).Error("PrivateRecMsg: Process")
		return
	}

	countU := recState.CountReceivedU()
	_, _, t := self.Params()
	if countU >= 2*t+1 {

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
					"Message": "shares does not coincide on the interpolationg polynomial",
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
			log.WithFields(
				log.Fields{
					"Error":   err,
					"Message": "Error constructiong Public Reconstruction msg",
				},
			).Error("PrivateRecMsg: Process")
			return
		}

		// Broadcast to the old committee
		go self.Broadcast(false, *publicReconstructMsg)
	}
}
