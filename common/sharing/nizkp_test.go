package sharing

import (
	"crypto/rand"
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

func TestValidNIZKP(t *testing.T) {

	for i := 0; i < 20; i++ {
		curve := curves.ED25519()
		SK_i := curve.Scalar.Random(rand.Reader)
		g := curve.NewGeneratorPoint()
		PK_i := g.Mul(SK_i) // g^SK_i

		PK_d := curve.Point.Random(rand.Reader)

		K_i_d := PK_d.Mul(SK_i)

		//completeness
		proofBytes := GenerateNIZKProof(curve, SK_i, PK_i, PK_d, K_i_d, g)

		proof, err := unpackProof(curve, proofBytes)

		assert.Nil(t, err)
		assert.Equal(t, Verify(proof, g, PK_i, PK_d, K_i_d, curve), true)
	}

}

func TestInvalidNIZKP(t *testing.T) {
	curve := curves.ED25519()
	SK_i := curve.Scalar.Random(rand.Reader)
	g := curve.NewGeneratorPoint()
	PK_i := g.Mul(SK_i) // g^SK_i

	PK_d := curve.Point.Random(rand.Reader)

	//soundness
	//trying to prove for random symmetric shared key
	for i := 0; i < 10; i++ {
		K_i_d := curve.Point.Random(rand.Reader)
		proofBytes := GenerateNIZKProof(curve, SK_i, PK_i, PK_d, K_i_d, g)
		proof, err := unpackProof(curve, proofBytes)
		assert.Nil(t, err)
		assert.Equal(t, Verify(proof, g, PK_i, PK_d, K_i_d, curve), false)
	}

	//trying to prove for random PK_i != g^SK_i
	for i := 0; i < 10; i++ {

		PK_i := curve.Point.Random(rand.Reader)
		K_i_d := PK_d.Mul(SK_i)
		proofBytes := GenerateNIZKProof(curve, SK_i, PK_i, PK_d, K_i_d, g)
		proof, err := unpackProof(curve, proofBytes)
		assert.Nil(t, err)
		assert.Equal(t, Verify(proof, g, PK_i, PK_d, K_i_d, curve), false)
	}
}
