package dacss

import (
	"encoding/hex"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var DacssOutputMessageType string = "dacss_output"

type DacssOutputMessage struct {
	AcssRoundDetails common.ACSSRoundDetails
	kind             string
	curveName        common.CurveName
	m                []byte
}

func NewDacssOutputMessage(roundDetails common.ACSSRoundDetails, data []byte, curveName common.CurveName) (*common.PSSMessage, error) {
	m := DacssOutputMessage{
		AcssRoundDetails: roundDetails,
		kind:             DacssOutputMessageType,
		curveName:        curveName,
		m:                data,
	}

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.AcssRoundDetails.PSSRoundDetails, string(m.kind), bytes)
	return &msg, nil
}

//TODO: This output handler is incomplete
// But is suficient for testing the end to end flow of the DACSS

func (m DacssOutputMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.Debugf("Received output message on %d", self.Details().Index)

	// Ignore if not received by self
	if sender.Index != self.Details().Index {
		return
	}

	_, isStored, err := self.State().AcssStore.Get(m.AcssRoundDetails.ToACSSRoundID())

	if err != nil {
		log.WithField("error", err).Error("NewDacssOutputMessage - Process()")
		return
	}

	if !isStored {
		log.WithField("error", "ACSS state not stored yet").Error("DacssEchoMessage - Process()")
		return
	}

	self.State().AcssStore.Lock()
	defer self.State().AcssStore.Unlock()

	log.Debugf("acss_output: round=%v, self=%v", m.AcssRoundDetails, self.Details().Index)

	priv := self.PrivateKey()

	msgData := common.AcssData{}

	// retrive the ACSSData
	err = bijson.Unmarshal(m.m, msgData)

	if err != nil {
		log.Errorf("Could not deserialize message data, err=%s", err)
		return
	}

	_, k, _ := self.Params()
	curve := common.CurveFromName(m.curveName)

	pubKeyPoint, err := common.PointToCurvePoint(self.Details().PubKey, m.curveName)

	if err != nil {
		log.Errorf("Error converting from point to point: %v", err)
		return
	}

	hexPubKey := hex.EncodeToString(pubKeyPoint.ToAffineCompressed())

	EphemeralPubkeyBytes, err := hex.DecodeString(msgData.DealerEphemeralPubKey)
	if err != nil {
		log.Errorf("Error decoding hex string: %v", err)
		return
	}

	dealerKey, err := curve.Point.FromAffineCompressed(EphemeralPubkeyBytes)

	if err != nil {
		log.Errorf("Error FromAffineCompressed: %v", err)
		return
	}
	key, err := sharing.CalculateSharedKey(dealerKey, priv)

	if err != nil {
		log.Errorf("Error CalculateSharedKey: %v", err)
		return
	}

	share, _, verified := sharing.Predicate(key, msgData.ShareMap[hexPubKey], msgData.Commitments, k, curve)

	if verified {
		log.Debugf("acss_verified: share=%v", *share)

		pubKey := m.AcssRoundDetails.PSSRoundDetails.Dealer.PubKey
		pubKeyCurvePoint, err := common.PointToCurvePoint(pubKey, m.curveName)
		if err != nil {
			log.WithField("error constructing PointToCurvePoint", err).Error("DacssOutputMessage")
			return
		}
		pubKeyHex := common.PointToHex(pubKeyCurvePoint)

		self.State().AcssStore.UpdateAccsState(
			m.AcssRoundDetails.ToACSSRoundID(),
			func(state *common.AccsState) {
				state.RBCState.Phase = common.Ended
				state.ValidShareOutput = true

				//store the shares against the dealer from which it received the valid share
				state.ReceivedShares[pubKeyHex] = share
			},
		)

	} else {
		log.Errorf("didnt pass acss_predicate")
	}

}
