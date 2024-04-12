package dpss

import (
	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var PublicRecHandlerType string = "dpss_public_rec"

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
		msg.Kind,
		msgBytes,
	)

	return &pssMessage, nil
}

func (msg *PublicRecMsg) Process(sender common.NodeDetails, self common.PSSParticipant) {
	self.State().BatchReconStore.Lock()
	defer self.State().BatchReconStore.Unlock()

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
	self.State().BatchReconStore.UpdateBatchRecState(
		msg.DPSSBatchRecDetails.ToBatchRecID(),
		func(recState *common.BatchRecState) {
			recState.ReconstructedUStore[sender.Index] = share
		},
	)

	// Check if there are at least T + t = n - t shares received
	recState, found, err := self.State().BatchReconStore.Get(
		msg.DPSSBatchRecDetails.ToBatchRecID(),
	)
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error trying to retrieve the batch reconstruction state",
			},
		).Error("PublicRecMsg: Process")
		return
	}
	if !found {
		log.WithFields(
			log.Fields{
				"Found":   found,
				"Message": "There is no state associated with the provided ID",
			},
		).Error("PublicRecMsg: Process")
		return
	}

	ReconstructedUCount := recState.CountReconstructedReceivedU()

	n, _, t := self.Params()
	T := n - 2*t

	// upon receiving at least T+t u_k, follow the next procedure:
	// 	Take the first T + 1 u_k values and interpolate a polynomial.
	// 	Check that the rest of the u_k values not used in the interpolation lie in the polynomial.
	// 	If all the u_k values lie in the polynomial, return the coefficients of the polynomial, otherwise return unhappy.

	if ReconstructedUCount >= T+t {
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

		//sending msg to new committee
		for _, node := range self.Nodes(!self.IsNewNode()) {
			localComputationMsg, err := NewLocalComputationMsg(msg.DPSSBatchRecDetails, msg.curveName, polynomialCoefficient)

			if err != nil {
				log.WithFields(
					log.Fields{
						"Error":   err,
						"Message": "Error constructiong local Reconstruction msg",
					},
				).Error("PublicRecMsg: Process")
				return
			}
			go self.Send(node, *localComputationMsg)
		}

	}
}
