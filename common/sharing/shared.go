package sharing

import (
	"fmt"
	"math/big"

	"github.com/arcana-network/dkgnode/curves"
)

func getFixedScalar(c *curves.Curve) (curves.Scalar, error) {
	k256Scalar := "6c47fa13c92d8b47d1579f112657c22ddd0c3a6ed1fb56c8fc80a086477bf89c"
	ed25519Scalar := "19d7725aab29dab57a2124400cb2ca69c9830f691104d1471b8cb0759cd17d1"

	if c.ID == curves.CurveSECP256K1 {
		b2, ok := new(big.Int).SetString(k256Scalar, 16)
		if !ok {
			return nil, fmt.Errorf("invalid hex for scalar for curve %d", c.ID)
		}
		s2, err := c.Scalar.SetBigInt(b2)
		return s2, err
	} else if c.ID == curves.CurveCV25519 {
		b2, ok := new(big.Int).SetString(ed25519Scalar, 16)
		if !ok {
			return nil, fmt.Errorf("invalid hex for scalar for curve %d", c.ID)
		}
		s2, err := c.Scalar.SetBigInt(b2)
		return s2, err
	} else {
		return nil, fmt.Errorf("invalid curve")
	}
}

func CurveParams(curveID curves.CurveID) (curves.Point, curves.Point) {
	var c *curves.Curve

	if curveID == curves.CurveSECP256K1 {
		c = curves.CurveK256()
	} /* else if curveID == crypto.CurveCV25519 {
		c = curves.ED25519()
	}*/

	scalar, err := getFixedScalar(c)
	if err != nil {
		return nil, nil
	}
	// g, h
	return c.NewGeneratorPoint().Mul(scalar), c.NewGeneratorPoint()
}
