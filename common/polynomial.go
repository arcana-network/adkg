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
	Coefficients []curves.Scalar // Coefficients of the polynomial.
	Curve        *curves.Curve   // Curve in which the coefficients live in.
}

// Creates a new polynomial with the given coefficients in a given curve.
func NewPolynomial(coeff []curves.Scalar, curve *curves.Curve) *Polynomial {
	newPoly := &Polynomial{
		Coefficients: coeff,
		Curve:        curve,
	}
	newPoly.Normalize()
	return newPoly
}

// Returns the degree of a polynomial.
func (p *Polynomial) Degree() int {
	return len(p.Coefficients) - 1
}

func (p *Polynomial) Equal(q *Polynomial) bool {
	p.Normalize()
	q.Normalize()
	if p.Curve.Name != q.Curve.Name {
		return false
	}

	if p.Degree() != q.Degree() {
		return false
	}

	for i, pCoeff := range p.Coefficients {
		if pCoeff.Cmp(q.Coefficients[i]) != 0 {
			return false
		}
	}

	return true
}

// Removes the coefficients that are zero after the most signifficant
// coefficient.
func (p *Polynomial) Normalize() {
	var i int
	for i = len(p.Coefficients) - 1; i >= 0; i-- {
		if p.Coefficients[i].Cmp(p.Curve.Scalar.Zero()) != 0 {
			break
		}
	}
	if i == -1 {
		p.Coefficients = []curves.Scalar{p.Curve.Scalar.Zero()}
	}
	if i >= 0 {
		p.Coefficients = p.Coefficients[:i+1]
	}
}

// Returns the addition of the polynomial p and the polynomial q.
func (p *Polynomial) Mul(q *Polynomial) (*Polynomial, error) {
	if p.Curve.Name != q.Curve.Name {
		return nil, errors.New("the scalars used in the polynomials come from different curves")
	}

	degreeNewPoly := p.Degree() + q.Degree()
	coeffsNewPoly := make([]curves.Scalar, degreeNewPoly+1)

	for i := range degreeNewPoly + 1 {
		coeff := p.Curve.Scalar.Zero()
		for j := range i + 1 {
			if j > p.Degree() {
				continue
			}
			if i-j > q.Degree() {
				continue
			}
			coeff = coeff.Add(p.Coefficients[j].Mul(q.Coefficients[i-j]))
		}
		coeffsNewPoly[i] = coeff
	}

	newPoly := NewPolynomial(coeffsNewPoly, p.Curve)
	newPoly.Normalize()
	return newPoly, nil
}

// Returns the multiplication of the polynomial p and the polynomial q.
func (p *Polynomial) Add(q *Polynomial) (*Polynomial, error) {
	if p.Curve.Name != q.Curve.Name {
		return nil, errors.New("the scalars used in the polynomials come from different curves")
	}

	degreeNewPoly := int(math.Max(float64(p.Degree()), float64(q.Degree())))
	coeffsNewPoly := make([]curves.Scalar, degreeNewPoly+1)

	for i := range degreeNewPoly + 1 {
		if i <= p.Degree() && i <= q.Degree() {
			coeffsNewPoly[i] = p.Coefficients[i].Add(q.Coefficients[i])
		} else if i <= p.Degree() {
			coeffsNewPoly[i] = p.Coefficients[i]
		} else if i <= q.Degree() {
			coeffsNewPoly[i] = q.Coefficients[i]
		}
	}

	newPoly := NewPolynomial(coeffsNewPoly, p.Curve)
	newPoly.Normalize()
	return newPoly, nil
}

// Given a polynomial p(x) and a constant c, computes the polynomial c * p(x)
func (p *Polynomial) MulByConst(constant curves.Scalar) *Polynomial {
	coeffsNewPoly := make([]curves.Scalar, p.Degree()+1)
	for i, coeff := range p.Coefficients {
		coeffsNewPoly[i] = coeff.Mul(constant)
	}

	newPoly := NewPolynomial(coeffsNewPoly, p.Curve)
	newPoly.Normalize()
	return newPoly
}

// Computes the Lagrange basis polynomial, that is it computes
// $l_j(x) = \prod_{0 \leq m \leq k, m \neq j} \frac{x - x_m}{x_j - x_m}$
func lagrangeBasis(j int, xAxisValues []curves.Scalar, curve *curves.Curve) (*Polynomial, error) {
	xj := xAxisValues[j]
	lagrangeBasisPoly := NewPolynomial(
		[]curves.Scalar{curve.Scalar.Zero()},
		curve,
	)

	for _, xm := range xAxisValues {
		if xj.Cmp(xm) == 0 {
			continue
		}

		// Computes xj - xm
		denominator := xj.Add(xm.Neg())

		// Computes the polynomial (x - xm) / (xj - xm) that we will call "the
		// linear polinomial". Notice that this polinomial is equal to
		// (1 / (xj - xm)) x + (-xm / (xj - xm))
		coeffsLinearPoly := make([]curves.Scalar, 2)
		denomInverted, err := denominator.Invert()
		if err != nil {
			return nil, err
		}

		coeffsLinearPoly[0] = denomInverted
		coeffsLinearPoly[1] = xm.Neg().Mul(denomInverted)
		linearPolynomial := NewPolynomial(coeffsLinearPoly, curve)

		lagrangeBasisPoly, err = lagrangeBasisPoly.Mul(linearPolynomial)
		if err != nil {
			return nil, err
		}
	}
	lagrangeBasisPoly.Normalize()
	return lagrangeBasisPoly, nil
}

// Computes the lists (x_1, ..., x_k) and (y_1, ..., y_k) from a datastructure
// with pairs (x_1, y_1), ..., (x_k, y_k)
func axisValues(points map[int]curves.Scalar, curve *curves.Curve) ([]curves.Scalar, []curves.Scalar) {
	xAxisValues := make([]curves.Scalar, len(points))
	yAxisValues := make([]curves.Scalar, len(points))
	index := 0
	for x, y := range points {
		xAxisValues[index] = curve.Scalar.New(x)
		yAxisValues[index] = y
		index++
	}
	return xAxisValues, yAxisValues
}

// Given points (x_1, y_1), ..., (x_k, y_k), computes the interpolation of the
// polynomial of degree k + 1 for those points.
func InterpolatePolynomial(points map[int]curves.Scalar, curve *curves.Curve) (*Polynomial, error) {
	xAxisValues, yAxisValues := axisValues(points, curve)
	interPolinomial := NewPolynomial(
		[]curves.Scalar{curve.Scalar.Zero()},
		curve,
	)

	var err error
	for j, y := range yAxisValues {
		lagrangeBasisPoly, errLagBas := lagrangeBasis(j, xAxisValues, curve)
		if errLagBas != nil {
			return nil, errLagBas
		}
		interPolinomial, err = interPolinomial.Add(lagrangeBasisPoly.MulByConst(y))
		if err != nil {
			return nil, err
		}
	}

	return interPolinomial, nil
}

// Evaluate evaluates the polynomial at the given point
func (p *Polynomial) Evaluate(x curves.Scalar) curves.Scalar {
	curve := p.Curve
	result := curve.Scalar.Zero()
	power := curve.Scalar.One()

	// Iterate over the coefficients and compute the polynomial value
	for _, coeff := range p.Coefficients {
		// Add the product of the coefficient and the current power of x to the result
		term := coeff.Mul(power)
		result = result.Add(term)

		// Update the power of x for the next iteration
		power = power.Mul(x)
	}

	return result
}
