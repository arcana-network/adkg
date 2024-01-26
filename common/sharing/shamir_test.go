package sharing

import (
	"crypto/rand"
	mrand "math/rand"
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

// shamirSetup creates a Shamir object with the curve K256 and choosing the
// threshold and limit at random.
func shamirSetup() *Shamir {
	// Select the curve at random
	curves := []*curves.Curve{curves.K256(), curves.ED25519()}
	curve := curves[mrand.Intn(len(curves))]

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

	shamir, _ := NewShamir(threshold, limit, curve)
	return shamir
}

// TestConstructionReconstruction computes the shares of a secret and then
// apply the reconstruction algorithm. Both values, the secret and the
// reconstruction, should be the same.
func TestConstructionReconstruction(t *testing.T) {
	shamir := shamirSetup()
	secret := shamir.curve.Scalar.Random(rand.Reader)
	shares, err := shamir.Split(secret, rand.Reader)

	if err != nil {
		t.Errorf("failure in the share generation: %v", err)
	}

	scalar, err := shamir.Combine(shares[:]...)
	if err != nil {
		t.Errorf("failure in the share reconstruction: %v", err)
	}

	comparison := secret.Cmp(scalar)

	if comparison != 0 {
		t.Errorf("the reconstructed value is not equal to the secret. Secret: %v, RecVal: %v", secret, scalar)
	}
}

// testGetPolyAndShares tests the correct generation of polynomials. This includes
// the checking of the degrees, the amount of shares generated and that the shares
// are the corresponding evaluation of the polynomials.
func TestGetPolyAndShares(t *testing.T) {
	shamir := shamirSetup()
	secret := shamir.curve.Scalar.Random(rand.Reader)
	shares, poly := shamir.getPolyAndShares(secret, rand.Reader)

	if len(poly.Coefficients) != int(shamir.threshold) {
		t.Errorf(
			"the polynomial has degree %d and the threshold is %d",
			len(poly.Coefficients),
			shamir.threshold,
		)
	}

	if len(shares) != int(shamir.limit) {
		t.Errorf(
			"the ammount of shares (%d) is different to the number of parties (%d).",
			len(shares),
			shamir.limit,
		)
	}

	for i, share := range shares {
		eval := poly.Evaluate(shamir.curve.Scalar.New(i + 1))
		shareField, err := shamir.curve.Scalar.SetBytes(share.Value)
		if err != nil {
			t.Error("the conversion from bytes to scalar failed.")
		}

		if eval.Cmp(shareField) != 0 {
			t.Error("a share and its evaluation in the polynomial are not the same.")
		}
	}

	if poly.Coefficients[0].Cmp(secret) != 0 {
		t.Error("the constant in the polynomial is different to the secret value.")
	}
}

// TestCombinePoints test if the combination proces in the exponent agrees with
// the power of the secret.
func TestCombinePoints(t *testing.T) {
	shamir := shamirSetup()
	secret := shamir.curve.Scalar.Random(rand.Reader)

	shares, err := shamir.Split(secret, rand.Reader)
	if err != nil {
		t.Errorf("failure while creating the shares: %v", err)
	}

	point, err := shamir.CombinePoints(shares[:]...)
	if err != nil {
		t.Errorf("failure during the reconstruction in the exponent: %v", err)
	}

	secretPoint := shamir.curve.ScalarBaseMult(secret)
	if !secretPoint.Equal(point) {
		t.Errorf("the reconstructed point and the secret do not meet: %v", err)
	}
}

// Test that the Split() function should return an error if the secret is zero.
func TestSplitZeroValue(t *testing.T) {
	shamir := shamirSetup()
	secret := shamir.curve.Scalar.Zero()

	_, err := shamir.Split(secret, rand.Reader)
	if err == nil {
		t.Error("the spliting process should return an error if the secret is zero.")
	}
}

// Test that the Combine() function should return an error if there are not
// enough shares.
func TestLessShamirSharesThanThreshold(t *testing.T) {
	shamir := shamirSetup()
	secret := shamir.curve.Scalar.Random(rand.Reader)

	shares, err := shamir.Split(secret, rand.Reader)
	if err != nil {
		t.Errorf("failure while creating the shares: %v", err)
	}

	_, err = shamir.Combine(shares[:shamir.threshold-1]...)
	if err == nil {
		t.Error("the combination of secret should not be possible.")
	}
}

// Test that the shares provided to the Combine are with different x-axis. Here
// we are testing both Combine() and CombinePoints().
func TestRepeatedXAxis(t *testing.T) {
	shamir := shamirSetup()
	secret := shamir.curve.Scalar.Random(rand.Reader)

	shares, err := shamir.Split(secret, rand.Reader)
	if err != nil {
		t.Errorf("failure while creating the shares: %v", err)
	}

	// Get two different indexes
	var i, j int
	for i == j {
		i = mrand.Intn(len(shares))
		j = mrand.Intn(len(shares))
	}

	// We make two shares equal to induce an error
	shares[i] = shares[j]

	_, err = shamir.Combine(shares[:]...)
	if err == nil {
		t.Error("the combination of shares should not work.")
	}

	_, err = shamir.CombinePoints(shares[:]...)
	if err == nil {
		t.Error("the combination of points from shares should not work")
	}
}

// Test if the generated Lagrange coefficients are generated correctly by
// comparing the interpolation using the coefficients and the actual secret.
// Both values should be the same.
func TestLagrangeCoefficients(t *testing.T) {
	shamir := shamirSetup()
	identities := make([]uint32, shamir.limit)
	for i := range identities {
		identities[i] = uint32(i + 1)
	}

	coeffs, err := shamir.LagrangeCoeffs(identities)
	if err != nil {
		t.Errorf("error constructing the Lagrange coefficients: %v", err)
	}

	secret := shamir.curve.Scalar.Random(rand.Reader)

	shares, err := shamir.Split(secret, rand.Reader)
	if err != nil {
		t.Errorf("failure while creating the shares: %v", err)
	}

	linearComb := shamir.curve.Scalar.Zero()
	for xPoint, c := range coeffs {
		shareScalar := shamir.curve.Scalar.Zero()
		shareScalar.SetBytes(shares[xPoint-1].Value)
		prod := c.Mul(shareScalar)
		linearComb = linearComb.Add(prod)
	}

	if linearComb.Cmp(secret) != 0 {
		t.Error("the coefficient don't interpolate the same secret.")
	}
}
