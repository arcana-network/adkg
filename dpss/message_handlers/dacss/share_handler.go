package dacss

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

// ShareMessageType tells wich message are we sending. In this case, the share
// message.
var ShareMessageType string = "dacss_share"

// DacssShareMessage has all the information for the initial message in the
// sharing phase.
type DacssShareMessage struct {
	RoundID common.RoundID    // ID of the round.
	Kind    string 						// Type of the message.
	Curve   *curves.Curve     // Curve used in the messages.
}

// NewDacssShareMessage creates a new share message from the provided ID and
// curve.
func NewDacssShareMessage(roundID common.RoundID, curve *curves.Curve) (*common.DKGMessage, error) {
	m := &DacssShareMessage{
		RoundID: roundID,
		Kind:    ShareMessageType,
		Curve:   curve,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (msg *DacssShareMessage) Process(sender common.KeygenNodeDetails, self common.PSSParticipant) {
	if sender.Index != self.ID() {
		return
	}

	// Generate the secret
	secret := acss.GenerateSecret(msg.Curve)

	// Generate the private key
	privKey := self.PrivateKey()

	makeMessageAndSend(false, self, msg, secret, privKey)
	makeMessageAndSend(true, self, msg, secret, privKey)
}

func makeMessageAndSend(isNewCommittee bool, self common.PSSParticipant, msg *DacssShareMessage, secret curves.Scalar, privateKey curves.Scalar) {
	n, k, _ := self.Params(isNewCommittee)

	// Generates shares and commitments
	commitments, shares, _ := acss.GenerateCommitmentAndShares(secret, uint32(k), uint32(n), msg.Curve)
	// Compress commitments
	compressedCommitments := acss.CompressCommitments(commitments)

	// Init share map
	shareMap := make(map[uint32][]byte, n)

	// encrypt each share with node respective generated symmetric key, add to share map
	for _, share := range shares {
		nodePublicKey := self.PublicKey(int(share.Id), isNewCommittee)
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
	for _, n := range self.Nodes(isNewCommittee) {
		go func(node common.KeygenNodeDetails) {
			roundID := reCreateRoundID(msg.RoundID, isNewCommittee)
			proposeMsg, err := NewDacssProposeMessage(roundID, msgData, msg.Curve, self.ID(), isNewCommittee)
			if err != nil {
				log.WithField("error", err).Error("NewDacssProposeMessage")
				return
			}
			self.Send(node, *proposeMsg)
		}(n)
	}
}

// Re-creating roundID base on committeeType
func reCreateRoundID(id common.RoundID, newCommittee bool) common.RoundID {

	var committeeType int
	if newCommittee {
		committeeType = 1
	} else {
		committeeType = 0
	}
	committeeID := *new(big.Int).SetInt64(int64(committeeType))
	r := &common.RoundDetails{}
	s := string(id)
	substrings := strings.Split(s, common.Delimiter4)
	if len(substrings) != 3 {
		log.Error("expected length of 3, ", len(substrings))
	}
	r.ADKGID = common.ADKGID(strings.Join([]string{substrings[0], committeeID.Text(16)}, common.Delimiter3))
	r.Kind = substrings[1]
	index, err := strconv.Atoi(substrings[2])
	if err != nil {
		log.Error("AtoiError", err)
	}
	r.Dealer = index

	return r.ID()
}
