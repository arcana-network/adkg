package dacss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

// The message for the actual execution of the Implicate flow
var ImplicateExecuteMessageType string = "dacss_implicate_execute"

type ImplicateExecuteMessage struct {
	ACSSRoundDetails common.ACSSRoundDetails // ID of the specific ACSS round within DPSS.
	Kind             string                  // Type of the message
	CurveName        common.CurveName        // Name (indicator) of curve used in the messages.
	SymmetricKey     []byte                  // Compressed Affine Point
	Proof            []byte                  // Contains d, R, S
	SenderPubkeyHex  string                  // Hex of Compressed Affine Point
}

func NewImplicateExecuteMessage(acssRoundDetails common.ACSSRoundDetails, curveName common.CurveName, symmetricKey []byte, proof []byte, senderPubkeyHex string) (*common.PSSMessage, error) {
	m := &ImplicateExecuteMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ImplicateExecuteMessageType,
		CurveName:        curveName,
		SymmetricKey:     symmetricKey,
		Proof:            proof,
		SenderPubkeyHex:  senderPubkeyHex,
	}

	// Use bijson because of bigint in ACSSRoundDetails
	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, m.Kind, bytes)
	return &msg, nil
}

/*
The ImplicateExecuteHandler verifies whether the "implication" is correct and if so, goes into ShareRecovery.
For this step the Node needs to have the shareMap for the specific ACSS round.

Steps:
 1. Check self is the sender
    If not: return
 2. Retrieve the ACSS state for the specific ACSS round
    If not found: error and return
 3. Check whether we are already in Share Recovery for this ACSS round
    If so: return
 4. Verify ZKP (this should pass) and Predicate on the share (this should fail)
    If both are as expected, proceed to Share Recovery
    Otherwise: return
*/
func (msg *ImplicateExecuteMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	if !self.Details().IsEqual(sender) {
		return
	}

	self.State().AcssStore.Lock()
	defer self.State().AcssStore.Unlock()

	dacssState, found, err := self.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	if err != nil {
		// TODO error
	}

	// If for this specific ACSS round, we are already in Share Recovery, ignore msg
	if dacssState.ShareRecoveryOngoing {
		return
	}

	// At this point we should have the sharemap for this acss round
	if !found || dacssState.AcssData.IsUninitialized() {
		// TODO error
	}

	curve := curves.GetCurveByName(string(msg.CurveName))

	proof, err := sharing.UnpackProof(curve, msg.Proof)
	if err != nil {
		log.Errorf("Can't unpack nizk proof in Implicate flow for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
		return
	}

	g := curve.NewGeneratorPoint()

	_, k, _ := self.Params()

	// Verify ZKP that the symmetric key is actually generated correctly

	// Retrieve all data wrt the specific ACSS round from Node's storage
	// from this we also get the dealer's ephemeral public key
	AcssData := dacssState.AcssData
	share := AcssData.ShareMap[msg.SenderPubkeyHex]
	commitments := AcssData.Commitments

	// Convert hex of dealer's ephemeral key to a point
	dealerPubkey := AcssData.DealerEphemeralPubKey
	PK_d, err := common.HexToPoint(msg.CurveName, dealerPubkey)
	if err != nil {
		log.Errorf("Hex of dealer's ephemeral key couldn't be transformed to a Point, Implicate - ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
		return
	}

	K_i_d, err := curve.Point.FromAffineCompressed(msg.SymmetricKey)
	if err != nil {
		log.Errorf("Can't unpack symmetric key, Implicate - ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
		return
	}

	PK_i, err := common.HexToPoint(msg.CurveName, msg.SenderPubkeyHex)
	if err != nil {
		log.Errorf("Hex of implicate initiator pubkey couldn't be transformed to a Point, Implicate - ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
		return
	}

	// Verify ZKP
	symmKeyVerified := sharing.Verify(proof, g, PK_i, PK_d, K_i_d, curve)
	if !symmKeyVerified {
		log.Errorf("Verification of ZKP failed in Implicate flow, ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)

		return
	}

	// Check Predicate for share
	_, _, verified := sharing.Predicate(K_i_d, share, commitments, k, common.CurveFromName(msg.CurveName))

	// If the Predicate checks out, the Implicate flow was started for no reason
	if verified {
		log.Errorf("Predicate doesn't fail for share. Implicate invalid, ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)

		return
	}

	// If all checks have passed, the share for sender was indeed incorrect
	// we proceed onto share recovery

	// send ShareRecoveryMsg to self
	recoveryMsg, err := NewShareRecoveryMessage(msg.ACSSRoundDetails)
	self.ReceiveMessage(self.Details(), *recoveryMsg)
}