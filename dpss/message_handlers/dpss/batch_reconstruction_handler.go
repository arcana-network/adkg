package dpss

import "github.com/arcana-network/dkgnode/common"

type PssBatchReconstructionMessage struct {
}

func NewPssBatchReconstructionMessage(
	pssRoundDetails common.PSSRoundDetails,
	rValues []byte,
	curveName common.CurveName,
) (*common.PSSMessage, error) {
	return &common.PSSMessage{}, nil
}
