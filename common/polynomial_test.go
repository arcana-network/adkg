package common

import (
	"crypto/rand"
	mrand "math/rand"
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPolyAddWithZero(test *testing.T) {
	curve := polyTestCurve()

	// Generates a random polynomial with a random degree.
	randomDegree := mrand.Intn(1000)
	randomPoly := generateRandomPolynomial(randomDegree, curve)

	zeroPoly := NewPolynomial([]curves.Scalar{curve.Scalar.Zero()}, curve)

	sumPoly, err := randomPoly.Add(zeroPoly)
	assert.Nil(test, err)

	assert.True(test, sumPoly.Equal(randomPoly))
}

func TestNormalize(test *testing.T) {

}

func TestPolyAddRandom(test *testing.T) {

}

func TestPolyMulWithZero(test *testing.T) {
	log.SetLevel(log.DebugLevel)
	curve := polyTestCurve()

	// Generates a random polynomial with a random degree.
	randomDegree := mrand.Intn(1000)
	randomPoly := generateRandomPolynomial(randomDegree, curve)

	zeroPoly := NewPolynomial([]curves.Scalar{curve.Scalar.Zero()}, curve)

	mulPoly, err := randomPoly.Mul(zeroPoly)
	assert.Nil(test, err)

	log.WithFields(
		log.Fields{
			"Coefficients": mulPoly.Coefficients,
		},
	).Debug("TestPolyMulWithZero")
	assert.True(test, mulPoly.Equal(zeroPoly))
}

func TestPolyMulWithOne(test *testing.T) {
	curve := polyTestCurve()

	// Generates a random polynomial with a random degree.
	randomDegree := mrand.Intn(1000)
	randomPoly := generateRandomPolynomial(randomDegree, curve)

	onePoly := NewPolynomial([]curves.Scalar{curve.Scalar.One()}, curve)

	mulPoly, err := randomPoly.Mul(onePoly)
	assert.Nil(test, err)

	assert.True(test, mulPoly.Equal(randomPoly))
}

func TestPolyMul(test *testing.T) {

}

func TestPolyMulByConsWithZero(test *testing.T) {

}

func TestPolyMulByConsWithOne(test *testing.T) {

}

func TestPolyMulByConsRandom(test *testing.T) {

}

func TestLagrangeBasis(test *testing.T) {

}

func generateRandomPolynomial(degree int, curve *curves.Curve) *Polynomial {
	coefficients := make([]curves.Scalar, degree+1)
	for i := range degree + 1 {
		coefficients[i] = curve.Scalar.Random(rand.Reader)
	}

	return NewPolynomial(coefficients, curve)
}

func polyTestCurve() *curves.Curve {
	return curves.K256()
}

func differentPolyTestCurves() (*curves.Curve, *curves.Curve) {
	return curves.K256(), curves.ED25519()
}
