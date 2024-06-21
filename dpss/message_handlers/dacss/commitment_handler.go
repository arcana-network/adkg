package dacss

import (
	"encoding/binary"
	"encoding/hex"
	"math"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/keyset"
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

	// Use defer because the state is needed until the end of the function.
	defer self.State().AcssStore.Unlock()

	state, found, err := self.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	if err != nil {
		common.LogStateRetrieveError("DacssCommitmentHandler", "Process", err)
		return
	}
	if !found {
		common.LogStateNotFoundError("DacssCommitmentHandler", "Process", found)
		return
	}

	if state.ValidShareOutput {
		log.WithFields(
			log.Fields{
				"Message": "The share output is already marked valid",
			},
		).Debug("DACSSCommitmentMessage: Process")
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
	state, err = self.State().AcssStore.UpdateAccsState(
		msg.ACSSRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.ReceivedCommitments[sender.Index] = true
			commitmentStrEncoding := hex.EncodeToString(msg.CommitmentSecretHash)
			state.CommitmentCount[commitmentStrEncoding]++
		},
	)
	if err != nil {
		common.LogStateUpdateError("DACSSCommitmentMessage", "Process", common.AcssStateType, err)
		return
	}
	n, _, t := self.Params()
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
					"Message": "Start MVBA",
				},
			).Debug("DACSSCommitmentMessage: Process")

			// DONE: Call the message to start MVBA here.
			{
				if self.IsNewNode() {
					return
				}
				pssState, complete := self.State().PSSStore.GetOrSetIfNotComplete(msg.ACSSRoundDetails.PSSRoundDetails.PssID)
				if complete {
					return
				}
				pssState.Lock()
				defer pssState.Unlock()

				if pssState.KeysetProposed {
					return
				}
				numShares := msg.ACSSRoundDetails.PSSRoundDetails.BatchSize
				alpha := int(math.Ceil(float64(numShares) / float64((n - 2*t))))
				TSet, completed := pssState.CheckForThresholdCompletion(alpha, n-t)
				if completed {
					pssState.T[self.Details().Index] = TSet
					// FIXME: Calling below function to trigger listeners, can change func name
					pssState.GetTSet(n, t)
					pssState.KeysetProposed = true
					var output [8]byte
					binary.BigEndian.PutUint64(output[:], uint64(TSet))

					round := common.PSSRoundDetails{
						PssID:     msg.ACSSRoundDetails.PSSRoundDetails.PssID,
						Dealer:    self.Details(),
						BatchSize: msg.ACSSRoundDetails.PSSRoundDetails.BatchSize,
					}
					msg, err := keyset.NewProposeMessage(round, output[:], msg.CurveName)
					if err != nil {
						log.Errorf("Error while creating keyset propose: %v", err)
						return
					}
					go self.Broadcast(false, *msg)
				}
			}

		}
	} else {
		log.WithFields(
			log.Fields{
				"Threshold": t + 1,
				"Message":   "No threshold commitment found",
			},
		).Debug("DACSSCommitmentMessage: Process")
	}
}
