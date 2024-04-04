package dpss

import "github.com/arcana-network/dkgnode/common"

var PrivateRecHandlerType string = "dpss_private_rec"

type PrivateRecMsg struct {
	DPSSBatchRecDetails common.DPSSBatchRecDetails
	Kind                string
	curveName           common.CurveName
	uShare              []byte
}

func NewPrivateRecMsg(
	dpssBatchRecDetails common.DPSSBatchRecDetails,
	curve common.CurveName,
	uShare []byte,
) (*common.PSSMessage, error) {
	return &common.PSSMessage{}, nil
}
