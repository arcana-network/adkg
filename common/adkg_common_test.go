package common

import (
	"crypto/rand"
	"math"
	"math/big"
	mrand "math/rand"
	"testing"
)

// generateRandADKGID generates a random ADKGID and returns the index that was
// used during its construction.
func generateRandADKGID() (ADKGID, big.Int, error) {
	index, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return GenerateADKGID(*big.NewInt(0)), *big.NewInt(0), err
	}

	adkgId := GenerateADKGID(*index)
	return adkgId, *index, nil
}

func generateRandRoundDetails() (RoundDetails, error) {
	adkgId, _, err := generateRandADKGID()
	if err != nil {
		return RoundDetails{}, err
	}

	dealer := mrand.Int()
	kind := "acss"

	roundDetails := RoundDetails{
		ADKGID: adkgId,
		Dealer: dealer,
		Kind:   kind,
	}

	return roundDetails, nil
}

// Tests if the GetIndex extracts the correct index that was used in the ADKGID
// generation.
func TestADKGIDExtraction(t *testing.T) {
	adkgId, index, err := generateRandADKGID()
	if err != nil {
		t.Errorf("error generating the random index: %v", err)
	}

	retIndex, err := adkgId.GetIndex()
	if err != nil {
		t.Errorf("error extracting the index: %v", err)
	}

	if retIndex.Cmp(&index) != 0 {
		t.Errorf("the indexes %v and %v are not equal", retIndex, index)
	}
}

// Test
func TestRoundId(t *testing.T) {
	roundDetails, err := generateRandRoundDetails()
	if err != nil {
		t.Errorf("error generating the random RoundDetails: %v", err)
	}

	idRound := roundDetails.ID()

	extractedRoundDetails := new(RoundDetails)
	extractedRoundDetails.FromID(idRound)
}
