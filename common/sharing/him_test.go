package sharing

import (
	"math/rand"
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

func TestCreateHIM(t *testing.T) {
	him := CreateHIM(7, curves.K256())
	det := determinant(him)
	assert.NotEqual(t, det, curves.K256().Scalar.Zero())

	// every (non-trivial) square submatrix should be invertible
	// iterating for 1000 times
	for i := 0; i < 1000; i++ {

		himIndices := make([]int, 7)
		for i := 0; i < 7; i++ {
			himIndices[i] = i
		}

		// Shuffle the slice
		for i := 7 - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			himIndices[i], himIndices[j] = himIndices[j], himIndices[i]
		}

		// size of the square submatrix
		size := rand.Intn(7) + 1

		// the rows and columns set
		r := himIndices[:size]
		s := himIndices[7-size:]

		submatrix := make([][]curves.Scalar, size)
		for i := 0; i < size; i++ {
			submatrix[i] = make([]curves.Scalar, size)
			for j := 0; j < size; j++ {
				submatrix[i][j] = him[r[i]][s[j]]
			}
		}
		det = determinant(submatrix)
		assert.NotEqual(t, det, curves.K256().Scalar.Zero())
	}

}

func TestHimMultiplication(t *testing.T) {
	him := CreateHIM(10, curves.K256())
	curve := curves.K256()

	// valid multiplication
	vector := make([]curves.Scalar, 10)

	for i := 0; i < 10; i++ {
		vector[i] = curve.Scalar.Zero()
	}

	result, err := HimMultiplication(him, vector)
	assert.Nil(t, err)
	assert.Equal(t, result, vector)

	// check if the matrix column != vector length
	vector = make([]curves.Scalar, 7)
	_, err = HimMultiplication(him, vector)
	assert.NotNil(t, err)

	// another valid multiplication test

	vector = make([]curves.Scalar, 10)
	for i := 0; i < 10; i++ {
		vector[i] = curve.Scalar.One()
	}

	//Expected result
	ExpectedResult := make([]curves.Scalar, 10)

	//initialize
	for i := 0; i < 10; i++ {
		ExpectedResult[i] = curve.Scalar.Zero()
	}

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			ExpectedResult[i] = ExpectedResult[i].Add(him[i][j])
		}
	}

	result, err = HimMultiplication(him, vector)
	assert.Nil(t, err)
	assert.Equal(t, ExpectedResult, result)
}

// Function to calculate the determinant of a square matrix
func determinant(matrix [][]curves.Scalar) curves.Scalar {
	// Check if the matrix is square
	n := len(matrix)
	if n != len(matrix[0]) {
		panic("Input matrix is not square")
	}

	// Base case for 1x1 matrix
	if n == 1 {
		return matrix[0][0]
	}

	curve := curves.K256()
	det := curve.Scalar.Zero()

	// Iterate over the first row to compute the determinant
	for j, element := range matrix[0] {
		// Generate the submatrix by excluding the first row and current column
		submatrix := make([][]curves.Scalar, n-1)
		for i := 1; i < n; i++ {
			submatrix[i-1] = make([]curves.Scalar, n-1)
			copy(submatrix[i-1], matrix[i][:j])
			copy(submatrix[i-1][j:], matrix[i][j+1:])
		}

		// Calculate the determinant recursively for the submatrix and accumulate
		// det += element * determinant(submatrix) * (-1) ^ (j % 2)
		if j%2 == 0 {
			det = det.Add(element.Mul(determinant(submatrix)))
		}
		if j%2 == 1 {
			det = det.Add((element.Mul(determinant(submatrix))).Mul(curve.Scalar.New(-1)))
		}

	}

	return det
}
