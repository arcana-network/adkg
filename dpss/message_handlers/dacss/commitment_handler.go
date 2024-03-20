package dacss

import (
	"encoding/hex"

	"github.com/arcana-network/dkgnode/common"
	arcanasharing "github.com/arcana-network/dkgnode/common/sharing"
	"github.com/coinbase/kryptology/pkg/sharing"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var DacssCommitmentMessageType string = "dacss_commitment"

// Represents a COMMITMENT message as in Line 204, Algorithm 4, DPS paper.
type DacssCommitmentMessage struct {
	ACSSRoundDetails common.ACSSRoundDetails // Details of the current round.
	CommitmentsHash  []byte                  // Hash of the commitments.
	Kind             string                  // Type of the message.
	CurveName        common.CurveName        // Curve that is being used.
}

func NewDacssCommitmentMessage(
	acssRoundDetails common.ACSSRoundDetails,
	curve common.CurveName,
	commitments *sharing.FeldmanVerifier,
) (*common.PSSMessage, error) {

	// Concatenate all the commitments in a big list to compute the hash.
	concatCommitments := arcanasharing.CompressCommitments(commitments)
	commitmentsHash := common.HashByte(concatCommitments)
	log.WithFields(
		log.Fields{
			"ConcatCommitments": concatCommitments,
			"HashCommitments":   commitmentsHash,
		},
	).Info("NewDACSSCommitmentMessage")

	m := DacssCommitmentMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             DacssCommitmentMessageType,
		CurveName:        curve,
		CommitmentsHash:  commitmentsHash,
	}

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

// Processes the reception of a COMMITMENT message. See Line 204, Algorithm 4, DPS paper.
func (msg *DacssCommitmentMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.WithFields(
		log.Fields{
			"Sender":   sender.Index,
			"Receiver": self.Details().Index,
			"Message":  "Received Commitment message",
		},
	).Info("DACSSCommitmentMessage: Process")

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
	state.ReceivedCommitments[sender.Index] = true
	commitmentStrEncoding := hex.EncodeToString(msg.CommitmentsHash)
	state.CommitmentCount[commitmentStrEncoding]++

	_, _, t := self.Params()
	commitmentHexHash, found := state.FindThresholdCommitment(t + 1)
	if found {
		// Computes the hash of the own commitment
		if commitmentHexHash == state.OwnCommitmentsHash {
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
