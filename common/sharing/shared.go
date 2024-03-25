package sharing

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/sharing"
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

func GenerateSecret(c *curves.Curve) curves.Scalar {
	secret := c.Scalar.Random(rand.Reader)
	return secret
}

// Concatenates the byte representation of each commitment into an array of
// bytes.
func CompressCommitments(v *sharing.FeldmanVerifier) []byte {
	c := make([]byte, 0)
	for _, v := range v.Commitments {
		e := v.ToAffineCompressed() // 33 bytes
		c = append(c, e[:]...)
	}
	return c
}

func DecompressCommitments(k int, c []byte, curve *curves.Curve) ([]curves.Point, error) {
	commitment := make([]curves.Point, 0)
	for i := 0; i < k; i++ {
		length := 33
		if curve.Name == "ed25519" {
			length = 32
		}
		cI, err := curve.Point.FromAffineCompressed(c[i*length : (i*length)+length])
		if err == nil {
			commitment = append(commitment, cI)
		} else {
			return nil, err
		}
	}

	return commitment, nil
}

// Takes an array of shares [r_1, ..., r_n] and returns the concatenation of
// shares in bytes, that is, it returns Bytes(r_1) || Bytes(r_2) || ... || Bytes(r_n)
func CompressShares(shares []curves.Scalar) []byte {
	sharesBytes := make([]byte, 0)
	for _, share := range shares {
		individualShareBytes := share.Bytes()
		sharesBytes = append(sharesBytes, individualShareBytes...)
	}
	return sharesBytes
}

// Takes the concatenation of bytes of a list of bytes and returns an array with
// with the actual shares. Specifically, it takes the array of bytes consisting of
// Bytes(r_1) || Bytes(r_2) || ... || Bytes(r_n) and returns the array of curve
// scalars [r_1, ..., r_n].
func DecompressShares(sharesBytes []byte, curve *curves.Curve, numShares int) ([]curves.Scalar, error) {
	// Trick to get the number of bytes needed to represent a scalar in the
	// specific curve.
	rndScalar := curve.Scalar.Random(rand.Reader)
	numBytes := len(rndScalar.Bytes())

	sharesScalar := make([]curves.Scalar, 0)
	for i := 0; i < numShares; i++ {
		indivShareScalarBytes := sharesBytes[i*numBytes : (i*numBytes)+numBytes]
		indivShareScalar, err := curve.Scalar.SetBytes(indivShareScalarBytes)
		if err != nil {
			return nil, err
		}
		sharesScalar = append(sharesScalar, indivShareScalar)
	}
	return sharesScalar, nil
}
