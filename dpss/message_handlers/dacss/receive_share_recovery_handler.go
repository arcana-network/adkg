package dacss

import (
	"reflect"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var ReceiveShareRecoveryMessageType string = "dacss_receive_share_recovery"

type ReceiveShareRecoveryMessage struct {
	ACSSRoundDetails common.ACSSRoundDetails
	Kind             string
	CurveName        common.CurveName // Name (indicator) of curve used in the messages.
	SymmetricKey     []byte           // Compressed Affine Point
	Proof            []byte           // Contains d, R, S
	AcssData         common.AcssData
}

func NewReceiveShareRecoveryMessage(acssRoundDetails common.ACSSRoundDetails, curveName common.CurveName, symmetricKey []byte, proof []byte, acssData common.AcssData) (*common.PSSMessage, error) {
	m := &ReceiveShareRecoveryMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ReceiveShareRecoveryMessageType,
		CurveName:        curveName,
		SymmetricKey:     symmetricKey,
		Proof:            proof,
		AcssData:         acssData,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, m.Kind, bytes)
	return &msg, nil
}

func (msg *ReceiveShareRecoveryMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	// Ignore if message comes from self
	if self.Details().IsEqual(sender) {
		return
	}

	self.State().AcssStore.Lock()
	defer self.State().AcssStore.Unlock()

	acssState, found, err := self.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	if err != nil || !found || len(acssState.AcssDataHash) == 0 {
		// TODO error
	}

	// Hash the received acssData
	hash, err := common.HashAcssData(msg.AcssData)
	if err != nil {
		// TODO error
	}

	if reflect.DeepEqual(acssState.AcssDataHash, hash) {
		// TODO error
		return
	}

	// TODO Ignore if Node already has recovered a share. This connects to how we conclude the ACSS round TBD
	curve := curves.GetCurveByName(string(msg.CurveName))

	proof, err := sharing.UnpackProof(curve, msg.Proof)
	if err != nil {
		// TODO error
	}

	g := curve.NewGeneratorPoint()
	_, k, t := self.Params()

	PK_j, err := common.PointToCurvePoint(self.Details().PubKey, msg.CurveName)
	// Convert hex of dealer's ephemeral key to a point
	dealerPubkey := msg.AcssData.DealerEphemeralPubKey
	PK_d, err := common.HexToPoint(msg.CurveName, dealerPubkey)

	K_j_d, err := curve.Point.FromAffineCompressed(msg.SymmetricKey)
	if err != nil {
		// TODO error
	}

	// Verify ZKP
	symmKeyVerified := sharing.Verify(proof, g, PK_j, PK_d, K_j_d, curve)

	if !symmKeyVerified {
		log.Errorf("Verification of ZKP failed in Implicate flow, ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)

		return
	}

	share_j := msg.AcssData.ShareMap[common.PointToHex(PK_j)]
	commitments := msg.AcssData.Commitments

	// Check Predicate for share
	shamirShare_j, _, verified := sharing.Predicate(K_j_d, share_j, commitments, k, common.CurveFromName(msg.CurveName))

	// IF ok: store the decrypted share
	if verified {
		self.State().AcssStore.UpdateAccsState(msg.ACSSRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
			state.VerifiedRecoveryShares[sender.Index] = shamirShare_j
		})
	}
	// If node has received >= t+1 verified shares: interpolate and output
	// TODO check Do we have to get the state again here?
	acssState, _, _ = self.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	if len(acssState.VerifiedRecoveryShares) >= t+1 {
		// TODO interpolate
		// Not clear what existing functions to use for this
	}

}
