package dacss

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/keygen/messages"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

// TODO docs
// TODO tests
// TODO error handling

var ImplicateMessageType string = "dacss_implicate"

type ImplicateMessage struct {
	RoundID      common.PSSRoundID
	AcssRoundID  common.ACSSRoundID
	Kind         string
	CurveName    common.CurveName
	SymmetricKey []byte // Compressed Affine Point
	Proof        []byte // Contains d, R, S
	Data         messages.MessageData
}

func NewImplicateMessage(roundID common.PSSRoundID, acssID common.ACSSRoundID, curve *curves.Curve, symmetricKey []byte, proof []byte, messageData messages.MessageData) (*common.PSSMessage, error) {
	m := &ImplicateMessage{
		RoundID:      roundID,
		AcssRoundID:  acssID,
		Kind:         ImplicateMessageType,
		CurveName:    common.CurveName(curve.Name),
		SymmetricKey: symmetricKey,
		Proof:        proof,
		Data:         messageData,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (msg *ImplicateMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	// If for this specific ACSS round, we are already in Share Recovery, ignore msg
	if self.State().ShareRecoveryOngoing[msg.AcssRoundID] {
		return
	}

	curve := curves.GetCurveByName(string(msg.CurveName))

	proof, err := sharing.UnpackProof(curve, msg.Proof)
	if err != nil {
		log.Errorf("Can't unpack nizk proof in Implicate flow for ACSS round %s, err: %s", msg.AcssRoundID, err)
		return
	}

	g := curve.NewGeneratorPoint()

	_, k, _ := self.Params()

	// Verify ZKP that the symmetric key is actually generated correctly
	PK_i, _ := curve.Point.Set(&sender.PubKey.X, &sender.PubKey.Y) // TODO err
	// FIXME get the dealer's ephemeral pub key
	PK_d, _ := curve.Point.Set(&sender.PubKey.X, &sender.PubKey.Y) // TODO err
	// TODO error handling
	K_i_d, _ := curve.Point.FromAffineCompressed(msg.SymmetricKey)
	symmKeyVerified := sharing.Verify(proof, g, PK_i, PK_d, K_i_d, curve)
	if !symmKeyVerified {
		return
	}

	// Check Predicate for share
	// FIXME shareMap has to be verified, we cannot just accept it from the msg
	// FIXME how do we get the index of the share. And should it be part of the zkp?
	var index uint32 = 1
	_, _, verified := sharing.Predicate(K_i_d, msg.Data.ShareMap[index][:],
		msg.Data.Commitments[:], k, common.CurveFromName(msg.CurveName))

	if verified {
		return
	}

	// send ShareRecoveryMsg
	recoveryMsg, err := NewShareRecoveryMessage(msg.RoundID, msg.AcssRoundID)

	// TODO boradcast?
	self.Broadcast(!self.IsOldNode(), *recoveryMsg)

}
