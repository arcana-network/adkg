package dacss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/sharing"
	log "github.com/sirupsen/logrus"
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
		log.WithFields(log.Fields{
			"Error":   err,
			"Message": "Error while converting the message into bytes",
		}).Error("DACSSCommitmentMessage: NewDacssCommitmentMessage")
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, string(m.Kind), bytes)
	return &msg, nil
}

func (msg *DacssCommitmentMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	state, found, err := self.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	if !found {
		log.WithFields(
			log.Fields{
				"Found":   found,
				"Message": "State not found",
			},
		).Error("DACSSCommitmentMessage: Process")
		return
	}
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error retrieving the state of the node.",
			},
		).Error("DACSSCommitmentMessage: Process")
		return
	}

	self.State().AcssStore.Lock()
	defer self.State().AcssStore.Unlock()

	// Do nothing if the commitment have been already received by the sender.
	if state.ReceivedCommitments[sender.Index] {
		log.WithFields(
			log.Fields{
				"Sender":              sender.Index,
				"Received commitment": state.ReceivedCommitments[sender.Index],
				"Message":             "The commitments have been already received from this sender.",
			},
		).Info("DACSSCommitmentMessage: Process")
		return
	}

	// Mark that the sender already sent its commitments and increase the count
	// for the received commitment.
	commitmentSerialization := common.SerializePointCommitments(msg.Commitments)
	commitmentDb := state.GetStoreForCommitment(
		commitmentSerialization,
		msg.Commitments,
	)
	state.ReceivedCommitments[sender.Index] = true
	commitmentDb.Count++

	_, _, t := self.Params()
	commitmentInfo := state.FindThresholdCommitment(t + 1)
	if commitmentInfo != nil {
		commitmentsMatch := true
		for i, receivedCommitment := range msg.Commitments {
			if !receivedCommitment.Equal(commitmentInfo.Commitments[i]) {
				commitmentsMatch = false
			}
		}

		if commitmentsMatch {
			self.State().AcssStore.UpdateAccsState(
				msg.ACSSRoundDetails.ToACSSRoundID(),
				func(state *common.AccsState) {
					state.ValidShareOutput = true
				},
			)

			// TODO: Call the message to start MVBA here.
		}
	} else {
		log.WithFields(
			log.Fields{
				"Threshold": t + 1,
				"Message":   "There is no commitment record surpasing the threshold",
			},
		).Info("DACSSCommitmentMessage: Process")
	}
}
