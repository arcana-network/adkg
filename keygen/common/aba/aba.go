package aba

import (
	"crypto/sha256"
	"fmt"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

func Hash(g, h, gTilde, hTilde, gI, gITilde curves.Point, curve *curves.Curve) curves.Scalar {
	plaintext := make([]byte, 0)
	plaintext = append(plaintext, g.ToAffineCompressed()...)
	plaintext = append(plaintext, gI.ToAffineCompressed()...)
	plaintext = append(plaintext, h.ToAffineCompressed()...)
	plaintext = append(plaintext, gTilde.ToAffineCompressed()...)
	plaintext = append(plaintext, gITilde.ToAffineCompressed()...)
	plaintext = append(plaintext, hTilde.ToAffineCompressed()...)

	sum := sha256.Sum256(plaintext)
	c := curve.Scalar.Hash(sum[:])

	return c
}

func DerivePublicKey(nodeId, k int, curve *curves.Curve, Tj []int, commitment map[int][]curves.Point) curves.Point {
	x := curve.Scalar.New(nodeId)
	var gI curves.Point
	for l1, l2 := range Tj {
		i := curve.Scalar.One()
		rhs := commitment[l2][0]

		for j := 1; j < k; j++ {
			i = i.Mul(x)
			rhs = rhs.Add(commitment[l2][j].Mul(i))
		}

		if l1 == 0 {
			gI = rhs
		} else {
			gI = gI.Add(rhs)
		}
	}
	return gI
}

func LagrangeCoeffs(
	identities []int,
	curve *curves.Curve) (map[int]curves.Scalar, error) {

	xs := make(map[int]curves.Scalar, len(identities))
	for _, xi := range identities {
		xs[xi] = curve.Scalar.New(xi)
	}

	result := make(map[int]curves.Scalar, len(identities))
	for i, xi := range xs {
		num := curve.Scalar.One()
		den := curve.Scalar.One()
		for j, xj := range xs {
			if i == j {
				continue
			}

			num = num.Mul(xj)
			den = den.Mul(xj.Sub(xi))
		}
		if den.IsZero() {
			return nil, fmt.Errorf("divide by zero")
		}
		result[i] = num.Div(den)
	}

	return result, nil
}
