package dacss

import (
	"crypto/rand"
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
)

var InitMessageType string = "dacss_init"

// Represents the initialization message for the DPSS protocol.
type InitMessage struct {
	RoundID          common.PSSRoundID     // ID of the round.
	OldShares        []sharing.ShamirShare // Array of shares that will be converted.
	EphemeralKeypair common.KeyPair        // the dealer's ephemeral keypair at the start of the protocol (Section V(C)hbACSS)
	Kind             string                // Phase in which we are.
	CurveName        *common.CurveName     // Curve that we will use for the protocol.
}

// Creates a new initialization message for DPSS.
func NewInitMessage(roundId common.PSSRoundID, oldShares []sharing.ShamirShare, curve common.CurveName, ephemeralKeypair common.KeyPair) (*common.PSSMessage, error) {
	m := InitMessage{
		roundId,
		oldShares,
		ephemeralKeypair,
		InitMessageType,
		&curve,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(roundId, m.Kind, bytes)
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
	for range nGenerations {
		r := curve.Scalar.Random(rand.Reader)
		msg, err := NewDualCommitteeACSSShareMessage(r, self.Details(), msg.RoundID, curve, msg.EphemeralKeypair)
		if err != nil {
			return
		}
		//NOTE: since the msg is sent to self, we can keep the EmephemeralKeypair in the msg
		go self.Send(self.Details(), *msg)
	}
}
