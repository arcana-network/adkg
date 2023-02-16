package commitment

import (
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/kzg"
)

type Value struct {
	Points []curves.Point
}

type Commitment []byte
type Opening []byte

//type Scheme interface {
//	GenerateCommitmentValue() Value
//	Setup() kzg.SRS
//	Commit(commitmentValue Value) Commitment
//	Open() Opening
//	Check(opening Opening) bool // == Verify()?
//}

type Scheme interface {
	Open(point fr.Element) (kzg.OpeningProof, error)
	Commitments() []curves.Point
	Curve() *curves.Curve
}

type Verifier interface {
	Verify(commitment *kzg.Digest, proof *kzg.OpeningProof, point fr.Element) error
}
