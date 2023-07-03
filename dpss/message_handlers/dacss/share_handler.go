package dacss

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var AcssShareMessageType common.DPSSMessageType = "dacss_share"

type AcssShareMessage struct {
	roundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
}

func NewAcssShareMessage(roundID common.DPSSRoundID, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := AcssShareMessage{
		roundID: roundID,
		kind:    AcssShareMessageType,
		curve:   curve,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.roundID, m.kind, bytes)
	return &msg, nil
}

func (m *AcssShareMessage) Process(sender common.KeygenNodeDetails, self dpsscommon.DPSSParticipant) {
	if sender.Index != self.ID() {
		return
	}
	// Generate secret
	secret := acss.GenerateSecret(m.curve)

	// Generate Emphemeral keypair
	privateKey := self.SelfPrivateKey()

	makeMessageAndSend(false, self, m, secret, privateKey)
	makeMessageAndSend(true, self, m, secret, privateKey)

}

func makeMessageAndSend(newCommittee bool, self dpsscommon.DPSSParticipant, m *AcssShareMessage, secret curves.Scalar, privateKey curves.Scalar) {

	n, k, _ := self.Params(newCommittee)
	// Generate share and commitments
	commitments, shares, _ := acss.GenerateCommitmentAndShares(secret,
		uint32(k), uint32(n), m.curve)

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
	messageData := messages.MessageData{
		Commitments: compressedCommitments,
		ShareMap:    shareMap,
	}

	// Create propose message & broadcast
	roundid := reCreateRoundID(m.roundID, newCommittee)
	msg, err := NewAcssProposeMessage(roundid, messageData, m.curve, newCommittee)
	if err != nil {
		return
	}

	go self.Broadcast(newCommittee, *msg)
}

// Re-creating roundID base on committeeType
func reCreateRoundID(id common.DPSSRoundID, newCommittee bool) common.DPSSRoundID {

	var committeeType int
	if newCommittee {
		committeeType = 1
	} else {
		committeeType = 0
	}
	committeeID := *new(big.Int).SetInt64(int64(committeeType))
	r := &common.DPSSRoundDetails{}
	s := string(id)
	substrings := strings.Split(s, common.Delimiter4)
	if len(substrings) != 3 {
		log.Error("expected length of 3, ", len(substrings))
	}
	r.DPSSID = common.DPSSID(strings.Join([]string{substrings[0], committeeID.Text(16)}, common.Delimiter3))
	r.Kind = substrings[1]
	index, err := strconv.Atoi(substrings[2])
	if err != nil {
		log.Error("%s", err)
	}
	r.Dealer = index

	return r.ID()
}
