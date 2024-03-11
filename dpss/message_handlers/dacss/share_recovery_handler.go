package dacss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

// The message for initializing Share Recovery phase
var ShareRecoveryMessageType string = "dacss_share_recovery"

type ShareRecoveryMessage struct {
	ACSSRoundDetails common.ACSSRoundDetails // ID of the specific ACSS round within DPSS.
	Kind             string                  // Type of the message
	CurveName        common.CurveName        // Name (indicator) of curve used in the messages.
	AcssData         common.AcssData         // ShareMap, commitments & dealer's ephemeral pubkey
}

func NewShareRecoveryMessage(acssRoundDetails common.ACSSRoundDetails, acssData common.AcssData) (*common.PSSMessage, error) {
	m := &ShareRecoveryMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ShareRecoveryMessageType,
		AcssData:         acssData,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, m.Kind, bytes)
	return &msg, nil
}

/*
If the current node already has *output* a valid share, send ReceiveShareRecoveryMessage to the other nodes.
Otherwise, we only indicate that we have arrived in the ShareRecovery phase.
*/
func (msg *ShareRecoveryMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	// Check message comes from node itself
	if !self.Details().IsEqual(sender) {
		return
	}

	self.State().AcssStore.Lock()
	defer self.State().AcssStore.Unlock()

	// Retrieve the ACSS state for the specific ACSS round
	acssState, found, err := self.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	if err != nil || !found {
		return
	}

	// If share recovery is already ongoing, return
	if acssState.ShareRecoveryOngoing {
		return
	}

	// Set in the Node's state that we're in ShareRecovery phase
	self.State().AcssStore.UpdateAccsState(msg.ACSSRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.ShareRecoveryOngoing = true
	})

	// If the current node already has *output* a valid share,
	// send ReceiveShareRecoveryMessage to the other nodes

	// https://eprint.iacr.org/2021/159.pdf 501: "if Pi previously output valid shares (line 307) then
	// Multicast SKi and return"
	// A node has a valid share if it has reached the outputHandler in RBC.
	// At that point ValidShareOutput is set to true in the node's state
	if acssState.ValidShareOutput && len(acssState.AcssDataHash) != 0 {

		priv := self.PrivateKey()
		dealerPubKey, err := common.HexToPoint(msg.CurveName, msg.AcssData.DealerEphemeralPubKey)
		if err != nil {
			return
		}
		symmetricKey, err := sharing.CalculateSharedKey(dealerPubKey, priv)
		if err != nil {
			return
		}
		curve := common.CurveFromName(msg.CurveName)
		pubKeyPoint, err := common.PointToCurvePoint(self.Details().PubKey, msg.CurveName)
		if err != nil {
			log.Errorf("Error converting pubkey to curve point in Share Recovery for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
			return
		}
		proof := sharing.GenerateNIZKProof(curve, priv, pubKeyPoint, dealerPubKey, symmetricKey, curve.NewGeneratorPoint())

		receiveShareRecoveryMsg, err := NewReceiveShareRecoveryMessage(msg.ACSSRoundDetails, msg.CurveName, symmetricKey.ToAffineCompressed(), proof, msg.AcssData)
		if err != nil {
			log.Errorf("Error in creating ReceiveShareRecoveryMessage for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
			return
		}

		// Send the information to all other nodes
		for _, node := range self.Nodes(!self.IsOldNode()) {
			self.Send(node, *receiveShareRecoveryMsg)
		}
	}

	// If the shares weren't validated, we have to wait for other nodes to send us shares to the ReceiveShareRecoveryHandler
}
