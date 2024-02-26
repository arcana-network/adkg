package sharing

import (
	"crypto/rand"
	"crypto/sha256"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

//This file includes the NIZK Proof and verifcation generation for IMPLICATE

func GenerateNIZKProof(
	curve *curves.Curve,
	SK_i curves.Scalar, // Secret key of the node i
	PK_i curves.Point, // Public key of the node i
	PK_d curves.Point, // Public key of the dealer
	K_i_d curves.Point, // symmetric shared key of node i and dealer d
	g curves.Point, // generator of the group
) []byte {
	r := curve.NewScalar().Random(rand.Reader)

	R := g.Mul(r)
	S := PK_d.Mul(r)

	e := Hash(g, PK_i, PK_d, K_i_d, R, S, curve)

	//SK_i * e + r
	d := SK_i.MulAdd(e, r)

	proof := make([]byte, 0)
	proof = append(proof, d.Bytes()...)              // d is 32 bytes
	proof = append(proof, R.ToAffineCompressed()...) // 32 bytes for ed25519
	proof = append(proof, S.ToAffineCompressed()...) // 32 bytes for ed25519

	return proof
}
func verify(u *Proof, g, PK_i, PK_d, K_i_d curves.Point, curve *curves.Curve) bool {
	e := Hash(g, PK_i, PK_d, K_i_d, u.R, u.S, curve)

	alpha := (PK_i.Mul(e)).Add(u.R)
	alphaDash := g.Mul(u.d)

	beta := (K_i_d.Mul(e)).Add(u.S)
	betaDash := PK_d.Mul(u.d)

	if alpha.Equal(alphaDash) && beta.Equal(betaDash) {
		return true
	}

	return false
}

type Proof struct {
	d curves.Scalar
	R curves.Point
	S curves.Point
}

func unpackProof(curve *curves.Curve, proofBytes []byte) (*Proof, error) {
	proof := Proof{}

	d, err := curve.Scalar.SetBytes(proofBytes[:32])
	if err != nil {
		return nil, err
	}
	proof.d = d

	R, err := curve.Point.FromAffineCompressed(proofBytes[32:64])
	if err != nil {
		return nil, err
	}
	proof.R = R

	S, err := curve.Point.FromAffineCompressed(proofBytes[64:96])
	if err != nil {
		return nil, err
	}
	proof.S = S

	return &proof, nil
}

func Hash(g, PK_i, PK_d, K_i_d, R, S curves.Point, curve *curves.Curve) curves.Scalar {
	plaintext := make([]byte, 0)
	plaintext = append(plaintext, g.ToAffineCompressed()...)
	plaintext = append(plaintext, PK_i.ToAffineCompressed()...)
	plaintext = append(plaintext, PK_d.ToAffineCompressed()...)
	plaintext = append(plaintext, K_i_d.ToAffineCompressed()...)
	plaintext = append(plaintext, R.ToAffineCompressed()...)
	plaintext = append(plaintext, S.ToAffineCompressed()...)

	sum := sha256.Sum256(plaintext)
	c := curve.Scalar.Hash(sum[:])

	return c
}
