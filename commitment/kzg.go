package commitment

import (
	"crypto/rand"
	"math/big"

	"github.com/coinbase/kryptology/pkg/core/curves"
	kryptsharing "github.com/coinbase/kryptology/pkg/sharing"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/kzg"
	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/common/sharing"
)

type KZG struct {
	commitments []curves.Point
	curve       *curves.Curve
	// verifier
	srs *kzg.SRS
}

func (v *KZG) Commitments() []curves.Point {
	return v.commitments
}

func (v *KZG) Open(point fr.Element) (kzg.OpeningProof, error) {
	var frCommitment []fr.Element
	for _, commitment := range v.commitments {
		pointValue := commitment.Scalar().BigInt()
		frPoint := fr.NewElement(pointValue.Uint64())
		frCommitment = append(frCommitment, frPoint)
	}
	return kzg.Open(frCommitment, point, v.srs)
}

func (v *KZG) Curve() *curves.Curve {
	return v.curve
}

func NewKZGCommitment(threshold uint32, curve *curves.Curve, poly *kryptsharing.Polynomial) *KZG {
	v := new(KZG)
	v.curve = curve
	v.commitments = make([]curves.Point, threshold)
	for i := range v.commitments {
		base, _ := sharing.CurveParams(curve.Name)
		v.commitments[i] = base.Mul(poly.Coefficients[i])
	}
	return v
}

func NewKZGVerifier(commitments []curves.Point) *KZG {
	k := new(KZG)
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

func (v *KZG) Verify(commitment *kzg.Digest, proof *kzg.OpeningProof, point fr.Element) error {
	return kzg.Verify(commitment, proof, point, v.srs)
}
