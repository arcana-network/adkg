package dacss

import (
	"crypto/rand"
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

var InitMessageType string = "hbACSS_init"

// Represents the initialization message for the DPSS protocol.
type InitMessage struct {
	roundID   common.PSSRoundID // ID of the round.
	oldShares []curves.Scalar   // Array of shares that will be converted.
	Kind      string            // Phase in which we are.
	Curve     *curves.Curve     // Curve that we will use for the protocol.
}

// Creates a new initialization message for DPSS.
func NewInitMessage(roundId common.PSSRoundID, oldShares []curves.Scalar, curve curves.Curve) (*common.PSSMessage, error) {
	m := InitMessage{
		roundId,
		oldShares,
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
	// If the node is not an old node, this should not continue.
	if !self.IsOldNode() {
		return
	}

	// Step 101: Sample B / (n - 2t) random elements.
	nNodes, recThreshold, _ := self.Params()
	nGenerations := len(msg.oldShares) / (nNodes + 2*recThreshold)
	for range nGenerations {
		r := msg.Curve.Scalar.Random(rand.Reader)
		msg, err := NewDualCommitteeACSSShareMessage(r, self.Details(), msg.roundID, msg.Curve)
		if err != nil {
			return
		}
		go self.ReceiveMessage(self.Details(), *msg)
	}
}
