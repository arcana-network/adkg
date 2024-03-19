package dacss

import (
	"encoding/hex"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/torusresearch/bijson"

	log "github.com/sirupsen/logrus"
)

// ShareMessageType tells wich message are we sending. In this case, the share
// message.
var ShareMessageType string = "dacss_share"

// DualCommitteeACSSShareMessage has all the information for the initial message in the
// Dual-Committee ACSS Share protocol.
type DualCommitteeACSSShareMessage struct {
	ACSSRoundDetails   common.ACSSRoundDetails // All details off ACSS & DPSS round
	Kind               string                  // Type of the message.
	CurveName          common.CurveName        // Name of curve used in the messages.
	Secret             curves.Scalar           `json:"Secret"` // Scalar that will be shared.
	EphemeralSecretKey []byte                  // the dealer's ephemeral secret key at the start of the protocol (Section V(C)hbACSS)
	EphemeralPublicKey []byte                  // the dealer's ephemeral public key.
	Dealer             common.NodeDetails      // Information of the node that starts the Dual-Committee ACSS.
	NewCommitteeParams common.CommitteeParams  // n, k & t parameters of the new committee
}

// for marshalling and unmarshalling Secret(curve.Scalar)
func (m *DualCommitteeACSSShareMessage) UnmarshalJSON(data []byte) error {
	type Alias DualCommitteeACSSShareMessage
	aux := &struct {
		Secret bijson.RawMessage `json:"Secret"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := bijson.Unmarshal(data, aux); err != nil {
		return err
	}

	if len(aux.Secret) > 0 {
		epk, err := common.ScalarUnmarshalJson([]byte(aux.Secret))
		if err != nil {
			return err
		}
		m.Secret = epk
	}

	return nil
}

// convert DualCommitteeACSSShareMessage to json
func (m *DualCommitteeACSSShareMessage) MarshalJSON() ([]byte, error) {
	// Convert Secret to a suitable JSON representation
	scalarJSON, err := common.ScalarMarshalJson(m.Secret)
	if err != nil {
		return nil, err
	}

	// Marshal the rest of DualCommitteeACSSShareMessage as usual, but replace
	// Secret with its JSON representation
	type Alias DualCommitteeACSSShareMessage // Prevent recursion
	return bijson.Marshal(&struct {
		Secret bijson.RawMessage `json:"Secret"`
		*Alias
	}{
		Secret: bijson.RawMessage(scalarJSON),
		Alias:  (*Alias)(m),
	})
}

// NewDualCommitteeACSSShareMessage creates a new share message from the provided ID and
// curve.
func NewDualCommitteeACSSShareMessage(secret curves.Scalar, dealer common.NodeDetails, acssRoundDetails common.ACSSRoundDetails, curve *curves.Curve, ephemeralSecretKey []byte, ephemeralPublicKey []byte, newCommitteeParams common.CommitteeParams) (*common.PSSMessage, error) {
	m := &DualCommitteeACSSShareMessage{
		ACSSRoundDetails:   acssRoundDetails,
		Kind:               ShareMessageType,
		CurveName:          common.CurveName(curve.Name),
		Secret:             secret,
		EphemeralSecretKey: ephemeralSecretKey,
		EphemeralPublicKey: ephemeralPublicKey,
		Dealer:             dealer,
		NewCommitteeParams: newCommitteeParams,
	}
	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, m.Kind, bytes)
	return &msg, nil
}

func (msg DualCommitteeACSSShareMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	// Only the nodes of the Old Committee should start Dual ACSS.
	if self.IsNewNode() {
		log.Infof("DualCommitteeACSSShareMessage: Only Old nodes should start Dual ACSS. Not taking action.")
		return
	}

	// Node can receive this msg only from themselves. Compare pubkeys to be sure.
	if !self.Details().IsEqual(sender) {
		return
	}

	curve := common.CurveFromName(msg.CurveName)

	// Ephemeral Private key of the dealer
	privateKey, err := curve.Scalar.SetBytes(msg.EphemeralSecretKey)
	if err != nil {
		log.Errorf("DualCommitteeACSSShareMessage: error constructing the private key: %v", err)
		return
	}

	n_old, k_old, _ := self.Params()
	n_new := msg.NewCommitteeParams.N
	k_new := msg.NewCommitteeParams.K

	// Initiate ACSS for both old and new Committe
	ExecuteACSS(false, msg.Secret, self, privateKey, curve, n_old, k_old, msg, msg.EphemeralPublicKey)
	ExecuteACSS(true, msg.Secret, self, privateKey, curve, n_new, k_new, msg, msg.EphemeralPublicKey)
}

// ExecuteACSS starts the execution of the ACSS protocol with a given committee
// defined by the withNewCommittee flag.
func ExecuteACSS(withNewCommittee bool, secret curves.Scalar, sender common.PSSParticipant, privateKey curves.Scalar,
	curve *curves.Curve, n int, k int, msg DualCommitteeACSSShareMessage, dealerEphemeralPubkey []byte) {

	commitments, shares, err := sharing.GenerateCommitmentAndShares(secret, uint32(k), uint32(n), curve)
	if err != nil {
		log.Errorf("Error generating shares & commitments, err=%v", err)
		return
	}
	// Compress commitments
	compressedCommitments := sharing.CompressCommitments(commitments)

	// Init share map
	// a share gets stored for the pubkey of the receiving node
	shareMap := make(map[string][]byte, n)

	// encrypt each share with node respective generated symmetric key using Ephemeral Private key and add to share map
	for _, share := range shares {

		nodePublicKey := sender.GetPublicKeyFor(int(share.Id), withNewCommittee)
		if nodePublicKey == nil {
			log.Errorf("Couldn't obtain public key for node with id=%v", share.Id)
			return
		}

		// This Encrypt will be a symmetric key encryption Ki = PKi ^ SKd
		// TODO does this need encoding?
		// TODO this encryption doesn't do MAC, is that needed
		cipherShare, err := sharing.EncryptSymmetricCalculateKey(share.Bytes(), nodePublicKey, privateKey)
		if err != nil {
			log.Errorf("Error while encrypting secret share, err=%v", err)
			return
		}
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		pubkeyHex := common.PointToHex(nodePublicKey)
		shareMap[pubkeyHex] = cipherShare
	}

	// Create Dacss data for this round
	msgData := common.AcssData{
		Commitments:           compressedCommitments,
		ShareMap:              shareMap,
		DealerEphemeralPubKey: hex.EncodeToString(dealerEphemeralPubkey),
	}

	// Create propose message & broadcast
	// NOTE: This proposeMsg should NOT have Emephemeral Private key of the dealer but only the public key.
	proposeMsg, err := NewAcssProposeMessageroundID(msg.ACSSRoundDetails, msgData, msg.CurveName, withNewCommittee, msg.NewCommitteeParams)

	if err != nil {
		log.Errorf("Error while creating new AcssProposeMessage, err=%v", err)
		return
	}

	// ReliableBroadcast(C)
	go sender.Broadcast(withNewCommittee, *proposeMsg)
}
