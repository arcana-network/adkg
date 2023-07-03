package common

import (
	"errors"
	"sort"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

type Poly struct {
	Coeffs []curves.Scalar // Coefficients of the polynomial
}

func minusConst(curve *curves.Curve, c curves.Scalar) *Poly {
	neg := c.Neg()
	return &Poly{
		Coeffs: []curves.Scalar{neg, curve.Scalar.One()},
	}
}

// Mul multiples p and q together. The result is a polynomial of the sum of
// the two degrees of p and q. NOTE: it does not check for null coefficients
// after the multiplication, so the degree of the polynomial is "always" as
// described above. This is only for use in secret sharing schemes. It is not
// a general polynomial multiplication routine.
func (p *Poly) Mul(q *Poly, curve *curves.Curve) *Poly {
	d1 := len(p.Coeffs) - 1
	d2 := len(q.Coeffs) - 1
	newDegree := d1 + d2
	coeffs := make([]curves.Scalar, newDegree+1)
	for i := range coeffs {
		coeffs[i] = curve.Scalar.Zero()
	}
	for i := range p.Coeffs {
		for j := range q.Coeffs {

			tmp := q.Coeffs[j].Mul(p.Coeffs[i])
			coeffs[i+j] = tmp.Add(coeffs[i+j])
		}
	}
	return &Poly{coeffs}
}

// Add computes the component-wise sum of the polynomials p and q and returns it
// as a new polynomial.
func (p *Poly) Add(q *Poly) (*Poly, error) {

	if len(p.Coeffs) != len(q.Coeffs) {
		return nil, errors.New("incorrect coeff length")
	}
	coeffs := make([]curves.Scalar, len(p.Coeffs))

	for i := range coeffs {
		coeffs[i] = p.Coeffs[i].Add(q.Coeffs[i])
	}
	return &Poly{coeffs}, nil
}

// xyScalar returns the list of (x_i, y_i) pairs indexed. The first map returned
// is the list of x_i and the second map is the list of y_i, both indexed in
// their respective map at index i.
func xyScalar(curve *curves.Curve, shares map[int]curves.Scalar, t int) (map[int]curves.Scalar, map[int]curves.Scalar) {
	// we are sorting first the shares since the shares may be unrelated for
	// some applications. In this case, all participants needs to interpolate on
	// the exact same order shares.

	xs := make(map[int]curves.Scalar, 0)
	ys := make(map[int]curves.Scalar, 0)
	temp := make([]int, 0)
	for i := range shares {
		temp = append(temp, i)
	}
	sort.Ints(temp)
	for _, i := range temp {
		xs[i] = curve.Scalar.New(i)
		ys[i] = shares[i]
		if len(xs) == t {
			break
		}
	}

	return xs, ys
}

func lagrangeBasis(curve *curves.Curve, i int, xs map[int]curves.Scalar) *Poly {

	var basis = &Poly{
		Coeffs: []curves.Scalar{curve.Scalar.One()},
	}
	// compute lagrange basis l_j
	var den curves.Scalar
	var acc = curve.Scalar.One()
	for m, xm := range xs {
		if i == m {
			continue
		}
		basis = basis.Mul(minusConst(curve, xm), curve)
		den = xs[i].Sub(xm)   // den = xi - xm
		den, _ = den.Invert() // den = 1 / den
		acc = acc.Mul(den)    // acc = acc * den
	}

	// multiply all coefficients by the denominator
	for i := range basis.Coeffs {
		basis.Coeffs[i] = basis.Coeffs[i].Mul(acc)
	}
	return basis
}

func RecoverPriPoly(curve *curves.Curve, shares map[int]curves.Scalar, t int) (*Poly, error) {

	x, y := xyScalar(curve, shares, t)

	if len(x) != t {
		return nil, errors.New("share: not enough shares to recover private polynomial")
	}

	var accPoly *Poly
	var err error

	for j := range x {
		basis := lagrangeBasis(curve, j, x)
		for i := range basis.Coeffs {
			basis.Coeffs[i] = basis.Coeffs[i].Mul(y[j])
		}

		if accPoly == nil {
			accPoly = basis
			continue
		}

		// add all L_j * y_j together
		accPoly, err = accPoly.Add(basis)
		if err != nil {
			return nil, err
		}
	}
	return accPoly, nil
}
