package sharing

import "github.com/coinbase/kryptology/pkg/core/curves"

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
