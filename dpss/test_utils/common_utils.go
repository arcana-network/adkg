package testutils

import (
	"math/big"

	"github.com/arcana-network/dkgnode/common"
)

// Helpers test functions
func GetTestACSSRoundDetails(dealer common.PSSParticipant) common.ACSSRoundDetails {
	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: dealer.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: pssRoundDetails,
		ACSSCount:       1,
	}
	return acssRoundDetails
}
