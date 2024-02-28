package dacss

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

var ImplicateMessageType string = "dacss_implicate"

type ImplicateMessage struct {
	RoundID          common.PSSRoundID
	AcssRoundID      common.ACSSRoundID
	Kind             string
	CurveName        common.CurveName // TODO needed?
	SymmetricPubKey  []byte
	SymmetricPrivKey []byte
	Proof            []byte
}

func NewImplicateMessage(roundID common.PSSRoundID, acssID common.ACSSRoundID, curve *curves.Curve, symmetricPubKey []byte, symmetricPrivKey []byte, proof []byte) (*common.PSSMessage, error) {
	m := &ImplicateMessage{
		RoundID:          roundID,
		AcssRoundID:      acssID,
		Kind:             ImplicateMessageType,
		CurveName:        common.CurveName(curve.Name),
		SymmetricPubKey:  symmetricPubKey,
		SymmetricPrivKey: symmetricPrivKey,
		Proof:            proof,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (msg *ImplicateMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	// TODO
}
