package dacss

import (
	"crypto/rand"

	"math"

	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var InitMessageType string = "dacss_init"

// Represents the initialization message for the DPSS protocol.
type InitMessage struct {
	PSSRoundDetails    common.PSSRoundDetails // ID of the round.
	OldShares          []common.PrivKeyShare  // Array of shares that will be converted.
	EphemeralSecretKey []byte                 // the dealer's ephemeral secret key at the start of the protocol (Section V(C)hbACSS)
	EphemeralPublicKey []byte                 // the dealer's ephemeral public key.
	Kind               string                 // Phase in which we are.
	CurveName          *common.CurveName      // Curve that we will use for the protocol.
	NewCommitteeParams common.CommitteeParams // n, k & t parameters of the new committee
}

// Creates a new initialization message for DPSS.
func NewInitMessage(pssRoundDetails common.PSSRoundDetails, oldShares []common.PrivKeyShare, curve common.CurveName, ephemeralKeypair common.KeyPair, newCommitteeParams common.CommitteeParams) (*common.PSSMessage, error) {
	m := InitMessage{
		pssRoundDetails,
		oldShares,
		ephemeralKeypair.PrivateKey.Bytes(),
		ephemeralKeypair.PublicKey.ToAffineCompressed(),
		InitMessageType,
		&curve,
		newCommitteeParams,
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
	if self.IsNewNode() {
		log.WithFields(log.Fields{
			"IsNewNode": self.IsNewNode(),
			"Message":   "Self is expected to be an old node.",
		}).Error("DACSSInitMessage: Process")
		return
	}

	if !sender.IsEqual(self.Details()) {
		log.WithFields(log.Fields{
			"Sender.Index": sender.Index,
			"Self.Index":   self.Details().Index,
			"Message":      "Not equal. Expected to be equal",
		}).Error("DACSSInitMessage: Process")
		return
	}

	// Store the old shares in the local database
	self.State().ShareStore.Lock()

	// Using defer because the state is accessed until the end of the function.
	defer self.State().ShareStore.Unlock()

	self.State().ShareStore.Initialize(len(msg.OldShares))
	for i, share := range msg.OldShares {
		self.State().ShareStore.OldShares[i] = share
	}

	// Step 101: Sample B / (n - 2t) random elements.
	nNodes, _, recThreshold := self.Params()

	nGenerations := int(math.Ceil(float64(len(msg.OldShares)) / float64((nNodes - 2*recThreshold))))
	for i := range nGenerations {
		r := curve.Scalar.Random(rand.Reader)
		acssRoundDetails := common.ACSSRoundDetails{
			PSSRoundDetails: msg.PSSRoundDetails,
			ACSSCount:       i,
		}

		// store the random secret
		_, err := self.State().AcssStore.UpdateAccsState(
			acssRoundDetails.ToACSSRoundID(),
			func(state *common.AccsState) {
				state.RandomSecretShared[acssRoundDetails.ToACSSRoundID()] = &r
			},
		)
		if err != nil {
			common.LogStateUpdateError("InitHandler", "Process", common.AcssStateType, err)
			return
		}

		msg, err := NewDualCommitteeACSSShareMessage(r, self.Details(), acssRoundDetails, curve, msg.EphemeralSecretKey, msg.EphemeralPublicKey, msg.NewCommitteeParams)
		if err != nil {
			common.LogErrorNewMessage("InitMessageHandler", "Process", ShareMessageType, err)
			return
		}
		//NOTE: since the msg is sent to self, we can keep the EmephemeralKeypair in the msg
		go self.Send(self.Details(), *msg)
	}
}
