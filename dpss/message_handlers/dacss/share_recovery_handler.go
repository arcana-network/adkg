package dacss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/torusresearch/bijson"
)

var ShareRecoveryMessageType string = "dacss_share_recovery"

type ShareRecoveryMessage struct {
	ACSSRoundDetails common.ACSSRoundDetails // ID of the specific ACSS round within DPSS.
	Kind             string                  // Type of the message
	CurveName        common.CurveName        // Name (indicator) of curve used in the messages.
	SenderPubkeyHex  string                  // Hex of Compressed Affine Point
}

func NewShareRecoveryMessage(acssRoundDetails common.ACSSRoundDetails) (*common.PSSMessage, error) {
	m := &ShareRecoveryMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ShareRecoveryMessageType,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, m.Kind, bytes)
	return &msg, nil
}

func (msg *ShareRecoveryMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	// Check message comes from node itself
	if !self.Details().IsEqual(sender) {
		return
	}

	self.State().AcssStore.Lock()
	defer self.State().AcssStore.Unlock()

	dacssState, found, err := self.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	if err != nil || !found {
		return
	}

	// If share recovery is already ongoing, return
	if dacssState.ShareRecoveryOngoing {
		return
	}

	// Set in the Node's state that we're in ShareRecovery phase
	self.State().AcssStore.UpdateAccsState(msg.ACSSRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.ShareRecoveryOngoing = true
	})

	// If the current node already has a valid share,
	// send ReceiveShareRecoveryMessage to the other nodes
	if dacssState.SharesValidated {

		priv := self.PrivateKey()
		dealerPubKey, err := common.HexToPoint(msg.CurveName, msg.SenderPubkeyHex)
		if err != nil {
			return
		}
		symmetricKey, err := sharing.CalculateSharedKey(dealerPubKey, priv)
		if err != nil {
			return
		}
		curve := common.CurveFromName(msg.CurveName)
		pubKeyPoint, err := common.PointToCurvePoint(self.Details().PubKey, msg.CurveName)
		proof := sharing.GenerateNIZKProof(curve, priv, pubKeyPoint, dealerPubKey, symmetricKey, curve.NewGeneratorPoint())

		receiveShareRecoveryMsg, err := NewReceiveShareRecoveryMessage(msg.ACSSRoundDetails, msg.CurveName, symmetricKey.ToAffineCompressed(), proof)

		for _, node := range self.Nodes(!self.IsOldNode()) {
			self.Send(node, *receiveShareRecoveryMsg)
		}
	}

	// If the shares weren't validated, we have to wait for other nodes to send us shares to the ReceiveShareRecoveryHandler
}
