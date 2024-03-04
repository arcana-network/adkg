package dacss

import (
	"encoding/hex"
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/vivint/infectious"
)

var AcssReadyMessageType common.MessageType = "dacss_ready"

// Stores the information for the READY message in the RBC protocol.
type DacssReadyMessage struct {
	AcssRoundDetails common.ACSSRoundDetails
	Kind             common.MessageType
	Curve            *curves.Curve
	Share            infectious.Share
	Hash             []byte
}

func NewDacssReadyMessage(acssRoundDetails common.ACSSRoundDetails, share infectious.Share, hash []byte, curve *curves.Curve) (*common.PSSMessage, error) {
	m := DacssReadyMessage{
		Kind:             AcssReadyMessageType,
		Curve:            curve,
		Share:            share,
		Hash:             hash,
		AcssRoundDetails: acssRoundDetails,
	}

	bytes, err := json.Marshal(m)
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

//TODO: Process function

// func (m *DacssReadyMessage) Process(sender common.NodeDetails, p common.PSSParticipant) {
// 	log.Debugf("Received Ready message from %d on %d", sender, p.Details().Index)

// 	// Get state from node
// 	//TODO: needs to confirm the state
// 	state := p.State().ShareStore

// 	// Create default shareStore
// 	//Question: do we need to add more fields in the PSSSshareStore?
// 	defaultShare := &common.PSSShareStore{
// 		Shares: make(map[int]curves.Scalar),
// 	}

// 	// Get or set if it doesn't exist
// 	keygen, complete := state.GetOrSetIfNotComplete(m.RoundID, defaultShare)
// 	if complete {
// 		// if keygen is complete, ignore and return
// 		return
// 	}
// 	keygen.Lock()
// 	defer keygen.Unlock()
// 	// Make sure the echo received from a node is set to true
// 	defer func() { keygen.State.ReceivedReady[m.Sender()] = true }()

// 	if keygen.State.Phase == common.Ended {
// 		return
// 	}

// 	receivedReady, found := keygen.State.ReceivedReady[m.Sender()]
// 	if found && receivedReady {
// 		log.Debugf("Already received ready for %s from %d on %d", m.roundID, m.Sender(), p.ID())
// 		return
// 	}

// 	// Get keygen store by serializing the data of message
// 	cid := m.Fingerprint()
// 	c := common.GetCStore(keygen, cid)

// 	keygen.ReadyStore = append(keygen.ReadyStore, m.share)

// 	// increment the echo messages received
// 	c.RC = c.RC + 1
// 	n, k, f := p.Params(m.newCommittee)

// 	log.Debugf("cid=%v,ready_count=%d, threshold=%d, node=%d", cid, c.RC, k, p.ID())

// 	if c.RC >= k && !c.ReadySent && c.EC >= k {
// 		// Broadcast ready message
// 		readyMsg := NewAcssReadyMessage(m.roundID, m.share, m.hash, m.curve, p.ID(), m.newCommittee)
// 		go p.Broadcast(m.newCommittee, readyMsg)
// 	}

// 	for i := 0; i < f; i += 1 {
// 		log.Debugf("len(readstore)=%d, threshold=%d", len(keygen.ReadyStore), (2*f + 1 + i))
// 		if len(keygen.ReadyStore) >= (2*f + 1 + i) {
// 			// Create RS encoding
// 			f, err := infectious.NewFEC(k, n)
// 			if err != nil {
// 				log.Debugf("error during creation of fec, err=%s", err)
// 				return
// 			}

// 			M, err := acss.Decode(f, keygen.ReadyStore)
// 			if err != nil {
// 				log.Debugf("Decode faced an error, err=%s", err)
// 				return
// 			}
// 			hash := common.Hash(M)
// 			log.Debugf("HashCompare, hash=%v, mHash=%v", hash, m.hash)

// 			if bytes.Equal(hash, m.hash) {
// 				outputMsg := NewAcssOutputMessage(m.roundID, M, m.curve, p.ID(), "ready", m.newCommittee)
// 				go p.ReceiveMessage(outputMsg)
// 				defer func() { keygen.State.Phase = common.Ended }()
// 				// send to other committee
// 				msg := messages.MessageData{}
// 				err := msg.Deserialize(M)
// 				if err != nil {
// 					log.Debugf("Could not deserialize message data, err=%s", err)
// 					return
// 				}
// 				for _, n := range p.Nodes(!m.newCommittee) {
// 					go func(node common.DkgParticipant) {
// 						readyMsg := NewAcssCommitMessage(m.roundID, msg.Commitments, m.curve, p.ID(), m.newCommittee)
// 						p.Send(readyMsg, node)
// 					}(n)
// 				}
// 			}

// 		}
// 	}

// }
