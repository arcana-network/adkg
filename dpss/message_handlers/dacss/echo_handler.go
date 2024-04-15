package dacss

import (
	"encoding/hex"

	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
	"github.com/vivint/infectious"
)

var DacssEchoMessageType string = "dacss_echo"

// DacssEchoMessage represents the echo handler in the RBC protocol
type DacssEchoMessage struct {
	ACSSRoundDetails common.ACSSRoundDetails // Details of the current row
	Kind             string                  // Type of the message
	CurveName        common.CurveName        // Curve used for the computation
	Share            infectious.Share        // Shard comming from the RS Encoding.
	Hash             []byte                  // Hash of the shares.
}

// NewDacssEchoMessage creates an ECHO message in the RBC protocol.
func NewDacssEchoMessage(acssRoundDetails common.ACSSRoundDetails, share infectious.Share, hash []byte, curve common.CurveName, newCommittee bool) (*common.PSSMessage, error) {
	m := DacssEchoMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             DacssEchoMessageType,
		CurveName:        curve,
		Share:            share,
		Hash:             hash,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, string(m.Kind), bytes)
	return &msg, nil
}

func (msg *DacssEchoMessage) Fingerprint() string {
	var bytes []byte
	delimiter := common.Delimiter2
	bytes = append(bytes, msg.Hash...)
	bytes = append(bytes, delimiter...)

	bytes = append(bytes, msg.Share.Data...)
	bytes = append(bytes, delimiter...)

	bytes = append(bytes, byte(msg.Share.Number))
	bytes = append(bytes, delimiter...)
	hash := hex.EncodeToString(common.Keccak256(bytes))
	return hash
}

// Process handles the incomming ECHO message.
func (m DacssEchoMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.Debugf("Echo received: Sender=%d, Receiver=%d", sender.Index, self.Details().Index)
	// if sender.Index == self.Details().Index {
	// 	log.WithFields(
	// 		log.Fields{
	// 			"SenderIdx": sender.Index,
	// 			"SelfIdx":   self.Details().Index,
	// 			"Message":   "Ignore the echo message from self.",
	// 		},
	// 	).Error("DACSSEchoMessage: Process")
	// 	return // TODO check whether the Echo should indeed be ignored when coming from self
	// }

	self.State().AcssStore.Lock()

	// Using defer because the ACSS state is used until the end of the protocol.
	defer self.State().AcssStore.Unlock()

	acssState, isStored, err := self.State().AcssStore.Get(m.ACSSRoundDetails.ToACSSRoundID())
	if err != nil {
		common.LogStateRetrieveError("DacssEchoMessage", "Process", err)
		return
	}
	if isStored {
		// If the ECHO was already received, do nothing.
		receivedEcho, echoFound := acssState.RBCState.ReceivedEcho[sender.Index]
		if echoFound && receivedEcho {
			log.Debugf("Already received echo from %d", sender.Index)
			return
		}
	}

	// Check that the incoming message matches with the share of self (Line 11)
	// of Algorithm 4, "Asynchronous Data Disemination".

	// If the ECHO message has been not received, then update the received ECHO
	// and increase the counter.
	// In case the node didn't have state for this acssRound, it will initiate one
	err = self.State().AcssStore.UpdateAccsState(
		m.ACSSRoundDetails.ToACSSRoundID(),
		func(state *common.AccsState) {
			state.RBCState.ReceivedEcho[sender.Index] = true

			echoStore := state.RBCState.GetEchoStore(
				m.Fingerprint(),
				m.Hash,
				m.Share,
			)
			echoStore.Count++
		},
	)
	if err != nil {
		common.LogStateUpdateError("EchoHandler", "Process", common.AcssStateType, err)
		return
	}

	_, _, t := self.Params()

	acssState, found, err := self.State().AcssStore.Get(m.ACSSRoundDetails.ToACSSRoundID())
	if err != nil {
		common.LogStateRetrieveError("DacssEchoMessage", "Process", err)
		return
	}
	if !found {
		common.LogStateNotFoundError("DacssEchoMessage", "Proces", found)
	}
	// This deals with Line 11 of the RBC protocol. If the ECHO count for the
	// received message is 2t + 1, then send the READY message.
	msgRegistry := acssState.RBCState.GetEchoStore(
		m.Fingerprint(),
		m.Hash,
		m.Share,
	)

	if msgRegistry.Count >= 2*t+1 && !acssState.RBCState.IsReadyMsgSent {
		readyMsg, err := NewDacssReadyMessage(m.ACSSRoundDetails, m.Share, m.Hash, m.CurveName)
		if err != nil {
			common.LogErrorNewMessage("DacssEchoMessage", "Process", AcssReadyMessageType, err)
			return
		}
		err = self.State().AcssStore.UpdateAccsState(
			m.ACSSRoundDetails.ToACSSRoundID(),
			func(state *common.AccsState) {
				state.RBCState.IsReadyMsgSent = true
			},
		)
		if err != nil {
			common.LogStateUpdateError("EchoHandler", "Process", common.AcssStateType, err)
			return
		}

		go self.Broadcast(self.IsNewNode(), *readyMsg)
	}

	// This deals with the waiting for ECHO handler in Line 14 of the RBC
	// protocol.
	if acssState.RBCState.CountReady() >= t+1 && msgRegistry.Count >= t+1 {
		readyMsg, err := NewDacssReadyMessage(m.ACSSRoundDetails, msgRegistry.Shard, m.Hash, m.CurveName)
		if err != nil {
			common.LogErrorNewMessage("DacssEchoMessage", "Process", AcssReadyMessageType, err)
			return
		}
		err = self.State().AcssStore.UpdateAccsState(
			m.ACSSRoundDetails.ToACSSRoundID(),
			func(state *common.AccsState) {
				state.RBCState.IsReadyMsgSent = true
			},
		)
		if err != nil {
			common.LogStateUpdateError("EchoHandler", "Process", common.AcssStateType, err)
			return
		}

		go self.Broadcast(self.IsNewNode(), *readyMsg)
	}
}
