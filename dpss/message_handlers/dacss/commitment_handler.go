package dacss

import (
	"encoding/hex"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var DacssCommitmentMessageType string = "dacss_commitment"

// Represents a COMMITMENT message as in Line 204, Algorithm 4, DPS paper.
type DacssCommitmentMessage struct {
	ACSSRoundDetails     common.ACSSRoundDetails // Details of the current round.
	CommitmentSecretHash []byte                  // Hash of the commitments.
	Kind                 string                  // Type of the message.
	CurveName            common.CurveName        // Curve that is being used.
}

func NewDacssCommitmentMessage(
	acssRoundDetails common.ACSSRoundDetails,
	curve common.CurveName,
	commitmentSecret curves.Point,
) (*common.PSSMessage, error) {

	// Concatenate all the commitments in a big list to compute the hash.
	commitmentSecretHash := common.HashByte(commitmentSecret.ToAffineCompressed())
	log.WithFields(
		log.Fields{
			"HashCommitmentSecret": commitmentSecret,
		},
	).Info("NewDACSSCommitmentMessage")

	m := DacssCommitmentMessage{
		ACSSRoundDetails:     acssRoundDetails,
		Kind:                 DacssCommitmentMessageType,
		CurveName:            curve,
		CommitmentSecretHash: commitmentSecretHash,
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
	).Debug("DACSSCommitmentMessage: Process")

	log.WithFields(
		log.Fields{
			"AcssRoundDetails": msg.ACSSRoundDetails.ToACSSRoundID(),
			"Message":          "trying to access the state using the acss round details",
		},
	).Debug("DacssCommitmentMessage: Process")

	self.State().AcssStore.Lock()
	defer self.State().AcssStore.Unlock()

	state, found, err := self.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error retrieving the state of the node.",
			},
		).Error("DACSSCommitmentMessage: Process")
		return
	}
	if !found {
		log.WithFields(
			log.Fields{
				"Found":   found,
				"Message": "State not found",
			},
		).Error("DACSSCommitmentMessage: Process")
		return
	}

	// Do nothing if the commitment have been already received by the sender.
	if state.ReceivedCommitments[sender.Index] {
		log.WithFields(
			log.Fields{
				"Sender":              sender.Index,
				"Received commitment": state.ReceivedCommitments[sender.Index],
				"Message":             "The commitments have been already received from this sender.",
			},
		).Debug("DACSSCommitmentMessage: Process")
		return
	}

	// Mark that the sender already sent its commitments and increase the count
	// for the received commitment.
	self.State().AcssStore.UpdateAccsState(
		msg.ACSSRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.ReceivedCommitments[sender.Index] = true
			commitmentStrEncoding := hex.EncodeToString(msg.CommitmentSecretHash)
			state.CommitmentCount[commitmentStrEncoding]++
		},
	)

	// If the RBC hasn't ended, we should not do the check afterwards
	if state.RBCState.Phase == common.Ended {
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

				log.WithFields(
					log.Fields{
						"Message":   "commitment finished correctly. Start MBVA here",
						"SelfIdx":   self.Details().Index,
						"IsNewNode": self.IsNewNode,
					},
				).Debug("DacssCommitmentMessage: process")
			}
		} else {
			log.WithFields(
				log.Fields{
					"Threshold": t + 1,
					"Message":   "There is no commitment record surpasing the threshold",
				},
			).Info("DACSSCommitmentMessage: Process")
		}
	} else {
		log.WithFields(
			log.Fields{
				"RBCState.Phase": state.RBCState.Phase,
				"Message":        "the RBC has not ended yet",
			},
		).Debug("DACSSCommitmentMessage: Process")
	}
}
