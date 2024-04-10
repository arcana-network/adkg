package dpss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var PublicRecHandlerType string = "dpss_public_rec"

type PublicRecMsg struct {
	DPSSBatchRecDetails common.DPSSBatchRecDetails
	Kind                string
	curveName           common.CurveName
	UShare              []byte
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

// TODO: Implement.
func (msg *PublicRecMsg) Process(sender common.NodeDetails, self common.PSSParticipant) {
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
	//TODO: needs to confirm this T + t value
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

		TPlus1Share := make(map[int]curves.Scalar)
		remainingShare := make(map[int]curves.Scalar)

		// 	Take the first T + 1 u_k values and interpolate a polynomial.
		count := 0
		for key, value := range ReconstructedUStore {
			if count < t+1 {
				TPlus1Share[key] = value
			} else {
				remainingShare[key] = value
			}
			count++
		}

		interpolatePoly, err := common.InterpolatePolynomial(TPlus1Share, curve)

		if err != nil {
			log.WithFields(
				log.Fields{
					"Error":   err,
					"Message": "Error trying to interpolate ploynomial with t+1 shares",
				},
			).Error("PublicRecMsg: Process")
			return
		}

		for key, value := range remainingShare {
			keyScalar := curve.Scalar.New(key)
			evaluationResult := interpolatePoly.Evaluate(keyScalar)

			// If the evaluation doesn't coincide return error
			if evaluationResult.Cmp(value) != 0 {

				log.WithFields(
					log.Fields{
						"Message": "shares does not coincide on the interpolationg polynomial",
					},
				).Error("PublicRecMsg: Process")
				return
			}
		}

		//TODO: Send the msg to the next handler

	}
}
