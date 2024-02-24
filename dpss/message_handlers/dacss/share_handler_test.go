package dacss

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

/*
Node receives msg from itself

Checks: (if any check fails, triggers early return)
- is this an old node
- is the sender the same as the current node
- dual acss not yet started

Executes:
- Generates a secret
- For both committees:
  - generate commitments
  - generate shares
  - encrypt shares with ephemeral keys per receiver
  - broadcast
*/

// WIP!!

func TestStartDualAcss(t *testing.T) {
	testNode, transport := testutils.GetSingleNode(false, false)

	// Step 1: create a DualCommitteeACSSShareMessage
	roundId := common.NewPSSRoundID(big.Int{})
	testSecret := sharing.GenerateSecret(curves.K256())
	msg := DualCommitteeACSSShareMessage{
		RoundID:          roundId,
		Kind:             ShareMessageType,
		CurveName:        common.CurveName(curves.K256().Name),
		Secret:           testSecret,
		EphemeralKeypair: testNode.Keypair,
		Dealer:           testNode.Details(),
	}

	fmt.Print(msg)

	// step 2: call the process msg
	msg.Process(testNode.Details(), testNode)

	// 3. Check msg was broadcasted
	// TODO improve checks
	assert.True(t, len(transport.GetBroadcastedMessages()) > 0)

}
