package dacss

import (
	"reflect"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
	"github.com/vivint/infectious"
)

var AcssReadyMessageType string = "dacss_ready"

// Stores the information for the READY message in the RBC protocol.
type DacssReadyMessage struct {
	AcssRoundDetails common.ACSSRoundDetails
	Kind             string
	CurveName        common.CurveName
	Share            infectious.Share
	Hash             []byte
}

// ⟨READY, *, h⟩ msg in the RBC protocol
func NewDacssReadyMessage(acssRoundDetails common.ACSSRoundDetails, share infectious.Share, hash []byte, curve common.CurveName) (*common.PSSMessage, error) {
	m := DacssReadyMessage{
		Kind:             AcssReadyMessageType,
		CurveName:        curve,
		Share:            share,
		Hash:             hash,
		AcssRoundDetails: acssRoundDetails,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.AcssRoundDetails.PSSRoundDetails, string(m.Kind), bytes)
	return &msg, nil
}

// Algorithm 4: https://eprint.iacr.org/2021/777.pdf
// line 13

/*
upon receiving 𝑡 + 1 ⟨READY, ∗, ℎ⟩ messages and not having sent a READY message do

    Wait for 𝑡 + 1 matching ⟨ECHO,𝑚′𝑖, ℎ⟩
    send ⟨READY,𝑚′𝑖, ℎ⟩ to all
*/

// Process handles the incomming READY message.
// READY from self is taken into account
func (m *DacssReadyMessage) Process(sender common.NodeDetails, p common.PSSParticipant) {

	//TODO: cannot identlfy the old/new nodes just by index
	log.Debugf("Received Ready message from sender=%d on %d", sender.Index, p.Details().Index)

	p.State().AcssStore.Lock()

	// Using defer given that  the state is used until the end of the function.
	defer p.State().AcssStore.Unlock()

	// Get state from node
	state, found, err := p.State().AcssStore.Get(m.AcssRoundDetails.ToACSSRoundID())

	if err != nil {
		common.LogStateRetrieveError("DacssReadyHandler", "Process", err)
		return
	}

	// If the ready message from sender was already received, we do an early
	// return.
	if found && state.RBCState.ReceivedReady[sender.Index] {
		log.WithFields(
			log.Fields{
				"SenderIndex": sender.Index,
				"Message":     "The party already received message from this node",
			},
		).Debug("DacssReadyMessage: Process")
		return
	}

	// Adds this share to the list of READY message shards.
	state, err = p.State().AcssStore.UpdateAccsState(
		m.AcssRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.RBCState.ReadyMsgShards = append(
				state.RBCState.ReadyMsgShards,
				m.Share,
			)
		},
	)
	if err != nil {
		common.LogStateUpdateError("ReadyHandler", "Process", common.AcssStateType, err)
		return
	}

	// Returns if RBC ended
	if state.RBCState.Phase == common.Ended {
		return
	}

	// Make sure the ready msg received from a node is set to true. We need to
	// make sure also that the hashes match to increase the count.
	ownHash := state.AcssDataHash
	if reflect.DeepEqual(ownHash, m.Hash) {
		state.RBCState.ReceivedReady[sender.Index] = true
	}

	readyCount := state.RBCState.CountReady()

	n, _, t := p.Params()

	// Check if t+1 Ready msg received and not send ready msg
	if readyCount >= t+1 && !state.RBCState.IsReadyMsgSent {

		// TODO: Check this
		// Since ReceivedEcho map is set to true in the echo handler only when the there is a matching RS shares data
		// so it is sufficient to check the count
		echoInfo := state.RBCState.FindThresholdEchoMsg(t + 1)
		if echoInfo != nil {
			readyMsg, err := NewDacssReadyMessage(m.AcssRoundDetails, echoInfo.Shard, echoInfo.HashMessage, m.CurveName)

			if err != nil {
				common.LogErrorNewMessage("DacssReadyMessageHandler", "Propose", AcssReadyMessageType, err)
				return
			}

			state, err = p.State().AcssStore.UpdateAccsState(
				m.AcssRoundDetails.ToACSSRoundID(),
				func(state *common.AccsState) {
					state.RBCState.IsReadyMsgSent = true
				},
			)
			if err != nil {
				common.LogStateUpdateError("ReadyHandler", "Process", common.AcssStateType, err)
				return
			}

			go p.Broadcast(p.IsNewNode(), *readyMsg)
		}
	}

	for r := range t + 1 {
		if len(state.RBCState.ReadyMsgShards) >= 2*t+r+1 {
			// Creates RC encoding to reconstruct the message
			fec, err := infectious.NewFEC(t+1, n)
			if err != nil {
				log.WithField("error", err).Error("could not create the decoder")
				return
			}

			// Reconstruction of the message using RS encoding.
			rbcMsg, err := acss.Decode(fec, state.RBCState.ReadyMsgShards)
			if err != nil {
				log.WithField("error", err).Error("unable to decode the message")
				return
			}

			hashReconstMsg := common.HashByte(rbcMsg)
			if reflect.DeepEqual(hashReconstMsg, state.AcssDataHash) {
				// Send output msg to self
				outputMsg, err := NewDacssOutputMessage(m.AcssRoundDetails, rbcMsg, m.CurveName)

				if err != nil {
					common.LogErrorNewMessage("DacssOutputHandler", "Process", DacssOutputMessageType, err)
					return
				}
				go p.Send(p.Details(), *outputMsg)
			}
		}
	}
}
