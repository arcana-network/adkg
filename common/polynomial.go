package common

import (
	"errors"
	"math"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

// Represents a polynomial with coefficients in a field of scalars of a curve.
//
// A polynomial of the form a_0 + a_1 x + a_2 x^2 + ... + a_n x^n will be
// represented as [a_0, a_1, ..., a_n] and its degree will be n, which is the
// length of the array minus 1.
type Polynomial struct {
	Coefficients []curves.Scalar
	Curve        *curves.Curve
}

// Creates a new polynomial with the given coefficients in a given curve.
func NewPolynomial(coeff []curves.Scalar, curve *curves.Curve) Polynomial {
	return Polynomial{
		Coefficients: coeff,
		Curve:        curve,
	}
}

// Returns the degree of a polynomial.
func (p *Polynomial) Degree() int {
	return len(p.Coefficients) - 1
}

// Returns the addition of the polynomial p and the polynomial q.
func (p *Polynomial) Mul(q *Polynomial) (*Polynomial, error) {
	if p.Curve.Name != q.Curve.Name {
		return nil, errors.New("the scalars used in the polynomials come from different curves")
	}

	degreeNewPoly := p.Degree() + q.Degree()
	coeffsNewPoly := make([]curves.Scalar, degreeNewPoly)

	for i := range coeffsNewPoly {
		coeff := p.Curve.Scalar.Zero()
		for j := range i {
			coeff = coeff.Add(p.Coefficients[j].Mul(q.Coefficients[i-j]))
		}
		coeffsNewPoly[i] = coeff
	}

	newPoly := NewPolynomial(coeffsNewPoly, p.Curve)
	return &newPoly, nil
}

// Returns the multiplication of the polynomial p and the polynomial q.
func (p *Polynomial) Add(q *Polynomial) (*Polynomial, error) {
	if p.Curve.Name != q.Curve.Name {
		return nil, errors.New("the scalars used in the polynomials come from different curves")
	}

	degreeNewPoly := int(math.Max(float64(p.Degree()), float64(q.Degree())))
	coeffsNewPoly := make([]curves.Scalar, degreeNewPoly)

	for i := range degreeNewPoly {
		if i <= p.Degree() && i <= q.Degree() {
			coeffsNewPoly[i] = p.Coefficients[i].Add(q.Coefficients[i])
		} else if i <= p.Degree() {
			coeffsNewPoly[i] = p.Coefficients[i]
		} else if i <= q.Degree() {
			coeffsNewPoly[i] = q.Coefficients[i]
		}
	}

	newPoly := NewPolynomial(coeffsNewPoly, p.Curve)
	return &newPoly, nil
}
