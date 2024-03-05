package dacss

import (
	"reflect"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
	"github.com/vivint/infectious"
)

var DacssEchoMessageType string = "dacss_echo"

// DacssEchoMessage represents the echo handler in the RBC protocol
type DacssEchoMessage struct {
	ACSSRoundDetails common.ACSSRoundDetails // Details of the current row
	Kind             common.MessageType      // Type of the message
	Curve            *curves.Curve           // Curve used for the computation
	Share            infectious.Share        // Shard comming from the RS Encoding.
	Hash             []byte                  // Hash of the shares.
	NewCommittee     bool                    // Tells if this message was sent to an old or new committee.
}

// NewDacssEchoMessage creates an ECHO message in the RBC protocol.
func NewDacssEchoMessage(acssRoundDetails common.ACSSRoundDetails, share infectious.Share, hash []byte, curve *curves.Curve, sender int, newCommittee bool) (*common.PSSMessage, error) {
	m := DacssEchoMessage{
		ACSSRoundDetails: acssRoundDetails,
		NewCommittee:     newCommittee,
		Kind:             DacssEchoMessageType,
		Curve:            curve,
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

// Process handles the incomming ECHO message.
func (m *DacssEchoMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.Debugf("Echo received: Sender=%d, Receiver=%d", sender.Index, self.Details().Index)

	self.State().AcssStore.Lock()
	defer self.State().AcssStore.Unlock()

	acssState, isStored, err := self.State().AcssStore.Get(m.ACSSRoundDetails.ToACSSRoundID())
	if !isStored {
		log.WithField("error", "ACSS state not stored yet").Error("DacssEchoMessage - Process()")
		return
	}
	if err != nil {
		log.WithField("error", err).Error("DacssEchoMessage - Process()")
		return
	}
	ownShare := acssState.RBCState.OwnReedSolomonShard
	ownHash := acssState.RBCState.HashMsg

	// Check that the incoming message matches with the share of self (Line 11)
	// of Algorithm 4, "Asynchronous Data Disemination".
	if reflect.DeepEqual(ownShare.Data, m.Share.Data) && reflect.DeepEqual(m.Hash, ownHash) {
		acssState.RBCState.ReceivedEcho[sender.Index] = true
		_, t, _ := self.Params()
		if acssState.RBCState.CountEcho() >= 2*t+1 && !acssState.RBCState.IsReadyMsgSent {
			acssState.RBCState.IsReadyMsgSent = true
			readyMsg, err := NewDacssReadyMessage(m.ACSSRoundDetails, ownShare, m.Hash, m.Curve)
			if err != nil {
				log.WithField("error", err).Error("DacssEchoMessage - Process()")
				return
			}
			go self.Broadcast(m.NewCommittee, *readyMsg)
		}
	}
}
