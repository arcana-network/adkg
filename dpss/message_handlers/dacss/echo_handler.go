package dacss

import (
	"reflect"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
	"github.com/vivint/infectious"
)

var DacssEchoMessageType common.MessageType = "dacss_echo"

type DacssEchoMessage struct {
	ACSSRoundDetails common.ACSSRoundDetails
	CommitteeType    int
	Kind             common.MessageType
	Curve            *curves.Curve
	Share            infectious.Share
	Hash             []byte // Hash of the shares.
	NewCommittee     bool
}

func NewDacssEchoMessage(acssRoundDetails common.ACSSRoundDetails, s infectious.Share, hash []byte, curve *curves.Curve, sender int, newCommittee bool) (*common.PSSMessage, error) {
	m := DacssEchoMessage{
		ACSSRoundDetails: acssRoundDetails,
		NewCommittee:     newCommittee,
		Kind:             DacssEchoMessageType,
		Curve:            curve,
		Share:            s,
		Hash:             hash,
	}
	if newCommittee {
		m.CommitteeType = 1
	} else {
		m.CommitteeType = 0
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, string(m.Kind), bytes)
	return &msg, nil
}

func (m *DacssEchoMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.Debugf("Echo received: Sender=%d, Receiver=%d", sender.Index, self.Details().Index)

	self.State().AcssStore.Lock()
	defer self.State().AcssStore.Unlock()

	acssState, isStored, err := self.State().AcssStore.Get(m.ACSSRoundDetails.ToACSSRoundID())
	if !isStored {
		// TODO: Handle if the acssState is not stored yet.
	}
	if err != nil {
		// TODO: Handle error.
	}
	ownShare := acssState.RBCState.OwnReedSolomonShard
	ownHash := acssState.RBCState.HashMsg

	// Check that the incoming message matches with the share of self (Line 11)
	// of Algorithm 4, "Asynchronous Data Disemination".
	// TODO: Question - Do we need to compare the hash too?
	// TODO: We need to store the own share and the hash of the message in RBCState.
	if reflect.DeepEqual(ownShare.Data, m.Share.Data) && reflect.DeepEqual(m.Hash, ownHash) {
		acssState.RBCState.ReceivedEcho[sender.Index] = true
		_, t, _ := self.Params()
		if acssState.RBCState.CountEcho() >= 2*t+1 && !acssState.RBCState.IsReadyMsgSent {
			acssState.RBCState.IsReadyMsgSent = true
			readyMsg, err := NewDacssReadyMessage(m.ACSSRoundDetails, ownShare, m.Hash, m.Curve)
			if err != nil {
				// TODO: Handle error
			}
			self.Broadcast(m.NewCommittee, *readyMsg)
		}
	}
}

// func (m *DacssEchoMessage) Fingerprint() string {
// 	var bytes []byte
// 	delimiter := common.Delimiter2
// 	bytes = append(bytes, m.Hash...)
// 	bytes = append(bytes, delimiter...)

// 	bytes = append(bytes, m.Share.Data...)
// 	bytes = append(bytes, delimiter...)

// 	bytes = append(bytes, byte(m.Share.Number))
// 	bytes = append(bytes, delimiter...)
// 	hash := hex.EncodeToString(common.Keccak256(bytes))
// 	return hash
// }

// func (msg *DacssEchoMessage) Process(sender common.NodeDetails, self common.DkgParticipant) {
// 	log.Debugf("Echo received: Sender=%d, Receiver=%d", sender.Index, self.ID())
// 	// Get state from node
// 	state := self.State().KeygenStore

// 	// Create empty keygen state
// 	//TODO: needs to confirm
// 	defaultKeygen := &common.SharingStore{
// 		RoundID: msg.RoundID,
// 		State: common.RBCState{
// 			Phase:         common.Initial,
// 			ReceivedReady: make(map[int]bool),
// 			ReceivedEcho:  make(map[int]bool),
// 		},
// 		CStore: make(map[string]*common.CStore),
// 	}

// 	// Get or set if it doesn't exist
// 	keygen, complete := state.GetOrSetIfNotComplete(msg.RoundID, defaultKeygen)
// 	// log.Debugf("Keygen=%v, complete=%v", keygen, complete)
// 	if complete {
// 		// if keygen is complete, ignore and return
// 		return
// 	}

// 	keygen.Lock()
// 	defer keygen.Unlock()
// 	// Make sure the echo received from a node is set to true
// 	defer func() { keygen.State.ReceivedEcho[sender.Index] = true }()

// 	// Check if the echo has alreay been received.
// 	receivedEcho, found := keygen.State.ReceivedEcho[sender.Index]
// 	if receivedEcho && found {
// 		log.Debugf("Already received echo for %s from %d", msg.RoundID, sender.Index)
// 		return

// 		// Get keygen store by serializing the share and hash of the message.
// 		cid := msg.Fingerprint()
// 		c := common.GetCStore(keygen, cid)

// 		// increment the echo messages received
// 		c.EC = c.EC + 1

// 		// Broadcast ready message if echo count > 2f + 1
// 		_, _, f := self.Params()

// 		log.Debugf("echo_count=%d, required=%d", c.EC, 2*f+1)
// 		if c.EC >= (2*f + 1) {
// 			// Send Ready Message
// 			c.ReadySent = true
// 			for _, n := range self.Nodes() {
// 				go func(node common.NodeDetails) {
// 					// This corresponds to Line 12, Algorithm 4, RBC Protocol.
// 					readyMsg, err := NewDacssReadyMessage(msg.RoundID, msg.Share, msg.Hash, msg.Curve, self.ID(), msg.NewCommittee)
// 					if err != nil {
// 						log.WithField("error", err).Error("NewDacssReadyMessage")
// 						return
// 					}
// 					self.Send(node, *readyMsg)
// 				}(n)
// 			}
// 		}
// 	}
// }
