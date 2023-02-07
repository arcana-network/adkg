package commitment

import (
	"github.com/coinbase/kryptology/pkg/core/curves"
	kryptsharing "github.com/coinbase/kryptology/pkg/sharing"

	"github.com/arcana-network/dkgnode/common/sharing"
)

type KZGVerifier struct {
	commitments []curves.Point
}

func (v *KZGVerifier) Verify(share *sharing.ShamirShare) error {
	return nil
}

func (v *KZGVerifier) Commitments() []curves.Point {
	return v.commitments
}

func NewKZGVerifier(commitments []curves.Point) *KZGVerifier {
	k := new(KZGVerifier)
	k.commitments = commitments
	return k
}

func NewKZGCommitment(threshold uint32, curve *curves.Curve, poly *kryptsharing.Polynomial) *KZGVerifier {
	v := new(KZGVerifier)

	v.commitments = make([]curves.Point, threshold)
	for i := range v.commitments {
		base, _ := sharing.CurveParams(curve.Name)
		v.commitments[i] = base.Mul(poly.Coefficients[i])
	}
	return v
}
