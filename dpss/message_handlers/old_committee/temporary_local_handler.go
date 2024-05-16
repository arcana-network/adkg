// NOTE: This file is for testing the public reconstruction handler
// creating a dummy msg to pass in the public_rec_handler

package old_committee

import (
	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
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

func (msg *LocalComputationMsg) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.Info("LocalComputationMsg: Process")
}
