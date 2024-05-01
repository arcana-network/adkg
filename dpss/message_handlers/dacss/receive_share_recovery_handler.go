package dacss

import (
	"reflect"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	kryptology "github.com/coinbase/kryptology/pkg/sharing"

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

func (msg *ReceiveShareRecoveryMessage) Process(sender common.NodeDetails, receiver common.PSSParticipant) {

	// Even if according to receiver node, we're not yet in share recovery phase,
	// try to verify the share and store it in `VerifiedRecoveryShares` if it's valid.
	// This node might still be catching up and need the share later

	// Ignore if message comes from self
	if receiver.Details().IsEqual(sender) {
		return
	}

	receiver.State().AcssStore.Lock()

	// Holding the lock until the end given that the ACSS state is accessed
	// until the end of the function.
	defer receiver.State().AcssStore.Unlock()

	acssState, found, err := receiver.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	if err != nil {
		common.LogStateRetrieveError("ReceiveShareRecoveryHandler", "Process", err)
		return
	}
	if !found {
		common.LogStateNotFoundError("ReceiveShareRecoveryHandler", "Process", found)
		return
	}
	if len(acssState.AcssDataHash) == 0 {
		log.WithFields(
			log.Fields{
				"Message": "the ACSS data hash is empty",
			},
		).Error("ReceiveShareRecoveryHandler: Process")
		return
	}

	// If the current node already has a valid share, ignore the message
	if acssState.ReceivedShare != nil {
		log.Debugf("Node already has a valid share in Receive Share Recovery for ACSS round %s", msg.ACSSRoundDetails.ToACSSRoundID())
		return
	}

	// if state.VerifiedRecoveryShares[sender.Index] != nil, ignore the message
	// we've already received and processed the share from that node
	if acssState.VerifiedRecoveryShares[sender.Index] != nil {
		log.Debugf("already received and processed the share from %v node", sender.Index)
		return
	}

	// Hash the received acssData
	hash, err := common.HashAcssData(msg.AcssData)
	if err != nil {
		log.Errorf("Error hashing acssData in Receive Share Recovery for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
		return
	}

	// Check that the hash of received acssData matches the stored acssDataHash
	if !reflect.DeepEqual(acssState.AcssDataHash, hash) {
		log.Errorf("Received acssDataHash does not match the stored acssDataHash in Receive Share Recovery for ACSS round %s", msg.ACSSRoundDetails.ToACSSRoundID())
		return
	}

	curve := common.CurveFromName(msg.CurveName)

	proof, err := sharing.UnpackProof(curve, msg.Proof)
	if err != nil {
		log.Errorf("Error unpacking proof in Receive Share Recovery for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
		return
	}

	g := curve.NewGeneratorPoint()
	n, k, t := receiver.Params()

	// Convert sender's public key to a curve point
	PK_j, err := common.PointToCurvePoint(sender.PubKey, msg.CurveName)
	if err != nil {
		log.Errorf("Error converting sender's public key to a curve point in Receive Share Recovery for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
		return
	}

	// Convert hex of dealer's ephemeral key to a point
	dealerPubkey := msg.AcssData.DealerEphemeralPubKey
	PK_d, err := common.HexToPoint(msg.CurveName, dealerPubkey)
	if err != nil {
		log.Errorf("Error converting dealer's public key to a curve point in Receive Share Recovery for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
		return
	}

	// Convert received symmetric key to a curve point
	K_j_d, err := curve.Point.FromAffineCompressed(msg.SymmetricKey)
	if err != nil {
		log.Errorf("Error converting symmetric key to a curve point in Receive Share Recovery for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
		return
	}

	// Verify ZKP
	symmKeyVerified := sharing.Verify(proof, g, PK_j, PK_d, K_j_d, curve)

	if !symmKeyVerified {
		log.Errorf("Verification of ZKP failed in Receive Share Recovery for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
		return
	}

	// Extract the correct share
	senderPubkeyHex := common.PointToHex(PK_j)
	share_jAndHmac := msg.AcssData.ShareMap[senderPubkeyHex]

	share_j, _ := sharing.Extract(share_jAndHmac)

	commitments := msg.AcssData.Commitments

	// Check Predicate for share
	shamirShare_j, verifier, verified := sharing.Predicate(K_j_d, share_j, commitments, k, common.CurveFromName(msg.CurveName))

	// If the predicate doesn't check out, we can't store the share
	if !verified {
		log.Errorf("Predicate verification failed in Receive Share Recovery for ACSS round %s", msg.ACSSRoundDetails.ToACSSRoundID())
		return
	}

	// Store the obtained share
	receiver.State().AcssStore.UpdateAccsState(msg.ACSSRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
		state.VerifiedRecoveryShares[sender.Index] = shamirShare_j
	})

	// If node has received >= t+1 verified shares: interpolate and output
	// At this point we already know the acssState exists
	if len(acssState.VerifiedRecoveryShares) >= t+1 {

		//once the threshold is reached, update the state
		receiver.State().AcssStore.UpdateAccsState(
			msg.ACSSRoundDetails.ToACSSRoundID(),
			func(state *common.AccsState) {
				state.ValidShareOutput = true
			},
		)
		shamir, err := sharing.NewShamir(uint32(k), uint32(n), curve)
		if err != nil {
			log.Errorf("Error creating Shamir in Receive Share Recovery for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
			return
		}
		values := make([]*kryptology.ShamirShare, 0, len(acssState.VerifiedRecoveryShares))

		for _, v := range acssState.VerifiedRecoveryShares {
			values = append(values, v)
		}
		convertedShares := make([]*sharing.ShamirShare, len(values))
		for i, share := range values {
			convertedShares[i] = &sharing.ShamirShare{
				Id:    share.Id,
				Value: share.Value,
			}
		}
		// Obtain secret through interpolation
		evalForNode, err := shamir.ObtainEvalForX(convertedShares, uint32(receiver.Details().Index))

		if err != nil {
			log.Errorf("Error obtaining share value for node in Receive Share Recovery for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
			return
		}
		shareForNode := &kryptology.ShamirShare{
			Id:    uint32(receiver.Details().Index),
			Value: evalForNode.Bytes(),
		}

		// When finished, save the share + set RBC phase to ended
		receiver.State().AcssStore.UpdateAccsState(msg.ACSSRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
			state.ReceivedShare = shareForNode
			state.RBCState.Phase = common.Ended
		})

		// commit msg
		commitmentMsg, err := NewDacssCommitmentMessage(
			msg.ACSSRoundDetails,
			msg.CurveName,
			verifier.Commitments[0],
		)

		if err != nil {
			common.LogErrorNewMessage("ReceiveShareRecoveryMessage", "Process", DacssCommitmentMessageType, err)
			return
		}

		go receiver.Broadcast(!receiver.IsNewNode(), *commitmentMsg)
	}

}
