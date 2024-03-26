package sharing

import (
	"fmt"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

// TODO: This needs to be confirmed that the matrix is HIM.
// Vandermonde matrix
func CreateHIM(size int, curve *curves.Curve) [][]curves.Scalar {

	him := make([][]curves.Scalar, size)

	for i := 1; i <= size; i++ {
		him[i-1] = make([]curves.Scalar, size)
		him[i-1][0] = curve.Scalar.New(1)
		element := curve.Scalar.New(i)

		for j := 1; j < size; j++ {

			prev := him[i-1][j-1]
			him[i-1][j] = prev.Mul(element)
		}
	}

	return him
}

func HimMultiplication(matrix [][]curves.Scalar, vector []curves.Scalar) ([]curves.Scalar, error) {

	// column size must be equal to the length of vector
	if len(matrix[0]) != len(vector) {
		return nil, fmt.Errorf("number of columns in the matrix must be equal to the size of the vector")
	}

	rows := len(matrix)
	columns := len(matrix[0])
	result := make([]curves.Scalar, rows)

	// initialize the result vector
	curve := curves.K256()
	for i := 0; i < rows; i++ {
		result[i] = curve.Scalar.Zero()
	}

	// multiply
	for i := 0; i < rows; i++ {
		for j := 0; j < columns; j++ {
			result[i] = result[i].Add(matrix[i][j].Mul(vector[j]))
		}
	}

	return result, nil

}
