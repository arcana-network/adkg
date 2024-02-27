package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

// This code comes directly from kryptology and will have a different implementation when this lib is replaced
// (temporary code)

func PointMarshalJson(point curves.Point) ([]byte, error) {
	m := make(map[string]string, 2)
	m["type"] = point.CurveName()
	m["value"] = hex.EncodeToString(point.ToAffineCompressed())
	return json.Marshal(m)
}

func PointUnmarshalJson(input []byte) (curves.Point, error) {
	var m map[string]string

	err := json.Unmarshal(input, &m)
	if err != nil {
		return nil, err
	}
	curve := curves.GetCurveByName(m["type"])
	if curve == nil {
		return nil, fmt.Errorf("invalid type")
	}
	p, err := hex.DecodeString(m["value"])
	if err != nil {
		return nil, err
	}
	P, err := curve.Point.FromAffineCompressed(p)
	if err != nil {
		return nil, err
	}
	return P, nil
}

func ScalarMarshalJson(scalar curves.Scalar) ([]byte, error) {
	m := make(map[string]string, 2)
	m["type"] = scalar.Point().CurveName()
	m["value"] = hex.EncodeToString(scalar.Bytes())
	return json.Marshal(m)
}

func ScalarUnmarshalJson(input []byte) (curves.Scalar, error) {
	var m map[string]string

	err := json.Unmarshal(input, &m)
	if err != nil {
		return nil, err
	}
	curve := curves.GetCurveByName(m["type"])
	if curve == nil {
		return nil, fmt.Errorf("invalid type")
	}
	s, err := hex.DecodeString(m["value"])
	if err != nil {
		return nil, err
	}
	S, err := curve.Scalar.SetBytes(s)
	if err != nil {
		return nil, err
	}
	return S, nil
}
