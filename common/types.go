package common

import (
	"errors"
	"github.com/arcana-network/dkgnode/curves"
)

type CurveName string

var SECP256K1 CurveName = "secp256k1"
var ED25519 CurveName = "ed25519"

func CurveFromName(c CurveName) *curves.Curve {
	switch c {
	case SECP256K1:
		return curves.CurveK256()
	case ED25519:
		return curves.CurveED25519()
	default:
		return curves.CurveK256()
	}
}

const VERSION string = "1"

type phase string
type NodeID string

type ConnectionDetails struct {
	TMP2PConnection string
	P2PConnection   string
}

type DKGMessageRaw struct {
	RoundID RoundID
	Method  string
	Data    []byte
}

const (
	Initial   phase = "INITIAL"
	Started   phase = "STARTED"
	Proposing phase = "PROPOSING"
	Ended     phase = "ENDED"
)

func (p phase) IsValid() error {
	switch p {
	case Initial, Started, Proposing, Ended:
		return nil
	}
	return errors.New("invalid phase")
}

func CreateDKGMessage(r DKGMessageRaw) DKGMessage {
	return DKGMessage{
		Version: KeygenMessageVersion(VERSION),
		RoundID: r.RoundID,
		Method:  r.Method,
		Data:    r.Data,
	}
}

func CreateMessage(id RoundID, kind string, data []byte) DKGMessage {
	return CreateDKGMessage(DKGMessageRaw{
		RoundID: id,
		Method:  kind,
		Data:    data,
	})
}
