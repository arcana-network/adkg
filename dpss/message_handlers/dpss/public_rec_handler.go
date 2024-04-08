package dpss

import (
	"github.com/arcana-network/dkgnode/common"
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
