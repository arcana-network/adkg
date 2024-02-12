package keyderivation

import (
	"github.com/arcana-network/dkgnode/common"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/torusresearch/bijson"
)

var PubKeygenType string = "key_derivation_pubkey_gen"

type PubKeygenMessage struct {
	RoundID   common.RoundID
	Kind      string
	Curve     common.CurveName
	PublicKey common.Point
}

func NewPubKeygenMessage(id common.RoundID, curve common.CurveName, publicKey curves.Point) (*common.DKGMessage, error) {
	m := PubKeygenMessage{
		RoundID: id,
		Kind:    PubKeygenType,
		Curve:   curve,
	}

	m.PublicKey = kcommon.CurvePointToPoint(publicKey, curve)

	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}
