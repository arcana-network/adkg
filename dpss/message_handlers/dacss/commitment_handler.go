package dacss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/sharing"
	"github.com/torusresearch/bijson"
)

type DacssCommitmentMessageType string

type DacssCommitmentMessage struct {
	ACSSRoundDetails common.ACSSRoundDetails
	Commitments      []common.Point
	Kind             string
	CurveName        common.CurveName
}

func NewDacssCommitmentMessage(
	acssRoundDetails common.ACSSRoundDetails,
	curve common.CurveName,
	commitments *sharing.FeldmanVerifier,
) (*common.PSSMessage, error) {
	commitmentsPoint := make([]common.Point, 0)
	for _, commitment := range commitments.Commitments {
		point := common.CurvePointToPoint(commitment, curve)
		commitmentsPoint = append(commitmentsPoint, point)
	}

	m := DacssCommitmentMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             DacssEchoMessageType,
		CurveName:        curve,
		Commitments:      commitmentsPoint,
	}

	// TODO: Check if bijison serializes []common.Point correctly.
	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, string(m.Kind), bytes)
	return &msg, nil
}

func (msg *DacssCommitmentMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

}
