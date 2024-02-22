package dacss

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

// ShareMessageType tells wich message are we sending. In this case, the share
// message.
var ShareMessageType string = "DualCommitteeACSS_share"

// DualCommitteeACSSShareMessage has all the information for the initial message in the
// Dual-Committee ACSS Share protocol.
type DualCommitteeACSSShareMessage struct {
	RoundID common.PSSRoundID  // ID of the round.
	Kind    string             // Type of the message.
	Curve   *curves.Curve      // Curve used in the messages.
	Secret  curves.Scalar      // Scallar that will be shared.
	Dealer  common.NodeDetails // Information of the node that starts the Dual-Committee ACSS.
}

// NewDualCommitteeACSSShareMessage creates a new share message from the provided ID and
// curve.
func NewDualCommitteeACSSShareMessage(secret curves.Scalar, dealer common.NodeDetails, roundID common.PSSRoundID, curve *curves.Curve) (*common.DKGMessage, error) {
	m := &DualCommitteeACSSShareMessage{
		RoundID: roundID,
		Kind:    ShareMessageType,
		Curve:   curve,
		Secret:  secret,
		Dealer:  dealer,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (msg *DualCommitteeACSSShareMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	// TODO do we need to check whether sender (NodeDetails) are contained in the PssNodeDetails?
	// Or do we need to specifically check that the NodeDetails are equal for new/old comittee (and somehow have access to this expectation)
	if sender.Index != self.Details().Index {
		return
	}

	// Secret Share Encoding
	// Step 101 of hbACSS
	// Sample B random degree-t polynomials φ1(·)...φB(·) such that each φk(0) = sk and φk(i) is Pi’s share of sk

	// Step 102 of hbACSS
	// Polynomial Commitment
	// C←{PolyCommit(SP,φk(·))}k∈[B]

	// Generate B number of secrets
	var secrets []curves.Scalar
	for i := 0; i < msg.BatchSize; i++ {
		secrets = append(secrets, acss.GenerateSecret(msg.Curve))
	}

	// Generates shares and commitments
	var BatchCommitments []*sharing.FeldmanVerifier
	var Batchshares [][]sharing.ShamirShare
	var BatchcompressedCommitments []byte

	for i := 0; i < msg.BatchSize; i++ {
		commitments, shares, _ := acss.GenerateCommitmentAndShares(secrets[i], uint32(k), uint32(n), msg.Curve)
		BatchCommitments = append(BatchCommitments, commitments)

		//These shares needs to be "Encrypt and Disperse" for each node
		Batchshares = append(Batchshares, shares)

		// Compress commitments
		compressedCommitments := acss.CompressCommitments(commitments)
		BatchcompressedCommitments = append(BatchcompressedCommitments, compressedCommitments...)
	}

	// Create message data
	// Here the ShareMap is an empty map since we need Disperse & Retrive method for sending & receiving
	// and we are only broadcasting the commitments

	msgData := messages.MessageData{
		Commitments: BatchcompressedCommitments,
		ShareMap:    map[uint32][]byte{},
	}

	// Create propose message & broadcast
	proposeMsg, err := NewHbAacssProposeMessage(msg.RoundID, msgData, msg.Curve, self.Details().Index, true)

	if err != nil {
		log.Errorf("NewHbAcssPropose:err=%v", err)
		return
	}

	// Step 103
	// ReliableBroadcast(C)
	go self.Broadcast(true, *proposeMsg)

}

// func makeMessageAndSend(isNewCommittee bool, self common.PSSParticipant, msg *HbAcssShareMessage, secret curves.Scalar, privateKey curves.Scalar) {
// 	n, k, _ := self.Params()

// 	// Generates shares and commitments
// 	commitments, shares, _ := acss.GenerateCommitmentAndShares(secret, uint32(k), uint32(n), msg.Curve)
// 	// Compress commitments
// 	compressedCommitments := acss.CompressCommitments(commitments)

// 	// Init share map
// 	shareMap := make(map[uint32][]byte, n)

// 	// encrypt each share with node respective generated symmetric key, add to share map
// 	for _, share := range shares {
// 		nodePublicKey := self.PublicKey(int(share.Id), isNewCommittee)
// 		cipherShare, _ := acss.Encrypt(share.Bytes(), nodePublicKey, privateKey)
// 		log.Debugf("CIPHER_SHARE=%v", cipherShare)
// 		shareMap[share.Id] = cipherShare
// 	}

// 	// Create message data
// 	msgData := messages.MessageData{
// 		Commitments: compressedCommitments,
// 		ShareMap:    shareMap,
// 	}

// 	// Create propose message & broadcast.
// 	// Question: how to returns the nodes according to the new and old committees?
// 	for _, n := range self.Nodes(isNewCommittee) {
// 		go func(node common.NodeDetails) {
// 			roundID := reCreateRoundID(msg.RoundID, isNewCommittee)

// 			proposeMsg, err := NewDacssProposeMessage(roundID, msgData, msg.Curve, self.Details().Index, isNewCommittee)

// 			if err != nil {
// 				log.WithField("error", err).Error("NewAcssProposeMessage")
// 				return
// 			}
// 			// TODO: as of now ABA is combined with dacss.
// 			// later can be separated

// 			// rbcRouter := router.NewRbcRouter("dacss")
// 			// proposeMsg, err := rbcRouter.StartRbc(roundID, msgData, msg.Curve, self.ID())

// 			// if err != nil {
// 			// 	log.WithField("error", err).Error("NewDacssProposeMessage")
// 			// 	return
// 			// }

// 			self.Send(node, *proposeMsg)
// 		}(n)
// 	}
// }

// // Re-creating roundID base on committeeType
// func reCreateRoundID(id common.RoundID, newCommittee bool) common.RoundID {

// 	var committeeType int
// 	if newCommittee {
// 		committeeType = 1
// 	} else {
// 		committeeType = 0
// 	}
// 	committeeID := *new(big.Int).SetInt64(int64(committeeType))
// 	r := &common.RoundDetails{}
// 	s := string(id)
// 	substrings := strings.Split(s, common.Delimiter4)
// 	if len(substrings) != 3 {
// 		log.Error("expected length of 3, ", len(substrings))
// 	}
// 	r.ADKGID = common.ADKGID(strings.Join([]string{substrings[0], committeeID.Text(16)}, common.Delimiter3))
// 	r.Kind = substrings[1]
// 	index, err := strconv.Atoi(substrings[2])
// 	if err != nil {
// 		log.Error("AtoiError", err)
// 	}
// 	r.Dealer = index

// 	return r.ID()
// }
