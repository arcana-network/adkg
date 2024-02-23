package dacss

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

// ShareMessageType tells wich message are we sending. In this case, the share
// message.
var ShareMessageType string = "dacss_share"

// DualCommitteeACSSShareMessage has all the information for the initial message in the
// Dual-Committee ACSS Share protocol.
type DualCommitteeACSSShareMessage struct {
	RoundID   common.PSSRoundID  // ID of the round.
	Kind      string             // Type of the message.
	CurveName common.CurveName   // Name of curve used in the messages.
	Secret    curves.Scalar      // Scalar that will be shared.
	Dealer    common.NodeDetails // Information of the node that starts the Dual-Committee ACSS.
}

// NewDualCommitteeACSSShareMessage creates a new share message from the provided ID and
// curve.
func NewDualCommitteeACSSShareMessage(secret curves.Scalar, dealer common.NodeDetails, roundID common.PSSRoundID, curve common.CurveName) (*common.PSSMessage, error) {
	m := &DualCommitteeACSSShareMessage{
		RoundID:   roundID,
		Kind:      ShareMessageType,
		CurveName: curve,
		Secret:    secret,
		Dealer:    dealer,
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

	// FIXME TODO generate new priv key for ephemeral key
	// For now just reusing priv key
	privateKey := self.PrivateKey()

	// Generate share and commitments
	n, k, _ := self.Params()

	// TODO do we need to check whether DPSS has already started?

	ExecuteACSS(true, secret, self, privateKey, curve, n, k)
	ExecuteACSS(false, secret, self, privateKey, curve, n, k)
}

// ExecuteACSS starts the execution of the ACSS protocol with a given committee
// defined by the withOldCommitte flag.
func ExecuteACSS(withOldCommittee bool, secret curves.Scalar, self common.PSSParticipant, privateKey curves.Scalar,
	curve *curves.Curve, n int, k int) {
	// TODO implement this correctly

	// receivingNodes := self.Nodes(withOldCommittee)
	// commitments, shares, err := sharing.GenerateCommitmentAndShares(secret, uint32(k), uint32(n), curve)

	// // Create propose message & broadcast
	// proposeMsg, err := NewHbAacssProposeMessage(msg.RoundID, msgData, msg.Curve, self.Details().Index, true)

	// if err != nil {
	// 	log.Errorf("NewHbAcssPropose:err=%v", err)
	// 	return
	// }

	// // Step 103
	// // ReliableBroadcast(C)
	// go self.Broadcast(true, *proposeMsg)
}
