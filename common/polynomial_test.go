package common

import (
	"crypto/rand"
	"math"
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
	curve := polyTestCurve()

	// Generates a random polynomial with a random degree.
	randomDegree := mrand.Intn(1000)
	randomPoly := generateRandomPolynomial(randomDegree, curve)

	// Copy the random polynomial.
	newCoeffs := make([]curves.Scalar, randomPoly.Degree()+1)
	copy(newCoeffs, randomPoly.Coefficients)

	// Append zeros to the end.
	nZeros := mrand.Intn(100)
	for range nZeros {
		newCoeffs = append(newCoeffs, curve.Scalar.Zero())
	}

	randomPolyWithZeros := NewPolynomial(newCoeffs, curve)
	assert.True(test, randomPolyWithZeros.Equal(randomPoly))
}

func TestPolyAddRandom(test *testing.T) {
	curve := polyTestCurve()

	// Generates a random polynomial with a random degree.
	randomDegree1 := mrand.Intn(1000)
	randomDegree2 := mrand.Intn(1000)

	randomPoly1 := generateRandomPolynomial(randomDegree1, curve)
	randomPoly2 := generateRandomPolynomial(randomDegree2, curve)

	smallestDegree := int(math.Min(float64(randomDegree1), float64(randomDegree2)))
	highestDegree := int(math.Max(float64(randomDegree1), float64(randomDegree2)))

	sumPoly, err := randomPoly1.Add(randomPoly2)
	assert.Nil(test, err)

	for i := range smallestDegree + 1 {
		sumCoeff := randomPoly1.Coefficients[i].Add(
			randomPoly2.Coefficients[i],
		)

		assert.Zero(test, sumPoly.Coefficients[i].Cmp(sumCoeff))
	}

	if randomDegree1 != randomDegree2 {
		if randomDegree1 == highestDegree {
			for i := smallestDegree + 1; i <= highestDegree; i++ {
				assert.Zero(test, sumPoly.Coefficients[i].Cmp(randomPoly1.Coefficients[i]))
			}
		} else if randomDegree2 == highestDegree {
			for i := smallestDegree + 1; i <= highestDegree; i++ {
				assert.Zero(test, sumPoly.Coefficients[i].Cmp(randomPoly2.Coefficients[i]))
			}
		}
	}
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

func TestPolyMulByConsWithZero(test *testing.T) {
	curve := polyTestCurve()

	// Generates a random polynomial with a random degree.
	randomDegree := mrand.Intn(1000)
	randomPoly := generateRandomPolynomial(randomDegree, curve)
	zeroPoly := NewPolynomial([]curves.Scalar{curve.Scalar.Zero()}, curve)

	mulPoly := randomPoly.MulByConst(curve.Scalar.Zero())

	assert.True(test, mulPoly.Equal(zeroPoly))
}

func TestPolyMulByConsWithOne(test *testing.T) {
	curve := polyTestCurve()

	// Generates a random polynomial with a random degree.
	randomDegree := mrand.Intn(1000)
	randomPoly := generateRandomPolynomial(randomDegree, curve)

	mulPoly := randomPoly.MulByConst(curve.Scalar.One())

	assert.True(test, mulPoly.Equal(randomPoly))
}

func TestPolyMulByConsRandom(test *testing.T) {
	curve := polyTestCurve()

	// Generates a random polynomial with a random degree.
	randomDegree := mrand.Intn(1000)
	randomPoly := generateRandomPolynomial(randomDegree, curve)

	randomConstant := curve.Scalar.Random(rand.Reader)

	mulPoly := randomPoly.MulByConst(randomConstant)
	for i, coeff := range mulPoly.Coefficients {
		multiCoeff := randomConstant.Mul(randomPoly.Coefficients[i])
		assert.Equal(test, 0, coeff.Cmp(multiCoeff))
	}
}

func TestLagrangeBasis(test *testing.T) {
	// Generate a random number of different random values.
	const MAX_CAPACITY int = 100
	const MAX_VALUE int = 1000
	curve := polyTestCurve()

	nXAxis := mrand.Intn(MAX_CAPACITY)
	randomXAxis := make([]curves.Scalar, 0, nXAxis)
	differentChecker := make(map[int]bool) // Stores the numbers created so far to avoid repeated occurrences.
	for len(randomXAxis) < nXAxis {
		randomElem := mrand.Intn(MAX_VALUE)
		if !differentChecker[randomElem] {
			differentChecker[randomElem] = true
			scalarElem := curve.Scalar.New(randomElem)
			randomXAxis = append(randomXAxis, scalarElem)
		}
	}

	basisPolynomials := make([]*Polynomial, nXAxis)
	for j := range nXAxis {
		lagBasisPoly, err := lagrangeBasis(j, randomXAxis, curve)
		assert.Nil(test, err)

		basisPolynomials[j] = lagBasisPoly
	}

	for j, basisPoly := range basisPolynomials {
		for i, xElement := range randomXAxis {
			evaluation := basisPoly.Evaluate(xElement)
			if i == j {
				assert.Zero(test, evaluation.Cmp(curve.Scalar.One()))
			} else {
				assert.Zero(test, evaluation.Cmp(curve.Scalar.Zero()))
			}
		}
	}
}

func TestInterpolatePolynomial(test *testing.T) {
	curve := polyTestCurve()
	points := generateRandomPoints(100, curve)

	interpolatedPoly, err := InterpolatePolynomial(points, curve)
	assert.Nil(test, err)

	for x, y := range points {
		evaluation := interpolatedPoly.Evaluate(
			curve.Scalar.New(x),
		)

		assert.Zero(test, evaluation.Cmp(y))
	}
}

func TestInterpolatePolynomialConstant(test *testing.T) {
	curve := polyTestCurve()
	const MAX_POINTS int = 100
	const MAX_X_AXIS int = 1000

	yValue := curve.Scalar.Random(rand.Reader)

	pointsChecking := make([]curves.Scalar, MAX_POINTS)
	pointsInterpolation := make(map[int]curves.Scalar)
	for i := range MAX_POINTS {
		randomXCheck := mrand.Intn(MAX_X_AXIS)
		pointsChecking[i] = curve.Scalar.New(randomXCheck)

		randomXInterp := mrand.Intn(MAX_X_AXIS)
		pointsInterpolation[randomXInterp] = yValue
	}

	interpolatedPoly, err := InterpolatePolynomial(pointsInterpolation, curve)
	assert.Nil(test, err)

	for _, x := range pointsChecking {
		evaluation := interpolatedPoly.Evaluate(x)

		assert.Zero(test, evaluation.Cmp(yValue))
	}
}

func generateRandomPoints(length int, curve *curves.Curve) map[int]curves.Scalar {
	points := make(map[int]curves.Scalar)
	const MAX_X_AXIS int = 1000
	const MAX_Y_AXIS int = 1000
	for range length {
		randomX := mrand.Intn(MAX_X_AXIS)
		randomY := mrand.Intn(MAX_Y_AXIS)
		randomYScalar := curve.Scalar.New(randomY)
		points[randomX] = randomYScalar
	}
	return points
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
