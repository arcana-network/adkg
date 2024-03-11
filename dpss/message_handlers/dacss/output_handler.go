package dacss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/torusresearch/bijson"
)

var DacssOutputMessageType common.MessageType = "dacss_output"

type DacssOutputMessage struct {
	AcssRoundDetails common.ACSSRoundDetails
	kind             common.MessageType
	curveName        common.CurveName
	m                []byte
}

func NewDacssOutputMessage(roundDetails common.ACSSRoundDetails, data []byte, curveName common.CurveName) (*common.PSSMessage, error) {
	m := DacssOutputMessage{
		AcssRoundDetails: roundDetails,
		kind:             DacssOutputMessageType,
		curveName:        curveName,
		m:                data,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.AcssRoundDetails.PSSRoundDetails, string(m.kind), bytes)
	return &msg, nil
}
