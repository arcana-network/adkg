package commitment

import (
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg"

	kryptsharing "github.com/coinbase/kryptology/pkg/sharing"
)

type Value struct {
	Points []curves.Point
}

type Commitment []byte
type Opening []byte

type Scheme interface {
	GenerateCommitmentValue() Value
	Setup() kzg.SRS
	Commit(commitmentValue Value) Commitment
	Open() Opening
	Check(opening Opening) bool // == Verify()?
}

type Verifier interface {
	Open(polynomial *kryptsharing.Polynomial, point fr.Element, srs *kzg.SRS) (kzg.OpeningProof, error)
	Verify(commitment *kzg.Digest, proof *kzg.OpeningProof, point fr.Element, srs *kzg.SRS) error
	Commitments() []curves.Point
	Polynomial() *kryptsharing.Polynomial
	Curve() *curves.Curve
}
