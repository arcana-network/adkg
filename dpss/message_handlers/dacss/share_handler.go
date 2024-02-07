package dacss

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/adkg-proto/common"
	"github.com/arcana-network/adkg-proto/common/acss"
	"github.com/arcana-network/adkg-proto/messages"
)

var AcssShareMessageType common.MessageType = "dacss_share"

type DacssShareMessage struct {
	roundID common.RoundID
	sender  int
	kind    common.MessageType
	curve   *curves.Curve
}

// NewAcssShareMessage creates a new share message with the provided arguments.
func NewAcssShareMessage(roundID common.RoundID, curve *curves.Curve, sender int) common.DKGMessage {
	m := AcssShareMessage{
		roundID,
		sender,
		AcssShareMessageType,
		curve,
	}
	return &m
}

func (m *AcssShareMessage) Sender() int {
	return m.sender
}

func (m *AcssShareMessage) Kind() common.MessageType {
	return m.kind
}

func (m *AcssShareMessage) Process(p common.DkgParticipant) {
	if m.Sender() != p.ID() {
		return
	}
	// Generate secret
	secret := acss.GenerateSecret(m.curve)

	// Generate Emphemeral keypair
	privateKey := p.SelfPrivateKey()

	makeMessageAndSend(false, p, m, secret, privateKey)
	makeMessageAndSend(true, p, m, secret, privateKey)
}

func makeMessageAndSend(newCommittee bool, p common.DkgParticipant, m *AcssShareMessage, secret curves.Scalar, privateKey curves.Scalar) {

	n, k, _ := p.Params(newCommittee)
	// Generate share and commitments
	commitments, shares, _ := acss.GenerateCommitmentAndShares(secret,
		uint32(k), uint32(n), m.curve)

	// Compress commitments
	compressedCommitments := acss.CompressCommitments(commitments)

	// Init share map
	shareMap := make(map[uint32][]byte, n)

	// encrypt each share with node respective generated symmetric key, add to share map
	for _, share := range shares {
		nodePublicKey := p.PublicKey(int(share.Id))
		cipherShare, _ := acss.Encrypt(share.Bytes(), nodePublicKey, privateKey)
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		shareMap[share.Id] = cipherShare
	}

	// Create message data
	messageData := messages.MessageData{
		Commitments: compressedCommitments,
		ShareMap:    shareMap,
	}
	// Create propose message & broadcast
	for _, n := range p.Nodes(newCommittee) {
		go func(node common.DkgParticipant) {
			roundid := reCreateRoundID(m.roundID, newCommittee)
			proposeMsg := NewAcssProposeMessage(roundid, messageData, m.curve, p.ID(), newCommittee)

			p.Send(proposeMsg, node)
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
		log.Error("%s", err)
	}
	r.Dealer = index

	return r.ID()
}
