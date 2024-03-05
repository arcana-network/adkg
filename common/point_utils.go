package common

import (
	"encoding/hex"
	"errors"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

// Converts a point into hex value of the affine compressed representation
func PointToHex(point curves.Point) string {
	return hex.EncodeToString(point.ToAffineCompressed())
}

// HexToPoint converts a hex string to a curves.Point. It returns an error if the conversion fails.
// expects the hex to be created from the affine compressed representation
func HexToPoint(curveName CurveName, hexStr string) (curves.Point, error) {
	bytes, err := hex.DecodeString(hexStr)
	curve := CurveFromName(curveName)
	if err != nil {
		return curve.NewGeneratorPoint(), errors.New("invalid hex string")
	}

	point, err := curve.Point.FromAffineCompressed(bytes)
	if err != nil {
		// Return a generator Point and the error
		return curve.NewGeneratorPoint(), err
	}

	return point, nil
}
