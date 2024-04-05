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

	compression := CompressScalars(scalars)
	decompression, err := DecompressScalars(compression, curve, nScalars)
	assert.Nil(test, err)

	for i, shareDecompressed := range decompression {
		assert.Zero(test, shareDecompressed.Cmp(scalars[i]))
	}
}

func TestObtainEvalForX(t *testing.T) {
	curve := curves.K256()

	//iterating for 100 times
	for K := 0; K < 100; K++ {

		secret := curve.Scalar.Random(rand.Reader)

		k := mrand.Intn(10) + 2
		n := mrand.Intn(10) + 2*k + 1
		_, shares, err := GenerateCommitmentAndShares(secret, uint32(k), uint32(n), curve)
		assert.Nil(t, err)

		shamir, err := NewShamir(uint32(k), uint32(n), curve)
		assert.Nil(t, err)

		//converting to correct type
		sharesShamir := make([]*ShamirShare, 0)
		for i := 0; i < n; i++ {
			t := ShamirShare{
				Id:    shares[i].Id,
				Value: shares[i].Value,
			}
			sharesShamir = append(sharesShamir, &t)
		}
		result, err := shamir.ObtainEvalForX(sharesShamir, uint32(0))
		assert.Nil(t, err)
		assert.Equal(t, result, secret)

		for i := 1; i < n; i++ {
			result, err := shamir.ObtainEvalForX(sharesShamir, uint32(i))
			assert.Nil(t, err)
			ExpectedResult, err := curve.Scalar.SetBytes(sharesShamir[i-1].Value)
			assert.Nil(t, err)
			assert.Equal(t, result, ExpectedResult)

		}
	}
}
