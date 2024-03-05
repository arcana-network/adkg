package dacss

import (
	"encoding/hex"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

// TODO docs
// TODO tests
// TODO error handling

var ImplicateReceiveMessageType string = "dacss_implicate_receive"

type ImplicateReceiveMessage struct {
	ACSSRoundDetails common.ACSSRoundDetails // ID of the specific ACSS round within DPSS.
	Kind             string
	CurveName        common.CurveName
	SymmetricKey     []byte // Compressed Affine Point
	Proof            []byte // Contains d, R, S
}

func NewImplicateReceiveMessage(acssRoundDetails common.ACSSRoundDetails, curveName common.CurveName, symmetricKey []byte, proof []byte) (*common.PSSMessage, error) {
	m := &ImplicateReceiveMessage{
		ACSSRoundDetails: acssRoundDetails,
		Kind:             ImplicateReceiveMessageType,
		CurveName:        curveName,
		SymmetricKey:     symmetricKey,
		Proof:            proof,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.ACSSRoundDetails.PSSRoundDetails, m.Kind, bytes)
	return &msg, nil
}

func (msg *ImplicateReceiveMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {

	// First check whether the sharemap for this acss round has already been stored
	dacssState, found, err := self.State().AcssStore.Get(msg.ACSSRoundDetails.ToACSSRoundID())
	if err != nil {
		log.Errorf("Error retrieving ACSS state in implicate flow for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
		return
	}

	// If for this specific ACSS round, we are already in Share Recovery, ignore msg
	if dacssState.ShareRecoveryOngoing {
		return
	}

	curve := curves.GetCurveByName(string(msg.CurveName))

	PK_i, err := curve.Point.Set(&sender.PubKey.X, &sender.PubKey.Y)
	senderPubkeyHex := hex.EncodeToString(PK_i.ToAffineCompressed())

	// If there's no state for this round or the shareMap has not yet been stored
	// we store the symmetric key, proof and sender's public key as hex value
	// The implicate flow should be continued as soon as we have the sharemap
	if !found || dacssState.AcssData.IsUninitialized() {
		self.State().AcssStore.UpdateAccsState(msg.ACSSRoundDetails.ToACSSRoundID(), func(state *common.AccsState) {
			implicateInformation := common.ImplicateInformation{
				SymmetricKey:    msg.SymmetricKey,
				Proof:           msg.Proof,
				SenderPubkeyHex: senderPubkeyHex,
			}
			state.ImplicateInformationSlice = append(state.ImplicateInformationSlice, implicateInformation)
		})
	} else {
		// If the have the shareMap Implicate flow can continue; Send ImplicateExecuteMessage
		implicateExecuteMessage, err := NewImplicateExecuteMessage(msg.ACSSRoundDetails, msg.CurveName, msg.SymmetricKey, msg.Proof, senderPubkeyHex)
		if err != nil {
			log.Errorf("Error creating implicate execute msg in implicate flow for ACSS round %s, err: %s", msg.ACSSRoundDetails.ToACSSRoundID(), err)
			return
		}
		self.ReceiveMessage(self.Details(), *implicateExecuteMessage)
	}

}
