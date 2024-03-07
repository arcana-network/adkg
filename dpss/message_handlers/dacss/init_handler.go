package dacss

import (
	"crypto/rand"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/torusresearch/bijson"
)

var InitMessageType string = "dacss_init"

// Represents the initialization message for the DPSS protocol.
type InitMessage struct {
	PSSRoundDetails    common.PSSRoundDetails // ID of the round.
	OldShares          []sharing.ShamirShare  // Array of shares that will be converted.
	EphemeralSecretKey []byte                 // the dealer's ephemeral secret key at the start of the protocol (Section V(C)hbACSS)
	EphemeralPublicKey []byte                 // the dealer's ephemeral public key.
	Kind               string                 // Phase in which we are.
	CurveName          *common.CurveName      // Curve that we will use for the protocol.
}

// Creates a new initialization message for DPSS.
func NewInitMessage(pssRoundDetails common.PSSRoundDetails, oldShares []sharing.ShamirShare, curve common.CurveName, ephemeralKeypair common.KeyPair) (*common.PSSMessage, error) {
	m := InitMessage{
		pssRoundDetails,
		oldShares,
		ephemeralKeypair.PrivateKey.Bytes(),
		ephemeralKeypair.PublicKey.ToAffineCompressed(),
		InitMessageType,
		&curve,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(pssRoundDetails, m.Kind, bytes)
	return &msg, nil
}

// Process processes an incommint InitMessage.
func (msg InitMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	curve := common.CurveFromName(*msg.CurveName)
	// If the node is not an old node, this should not continue.
	if !self.IsOldNode() {
		return
	}

	if !sender.IsEqual(self.Details()) {
		return
	}

	// Step 101: Sample B / (n - 2t) random elements.
	nNodes, recThreshold, _ := self.Params()
	nGenerations := len(msg.OldShares) / (nNodes - 2*recThreshold)
	for i := range nGenerations {
		r := curve.Scalar.Random(rand.Reader)
		acssRoundDetails := common.ACSSRoundDetails{
			PSSRoundDetails: msg.PSSRoundDetails,
			ACSSCount:       i,
		}

		msg, err := NewDualCommitteeACSSShareMessage(r, self.Details(), acssRoundDetails, curve, msg.EphemeralSecretKey, msg.EphemeralPublicKey)
		if err != nil {
			return
		}
		//NOTE: since the msg is sent to self, we can keep the EmephemeralKeypair in the msg
		self.Send(self.Details(), *msg)
	}
}
