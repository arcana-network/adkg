package common

import (
	"crypto/rand"

	// FIXME replace import
	"github.com/coinbase/kryptology/pkg/core/curves"
)

type KeyPair struct {
	PublicKey  curves.Point
	PrivateKey curves.Scalar
}

// generates a random key pair using the specified curve.
func GenerateKeyPair(curve *curves.Curve) KeyPair {
	g := curve.NewGeneratorPoint()
	privateKey := curve.NewScalar().Random(rand.Reader)
	publicKey := g.Mul(privateKey)
	return KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
}
