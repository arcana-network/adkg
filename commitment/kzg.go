package commitment

import (
	"crypto/rand"
	"log"
	"math/big"

	"github.com/coinbase/kryptology/pkg/core/curves"
	kryptsharing "github.com/coinbase/kryptology/pkg/sharing"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg"

	"github.com/arcana-network/dkgnode/common/sharing"
)

type KZGVerifier struct {
	commitments []curves.Point
	curve       *curves.Curve
	srs         *kzg.SRS
}

func (v *KZGVerifier) Verify(commitment *kzg.Digest, proof *kzg.OpeningProof, point fr.Element, srs *kzg.SRS) error {
	return v.Verify(commitment, proof, point, srs)
}

func (v *KZGVerifier) Commitments() []curves.Point {
	return v.commitments
}

func (v *KZGVerifier) Polynomial() *kryptsharing.Polynomial {
	return v.Polynomial()
}

func (v *KZGVerifier) Open(polynomial *kryptsharing.Polynomial, point fr.Element, srs *kzg.SRS) (kzg.OpeningProof, error) {
	return v.Open(polynomial, point, srs)
}

func (v *KZGVerifier) Curve() *curves.Curve {
	return v.curve
}
func NewKZGVerifier(commitments []curves.Point) *KZGVerifier {
	k := new(KZGVerifier)
	k.commitments = commitments
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(130), nil).Sub(max, big.NewInt(1))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatal(err)
	}
	k.srs, err = kzg.NewSRS(16, n)
	if err != nil {
		log.Fatal(err)
	}
	return k
}

func NewKZGCommitment(threshold uint32, curve *curves.Curve, poly *kryptsharing.Polynomial) *KZGVerifier {
	v := new(KZGVerifier)
	v.curve = curve
	v.commitments = make([]curves.Point, threshold)
	for i := range v.commitments {
		base, _ := sharing.CurveParams(curve.Name)
		v.commitments[i] = base.Mul(poly.Coefficients[i])
	}
	return v
}
