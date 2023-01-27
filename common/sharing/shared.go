package sharing

import (
	"fmt"
	"math/big"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

func getFixedScalar(c *curves.Curve) (curves.Scalar, error) {
	k256Scalar := "6c47fa13c92d8b47d1579f112657c22ddd0c3a6ed1fb56c8fc80a086477bf89c"
	ed25519Scalar := "19d7725aab29dab57a2124400cb2ca69c9830f691104d1471b8cb0759cd17d1"

	if c.Name == curves.K256Name {
		b2, ok := new(big.Int).SetString(k256Scalar, 16)
		if !ok {
			return nil, fmt.Errorf("invalid hex for scalar for curve %s ", c.Name)
		}
		s2, err := c.Scalar.SetBigInt(b2)
		return s2, err
	} else if c.Name == curves.ED25519Name {
		b2, ok := new(big.Int).SetString(ed25519Scalar, 16)
		if !ok {
			return nil, fmt.Errorf("invalid hex for scalar for curve %s ", c.Name)
		}
		s2, err := c.Scalar.SetBigInt(b2)
		return s2, err
	} else {
		return nil, fmt.Errorf("Invalid curve")
	}
}

func CurveParams(curveName string) (curves.Point, curves.Point) {
	var c *curves.Curve

	if curveName == "secp256k1" {
		c = curves.K256()
	} else if curveName == "ed25519" {
		c = curves.ED25519()
	}

	scalar, err := getFixedScalar(c)
	if err != nil {
		return nil, nil
	}
	// g, h
	return c.NewGeneratorPoint().Mul(scalar), c.NewGeneratorPoint()
}
