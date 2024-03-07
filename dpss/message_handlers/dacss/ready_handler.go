package dacss

import (
	"encoding/hex"

	"github.com/arcana-network/dkgnode/common"
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
func NewDacssReadyMessage(acssRoundDetails common.ACSSRoundDetails, share infectious.Share, hash []byte, curve common.CurveName, newCommittee bool) (*common.PSSMessage, error) {
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

func (m *DacssReadyMessage) Fingerprint() string {
	var bytes []byte
	delimiter := common.Delimiter2
	bytes = append(bytes, m.Hash...)
	bytes = append(bytes, delimiter...)

	bytes = append(bytes, m.Share.Data...)
	bytes = append(bytes, delimiter...)

	bytes = append(bytes, byte(m.Share.Number))
	bytes = append(bytes, delimiter...)
	hash := hex.EncodeToString(common.Keccak256(bytes))
	return hash
}

// Algorithm 4: https://eprint.iacr.org/2021/777.pdf
// line 13

/*
upon receiving 𝑡 + 1 ⟨READY, ∗, ℎ⟩ messages and not having sent a READY message do

    Wait for 𝑡 + 1 matching ⟨ECHO,𝑚′𝑖, ℎ⟩
    send ⟨READY,𝑚′𝑖, ℎ⟩ to all
*/

func (m *DacssReadyMessage) Process(sender common.NodeDetails, p common.PSSParticipant) {

	//TODO: cannot identlfy the old/new nodes just by index
	log.Debugf("Received Ready message from sender=%d on %d", sender.Index, p.Details().Index)

	// Get state from node
	state, isStored, err := p.State().AcssStore.Get(m.AcssRoundDetails.ToACSSRoundID())

	if err != nil {
		log.WithField("error", err).Error("DacssReadyMessage - Process()")
		return
	}

	if !isStored {
		log.WithField("error", "ACSS state not stored yet").Error("DacssEchoMessage - Process()")
		return
	}

	p.State().AcssStore.Lock()
	defer p.State().AcssStore.Unlock()

	// Make sure the ready msg received from a node is set to true
	defer func() {
		state.RBCState.ReceivedReady[sender.Index] = true
	}()

	// returns if RBC ended
	if state.RBCState.Phase == common.Ended {
		return
	}

	receivedReady, found := state.RBCState.ReceivedReady[sender.Index]
	if found && receivedReady {
		log.Debugf("Already received ready for %s from %d on %d", m.AcssRoundDetails.ToACSSRoundID(), sender.Index, p.Details().Index)
		return
	}

	// ownShare := state.RBCState.OwnReedSolomonShard
	// ownHash := state.RBCState.HashMsg

	readyCount := state.RBCState.CountReady()

	_, t, _ := p.Params()

	//check if t+1 Ready msg received and not send ready msg
	if readyCount >= t+1 && !state.RBCState.IsReadyMsgSent {

		// TODO: Check this
		// Since ReceivedEco map is set to true in the echo handler only when the there is a matching RS shares data
		// so it is sufficient to check the count
		if state.RBCState.CountEcho() >= t+1 {
			readyMsg, err := NewDacssReadyMessage(m.AcssRoundDetails, m.Share, m.Hash, m.CurveName, p.IsOldNode())

			if err != nil {
				log.WithField("error", err).Error("DacssReadyMessage - Process()")
				return
			}

			p.State().AcssStore.UpdateAccsState(
				m.AcssRoundDetails.ToACSSRoundID(),
				func(state *common.AccsState) {
					state.RBCState.IsReadyMsgSent = true
				},
			)

			go p.Broadcast(p.IsOldNode(), *readyMsg)
		}

		//------OLD CODE------
		// if c.RC >= k && !c.ReadySent && c.EC >= k {
		// 	// Broadcast ready message
		// 	readyMsg := NewAcssReadyMessage(m.roundID, m.share, m.hash, m.curve, p.ID(), m.newCommittee)
		// 	go p.Broadcast(m.newCommittee, readyMsg)
		// }

		// for i := 0; i < f; i += 1 {
		// 	log.Debugf("len(readstore)=%d, threshold=%d", len(keygen.ReadyStore), (2*f + 1 + i))
		// 	if len(keygen.ReadyStore) >= (2*f + 1 + i) {
		// 		// Create RS encoding
		// 		f, err := infectious.NewFEC(k, n)
		// 		if err != nil {
		// 			log.Debugf("error during creation of fec, err=%s", err)
		// 			return
		// 		}

		// 		M, err := acss.Decode(f, keygen.ReadyStore)
		// 		if err != nil {
		// 			log.Debugf("Decode faced an error, err=%s", err)
		// 			return
		// 		}
		// 		hash := common.Hash(M)
		// 		log.Debugf("HashCompare, hash=%v, mHash=%v", hash, m.hash)

		// 		if bytes.Equal(hash, m.hash) {
		// 			outputMsg := NewAcssOutputMessage(m.roundID, M, m.curve, p.ID(), "ready", m.newCommittee)
		// 			go p.ReceiveMessage(outputMsg)
		// 			defer func() { keygen.State.Phase = common.Ended }()
		// 			// send to other committee
		// 			msg := messages.MessageData{}
		// 			err := msg.Deserialize(M)
		// 			if err != nil {
		// 				log.Debugf("Could not deserialize message data, err=%s", err)
		// 				return
		// 			}
		// 			for _, n := range p.Nodes(!m.newCommittee) {
		// 				go func(node common.DkgParticipant) {
		// 					readyMsg := NewAcssCommitMessage(m.roundID, msg.Commitments, m.curve, p.ID(), m.newCommittee)
		// 					p.Send(readyMsg, node)
		// 				}(n)
		// 			}
		// 		}

		// 	}
		// }
	}
}
