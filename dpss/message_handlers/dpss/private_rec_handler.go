package dpss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
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
		msg.Kind,
		msgBytes,
	)

	return &pssMessage, nil
}

func (msg *PrivateRecMsg) Process(sender common.NodeDetails, self common.PSSParticipant) {
	self.State().BatchReconStore.Lock()
	defer self.State().BatchReconStore.Unlock()

	// Deserialize the share.
	var share curves.Scalar
	err := bijson.Unmarshal(msg.UShare, share)
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while unmarshalling the u_i share",
			},
		).Error("PrivateRecMsg: Process")
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
	if countU > 2*t+1 {
		// Take the first t + 1 shares and interpolates the polynomial

		// Evaluate all the points in the polinomial and see if they coincide

		// If they don't coincide return error

		// If they coincide, send u_i to the respective party.

	}
}
