package dacss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

// ShareMessageType tells wich message are we sending. In this case, the share
// message.
var ShareMessageType common.MessageType = "dacss_share"

// DacssShareMessage has all the information for the initial message in the
// sharing phase.
type DacssShareMessage struct {
	roundID common.RoundID     // ID of the round.
	kind    common.MessageType // Type of the message.
	curve   *curves.Curve      // Curve used in the messages.
}

// NewDacssShareMessage creates a new share message from the provided ID and
// curve.
func NewDacssShareMessage(roundID common.RoundID, curve *curves.Curve) common.PSSMessage {
	message := &DacssShareMessage{
		roundID: roundID,
		kind:    ShareMessageType,
		curve:   curve,
	}
	return message
}

func (msg *DacssShareMessage) Process(sender common.KeygenNodeDetails, self common.PSSParticipant) {
	if sender.Index != self.ID() {
		return
	}

	// Generate the secret
	secret := acss.GenerateSecret(msg.curve)

	// Generate the private key
	privKey := self.PrivateKey()

	makeMessageAndSend(false, self, msg, secret, privKey)
	makeMessageAndSend(true, self, msg, secret, privKey)
}

func (msg *DacssShareMessage) Kind() common.MessageType {
	return msg.kind
}

func makeMessageAndSend(isNewCommittee bool, self common.PSSParticipant, msg *DacssShareMessage, secret curves.Scalar, privateKey curves.Scalar) {
	n, k, _ := self.Params(isNewCommittee)

	// Generates shares and commitments
	commitments, shares, _ := acss.GenerateCommitmentAndShares(secret, uint32(k), uint32(n), msg.curve)
	// Compress commitments
	compressedCommitments := acss.CompressCommitments(commitments)

	// Init share map
	shareMap := make(map[uint32][]byte, n)

	// encrypt each share with node respective generated symmetric key, add to share map
	for _, share := range shares {
		nodePublicKey := self.PublicKey(int(share.Id))
		cipherShare, _ := acss.Encrypt(share.Bytes(), nodePublicKey, privateKey)
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		shareMap[share.Id] = cipherShare
	}

	// Create message data
	msgData := messages.MessageData{
		Commitments: compressedCommitments,
		ShareMap:    shareMap,
	}

	// Create propose message & broadcast.
	// Question: how to returns the nodes according to the new and old committees?
	// TODO: Correct errors here.
	for _, n := range self.Nodes(isNewCommittee) {
		go func(node common.PSSParticipant) {
			roundID := reCreateRoundID(msg.roundID, isNewCommittee)
			proposeMsg := NewDacssProposeMessage(roundID, msgData, msg.curve, self.ID(), isNewCommittee)

			self.Send(proposeMsg, node)
		}(n)
	}
}
