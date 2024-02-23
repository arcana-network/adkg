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
	RoundID          common.PSSRoundID  // ID of the round.
	Kind             string             // Type of the message.
	CurveName        common.CurveName   // Name of curve used in the messages.
	Secret           curves.Scalar      // Scalar that will be shared.
	EphemeralKeypair common.KeyPair     // the dealer's ephemeral keypair at the start of the protocol (Section V(C)hbACSS)
	Dealer           common.NodeDetails // Information of the node that starts the Dual-Committee ACSS.
}

// NewDualCommitteeACSSShareMessage creates a new share message from the provided ID and
// curve.
func NewDualCommitteeACSSShareMessage(secret curves.Scalar, dealer common.NodeDetails, roundID common.PSSRoundID, curve *curves.Curve, EphemeralKeypair common.KeyPair) (*common.PSSMessage, error) {
	m := &DualCommitteeACSSShareMessage{
		RoundID:          roundID,
		Kind:             ShareMessageType,
		CurveName:        common.CurveName(curve.Name),
		Secret:           secret,
		EphemeralKeypair: EphemeralKeypair,
		Dealer:           dealer,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (msg *DualCommitteeACSSShareMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	// Node can receive this msg only from themselves. Compare pubkeys to be sure.
	if !self.Details().IsEqual(sender) {
		return
	}

	curve := common.CurveFromName(msg.CurveName)

	// Generate secret
	secret := sharing.GenerateSecret(curve)

	// Emephemeral Private key of the dealer
	privateKey := msg.EphemeralKeypair.PrivateKey

	// Generate share and commitments
	n, k, _ := self.Params()

	// TODO do we need to check whether DPSS has already started?

	ExecuteACSS(true, secret, self, privateKey, curve, n, k, msg)
	ExecuteACSS(false, secret, self, privateKey, curve, n, k, msg)
}

// ExecuteACSS starts the execution of the ACSS protocol with a given committee
// defined by the withOldCommitte flag.
func ExecuteACSS(withOldCommittee bool, secret curves.Scalar, self common.PSSParticipant, privateKey curves.Scalar,
	curve *curves.Curve, n int, k int, msg *DualCommitteeACSSShareMessage) {
	// TODO implement this correctly

	//TODO: how do we get this node's ID (see line 93 nodePublicKey)
	// receivingNodes := self.Nodes(withOldCommittee)

	commitments, shares, err := acss.GenerateCommitmentAndShares(secret, uint32(k), uint32(n), curve)

	// Compress commitments
	compressedCommitments := acss.CompressCommitments(commitments)

	// Init share map
	shareMap := make(map[uint32][]byte, n)

	// encrypt each share with node respective generated symmetric key using Ephemeral Private key and add to share map
	for _, share := range shares {

		nodePublicKey := self.PublicKey(int(share.Id), !withOldCommittee)

		// NOTICE!!! this Encrypt does not use privateKey
		// TODO: we need to implement correct encrypt

		// This Encrypt will be a symmetric key encryption Ki = PKi ^ SKd
		cipherShare, _ := acss.Encrypt(share.Bytes(), nodePublicKey, privateKey)
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		shareMap[share.Id] = cipherShare
	}

	// Create message data
	msgData := messages.MessageData{
		Commitments: compressedCommitments,
		ShareMap:    shareMap,
	}

	// Create propose message & broadcast
	//NOTE: This proposeMsg should NOT have Emephemeral Private key of the dealer but only the public key.
	proposeMsg, err := NewDacssProposeMessage(msg.RoundID, msgData, common.CurveFromName(msg.CurveName), self.Details().Index, true, msg.EphemeralKeypair.PublicKey)

	if err != nil {
		log.Errorf("NewHbAcssPropose:err=%v", err)
		return
	}

	// Step 103
	// ReliableBroadcast(C)
	go self.Broadcast(true, *proposeMsg)
}
