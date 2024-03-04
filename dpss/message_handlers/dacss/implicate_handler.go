package dacss

import (
	"encoding/hex"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

// TODO docs
// TODO tests
// TODO error handling

var ImplicateMessageType string = "dacss_implicate"

type ImplicateMessage struct {
	ACSSRoundDetails common.ACSSRoundDetails // ID of the specific ACSS round within DPSS.
	Kind             string
	CurveName        common.CurveName
	SymmetricKey     []byte // Compressed Affine Point
	Proof            []byte // Contains d, R, S
}

func NewImplicateMessage(acssRoundDetails common.ACSSRoundDetails, curve *curves.Curve, symmetricKey []byte, proof []byte) (*common.PSSMessage, error) {
	m := &ImplicateMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ImplicateMessageType,
		CurveName:        common.CurveName(curve.Name),
		SymmetricKey:     symmetricKey,
		Proof:            proof,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, m.Kind, bytes)
	return &msg, nil
}

func (msg *ImplicateMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	dacssState, found, err := self.State().DacssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	if err != nil {
		// TODO error
	}

	if !found {
		// TODO error
	}

	// If for this specific ACSS round, we are already in Share Recovery, ignore msg
	if dacssState.ShareRecoveryOngoing {
		return
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
	PK_i, _ := curve.Point.Set(&sender.PubKey.X, &sender.PubKey.Y) // TODO err

	// FIXME we have to consider the possibility the shareMap has not been received yet.
	// possible solution: upon receiving shareMap, if an implicate message was received,
	// trigger the following functionality
	// For now, we just execute directly

	// Retrieve all data wrt the specific ACSS round from Node's storage
	// from this we also get the dealer's ephemeral public key
	dacssData := dacssState.DacssData
	senderPubkeyHex := hex.EncodeToString(PK_i.ToAffineCompressed())
	share := dacssData.ShareMap[senderPubkeyHex]
	commitments := dacssData.Commitments

	// FIXME convert hex string dealerPubkey to PK_d: Point
	// dealerPubkey := dacssData.DealerEphemeralPubKey
	PK_d, _ := curve.Point.Set(&sender.PubKey.X, &sender.PubKey.Y) // TODO err
	// TODO error handling
	K_i_d, _ := curve.Point.FromAffineCompressed(msg.SymmetricKey)
	symmKeyVerified := sharing.Verify(proof, g, PK_i, PK_d, K_i_d, curve)
	if !symmKeyVerified {
		return
	}

	// Check Predicate for share
	_, _, verified := sharing.Predicate(K_i_d, share, commitments, k, common.CurveFromName(msg.CurveName))

	if verified {
		return
	}

	// send ShareRecoveryMsg
	recoveryMsg, err := NewShareRecoveryMessage(msg.ACSSRoundDetails)

	// TODO boradcast?
	self.Broadcast(!self.IsOldNode(), *recoveryMsg)

}
