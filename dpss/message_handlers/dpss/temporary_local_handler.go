// NOTE: This file is for testing the public reconstruction handler
// creating a dummy msg to pass in the public_rec_handler

package dpss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/torusresearch/bijson"
)

var LocalComputationMessageType string = "dpss_local_computation"

type LocalComputationMsg struct {
	DPSSBatchRecDetails common.DPSSBatchRecDetails
	Kind                string
	curveName           common.CurveName
	coefficients        [][]byte
	UserIds             []string
}

func NewLocalComputationMsg(
	dpssBatchRecDetails common.DPSSBatchRecDetails,
	curve common.CurveName,
	coefficients [][]byte,
	userIds []string,

) (*common.PSSMessage, error) {
	msg := LocalComputationMsg{
		DPSSBatchRecDetails: dpssBatchRecDetails,
		Kind:                LocalComputationMessageType,
		curveName:           curve,
		coefficients:        coefficients,
		UserIds:             userIds,
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
