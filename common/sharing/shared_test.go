package sharing

import (
	"crypto/rand"
	mrand "math/rand"
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

/*
Functions: CompressShares, DecompressShares.

Testcase: test that the decompression returns the correct scalars after a
compression, i.e. tests that the decompression is the left inverse of compression.
*/
func TestCompressionAndDecompressionShares(test *testing.T) {
	// Generates a random number of random scalars
	curve := curves.K256()
	nScalars := mrand.Intn(1000)
	scalars := make([]curves.Scalar, 0)
	for range nScalars {
		randomScalar := curve.Scalar.Random(rand.Reader)
		scalars = append(scalars, randomScalar)
	}

	compression := CompressShares(scalars)
	decompression, err := DecompressShares(compression, curve, nScalars)
	assert.Nil(test, err)

	for i, shareDecompressed := range decompression {
		assert.Zero(test, shareDecompressed.Cmp(scalars[i]))
	}
}
