package sharing

import (
	"crypto/rand"
	mrand "math/rand"
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

// feldmanSetup creates a Feldman object with parameters generated at random
// using the K256 curve.
func feldmanSetup() *Feldman {
	curve := curves.K256()

	// Generates a random threshold between min and < max
	var minT, maxT uint32
	minT = 2
	maxT = 255
	threshold := mrand.Uint32()%(maxT-minT) + minT

	// Generates a random limit between limit and <254
	var maxL, minL uint32
	maxL = 255
	minL = threshold + 1
	limit := mrand.Uint32()%(maxL-minL) + minL

	feldman, _ := NewFeldman(threshold, limit, curve)
	return feldman
}

// TestFeldmanVerification test if the spliting of a secret with its corresponding
// commitments are succesfully verified by the Verify() function.
func TestFeldmanVerfication(t *testing.T) {
	feldman := feldmanSetup()
	secret := feldman.Curve.Scalar.Random(rand.Reader)

	verifier, shares, err := feldman.Split(secret, rand.Reader)
	if err != nil {
		t.Error("failure during the share construction and commitment.")
	}

	// All shares should verify correctly
	for _, share := range shares {
		result := verifier.Verify(share)
		if result != nil {
			t.Errorf("the share %v was not successfully verified", *share)
		}
	}
}
