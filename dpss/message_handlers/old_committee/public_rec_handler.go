package old_committee

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/new_committee"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var PublicRecMessageType string = "dpss_public_rec"

type PublicRecMsg struct {
	DPSSBatchRecDetails common.DPSSBatchRecDetails
	Kind                string
	curveName           common.CurveName
	ReconstructedUShare []byte
}

func NewPublicRecMsg(
	dpssBatchRecDetails common.DPSSBatchRecDetails,
	curve common.CurveName,
	uShare []byte,
) (*common.PSSMessage, error) {
	msg := PublicRecMsg{
		DPSSBatchRecDetails: dpssBatchRecDetails,
		Kind:                PublicRecMessageType,
		curveName:           curve,
		ReconstructedUShare: uShare,
	}

	msgBytes, err := bijson.Marshal(msg)
	if err != nil {
		return nil, err
	}

	pssMessage := common.CreatePSSMessage(
		msg.DPSSBatchRecDetails.PSSRoundDetails,
		msg.Kind,
		msgBytes,
	)

	return &pssMessage, nil
}

func (msg *PublicRecMsg) Process(sender common.NodeDetails, self common.PSSParticipant) {
	self.State().BatchReconStore.Lock()

	// Holding the lock until the end given that the Batch Rec. store is
	// accessed until the end of the function.
	defer self.State().BatchReconStore.Unlock()

	// Check if there are at least T + t = n - t shares received
	recState, found, err := self.State().BatchReconStore.Get(
		msg.DPSSBatchRecDetails.ToBatchRecID(),
	)
	if err != nil {
		common.LogStateRetrieveError("PublicRecHandler", "Process", err)
		return
	}
	if !found {
		common.LogStateNotFoundError("PublicRecHandler", "Process", found)
		return
	}

	// Deserialize the share.
	curve := common.CurveFromName(msg.curveName)
	share, err := curve.Scalar.SetBytes(msg.ReconstructedUShare)
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while converting the share from bytes to scalar",
			},
		).Error("PublicRecMsg: Process")
		return
	}

	// Store the reconstructedU in the local state.
	err = self.State().BatchReconStore.UpdateBatchRecState(
		msg.DPSSBatchRecDetails.ToBatchRecID(),
		func(recState *common.BatchRecState) {
			recState.ReconstructedUStore[sender.Index] = share
		},
	)
	if err != nil {
		common.LogStateUpdateError("PublicRecHandler", "Process", common.BatchRecStateType, err)
		return
	}

	ReconstructedUCount := recState.CountReconstructedReceivedU()

	n, _, t := self.Params()
	T := n - 2*t

	// upon receiving at least T+t u_k, follow the next procedure:
	// 	Take the first T + 1 u_k values and interpolate a polynomial.
	// 	Check that the rest of the u_k values not used in the interpolation lie in the polynomial.
	// 	If all the u_k values lie in the polynomial, return the coefficients of the polynomial, otherwise return unhappy.

	if ReconstructedUCount >= T+t && !recState.SentLocalCompMsg {
		ReconstructedUStore := recState.ReconstructedUStore

		doMatch, interpolatePoly, err := common.CheckPointsLieInPoly(
			ReconstructedUStore,
			T-1,
			T+t,
			curve,
		)
		if err != nil {
			log.WithFields(
				log.Fields{
					"Message": "error checking that T + t points are in a polynomial of degree T - 1",
					"Error":   err,
				},
			).Error("PublicRecMsg: Process")
			return
		}
		if !doMatch {
			log.WithFields(
				log.Fields{
					"Message": "there are on T + t points that lie in a polynomial of degree T - 1",
				},
			).Error("PublicRecMsg: Process")
			return
		}

		// sharing the coefficients
		len := len(interpolatePoly.Coefficients)
		polynomialCoefficient := make([][]byte, len)

		for i := 0; i < len; i++ {
			polynomialCoefficient[i] = make([]byte, 32)
			polynomialCoefficient[i] = interpolatePoly.Coefficients[i].Bytes()
		}

		pssState, _ := self.State().PSSStore.Get(msg.DPSSBatchRecDetails.PSSRoundDetails.PssID)
		pssState.Lock()
		self.State().ShareStore.Lock()

		T := pssState.GetTSet(n, t)
		localComputationMsg, err := new_committee.NewLocalComputationMsg(
			msg.DPSSBatchRecDetails,
			msg.curveName,
			polynomialCoefficient,
			T,
			self.State().ShareStore.GetUserIDs(),
		)
		self.State().ShareStore.Unlock()
		pssState.Unlock()

		if err != nil {
			common.LogErrorNewMessage("PublicRecHandler", "Process", new_committee.LocalComputationMessageType, err)
			return
		}

		err = self.State().BatchReconStore.UpdateBatchRecState(
			msg.DPSSBatchRecDetails.ToBatchRecID(),
			func(state *common.BatchRecState) {
				state.SentLocalCompMsg = true
			},
		)
		if err != nil {
			common.LogStateUpdateError("PublicRecHandler", "Process", common.BatchRecStateType, err)
			return
		}

		// Broadcast to the new committee
		go self.Broadcast(true, *localComputationMsg)

		// Clear the state before starting the new batch
		err = self.State().Clean(msg.DPSSBatchRecDetails.PSSRoundDetails)
		if err != nil {
			log.WithFields(
				log.Fields{
					"Message": "error while cleanning the state",
					"Error":   err,
				},
			).Error("PublicRecMsgHandler: Process")
			return
		}

		// starting the next batch
		messageBroker := self.GetMessageBroker()
		if messageBroker != nil {
			messageBroker.PssMethods().StartNextPSSBatch()
		}

	}
}
