package dacss

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
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
	RoundID            common.PSSRoundID      // ID of the round.
	Kind               string                 // Type of the message.
	CurveName          common.CurveName       // Name of curve used in the messages.
	Secret             curves.Scalar          // Scalar that will be shared.
	EphemeralKeypair   common.KeyPair         // the dealer's ephemeral keypair at the start of the protocol (Section V(C)hbACSS)
	Dealer             common.NodeDetails     // Information of the node that starts the Dual-Committee ACSS.
	NewCommitteeParams common.CommitteeParams // n, k & t parameters of the new committee
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

	// Only the nodes of the Old Committee should start Dual ACSS.
	if !self.IsOldNode() {
		log.Infof("DualCommitteeACSSShareMessage: Only Old nodes should start Dual ACSS. Not taking action.")
		return
	}

	// Node can receive this msg only from themselves. Compare pubkeys to be sure.
	if !self.Details().IsEqual(sender) {
		return
	}

	// Check that Dual ACSS has not started yet for this node.
	if self.State().DualAcssStarted {
		log.Infof("DualCommitteeACSSShareMessage: DualAcss already started. Not taking action.")
		return
	}

	self.State().DualAcssStarted = true

	curve := common.CurveFromName(msg.CurveName)

	// Ephemeral Private key of the dealer
	privateKey := msg.EphemeralKeypair.PrivateKey

	n_old, k_old, _ := self.Params()
	n_new := msg.NewCommitteeParams.N
	k_new := msg.NewCommitteeParams.K

	// This is to make sure both runs of ACCS are done
	var wg sync.WaitGroup
	wg.Add(2)

	// Initiate ACSS for both old and new Committe
	ExecuteACSS(false, msg.Secret, self, privateKey, curve, n_old, k_old, msg, &wg)
	ExecuteACSS(true, msg.Secret, self, privateKey, curve, n_new, k_new, msg, &wg)

	wg.Wait()
}

// ExecuteACSS starts the execution of the ACSS protocol with a given committee
// defined by the withNewCommittee flag.
func ExecuteACSS(withNewCommittee bool, secret curves.Scalar, self common.PSSParticipant, privateKey curves.Scalar,
	curve *curves.Curve, n int, k int, msg *DualCommitteeACSSShareMessage, wg *sync.WaitGroup) {

	commitments, shares, err := sharing.GenerateCommitmentAndShares(secret, uint32(k), uint32(n), curve)

	// Compress commitments
	compressedCommitments := sharing.CompressCommitments(commitments)

	// Init share map
	shareMap := make(map[uint32][]byte, n)

	// encrypt each share with node respective generated symmetric key using Ephemeral Private key and add to share map
	for _, share := range shares {

		nodePublicKey := self.PublicKey(int(share.Id), withNewCommittee)
		if nodePublicKey == nil {
			log.Errorf("Couldn't obtain public key for node with id=%v", share.Id)
			return
		}

		// This Encrypt will be a symmetric key encryption Ki = PKi ^ SKd
		// TODO does this need encoding?
		// TODO this encryption doesn't do MAC, is that needed
		cipherShare, err := sharing.EncryptSymmetricCalculateKey(share.Bytes(), nodePublicKey, privateKey)
		// TODO remove, this is just for testing purposes
		test, errT := sharing.DecryptSymmetricCalculateKey(cipherShare, nodePublicKey, privateKey)
		fmt.Println("test", test)
		if errT != nil {
			log.Errorf("Error while DECRYPT secret share, err=%v", err)
			return
		}

		if err != nil {
			log.Errorf("Error while encrypting secret share, err=%v", err)
			return
		}
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		shareMap[share.Id] = cipherShare
	}

	// Create message data
	msgData := messages.MessageData{
		Commitments: compressedCommitments,
		ShareMap:    shareMap,
	}

	// Create propose message & broadcast
	// NOTE: This proposeMsg should NOT have Emephemeral Private key of the dealer but only the public key.
	proposeMsg, err := NewAcssProposeMessageroundID(msg.RoundID, msgData, msg.CurveName, withNewCommittee, msg.EphemeralKeypair.PublicKey)

	if err != nil {
		log.Errorf("Error while creating new AcssProposeMessage, err=%v", err)
		return
	}

	// ReliableBroadcast(C)
	go func() {
		self.Broadcast(withNewCommittee, *proposeMsg)
		defer wg.Done()
	}()

}
