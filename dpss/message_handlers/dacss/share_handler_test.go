package dacss

import (
	"encoding/hex"
	"log"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
	"github.com/torusresearch/bijson"
)

/*
Function: Process

Testcase: happy path. An old node, who is the dealer of this round receives the msg
and proceeds to send shares to both old & new comittee

Expectations:
- node broadcasts 2 msgs: 1 to old comittee, 1 to new comittee
- in the node's state DualAcssStarted is set to true
- the broadcasted msgs contain the correct amount of shares & commitments for the old & new committee parameters respectively
- shares are correctly encrypted, using the ephemeral key
- for both comittee, predicate verifies for the sent information
*/
func TestStartDualAcss(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	oldParams := defaultSetup.OldCommitteeParams
	newParams := defaultSetup.NewCommitteeParams
	testDealer := defaultSetup.GetSingleOldNodeFromTestSetup()
	transport := testDealer.Transport

	ephemeralKeypair := common.GenerateKeyPair(curves.K256())

	// Create a DualCommitteeACSSShareMessage
	msg := getTestMsg(testDealer, defaultSetup, ephemeralKeypair)

	// Pre-check: in the node's state DualAcssStarted is false
	assert.False(t, testDealer.State().DualAcssStarted)

	// Call the process on the msg
	msg.Process(testDealer.Details(), testDealer)

	// CHECKS

	broadcastedMsgs := transport.BroadcastedMessages
	// 1. Check msg was broadcasted twice (once to each committee)
	assert.Equal(t, 2, len(broadcastedMsgs))

	// 2. Check N_old shares & K_old commitments were sent to old committee
	// The first broadcasted msg was for the old committee
	msgContent_old := broadcastedMsgs[0]
	var proposeMsg_old AcssProposeMessage
	err := bijson.Unmarshal(msgContent_old.Data, &proposeMsg_old)
	if err != nil {
		log.Fatalf("Error parsing ProposeMsg JSON: %v", err)
	}

	sentShares_old := proposeMsg_old.Data.ShareMap
	sentCommitments_old := proposeMsg_old.Data.Commitments
	commitments_old, _ := sharing.DecompressCommitments(oldParams.K, sentCommitments_old, curves.K256())

	assert.Equal(t, testutils.DefaultN_old, len(sentShares_old))
	assert.Equal(t, testutils.DefaultK_old, len(commitments_old))

	// 3. Check N_new shares & K_new commitments were sent to the new committee
	msgContent_new := broadcastedMsgs[1]
	var proposeMsg_new AcssProposeMessage
	err = bijson.Unmarshal(msgContent_new.Data, &proposeMsg_new)
	if err != nil {
		log.Fatalf("Error parsing ProposeMsg JSON: %v", err)
	}

	sentShares_new := proposeMsg_new.Data.ShareMap
	sentCommitments_new := proposeMsg_new.Data.Commitments
	commitments_new, _ := sharing.DecompressCommitments(newParams.K, sentCommitments_new, curves.K256())

	assert.Equal(t, testutils.DefaultN_new, len(sentShares_new))
	assert.Equal(t, testutils.DefaultK_new, len(commitments_new))

	// 4. Check: Shares were correctly encrypted for node 2 of old committee
	pubkeyOldNode2 := testDealer.GetPublicKeyFor(2, false)
	share_node2 := sentShares_old[hex.EncodeToString(pubkeyOldNode2.ToAffineCompressed())]
	symm_key2, _ := sharing.CalculateSharedKey(pubkeyOldNode2, ephemeralKeypair.PrivateKey)
	_, _, verified_old := sharing.Predicate(symm_key2, share_node2, sentCommitments_old, defaultSetup.OldCommitteeParams.K, curves.K256())
	assert.True(t, verified_old)

	// 5. Check: Shares were correctly encrypted for node 3 of new committee
	pubkeyNewNode3 := testDealer.GetPublicKeyFor(3, true)
	share_node3 := sentShares_new[hex.EncodeToString(pubkeyNewNode3.ToAffineCompressed())]

	symm_key3, _ := sharing.CalculateSharedKey(pubkeyNewNode3, ephemeralKeypair.PrivateKey)
	_, _, verified_new := sharing.Predicate(symm_key3, share_node3, sentCommitments_new, defaultSetup.NewCommitteeParams.K, curves.K256())
	assert.True(t, verified_new)

	// 6. Check DualAcssStarted is true in the node's state
	assert.True(t, testDealer.State().DualAcssStarted)

}

func getTestMsg(testDealer *testutils.PssTestNode, defaultSetup *testutils.TestSetup, ephemeralKeypair common.KeyPair) DualCommitteeACSSShareMessage {

	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: testDealer.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: pssRoundDetails,
		ACSSCount:       1,
	}
	testSecret := sharing.GenerateSecret(curves.K256())
	msg := DualCommitteeACSSShareMessage{
		ACSSRoundDetails:   acssRoundDetails,
		Kind:               ShareMessageType,
		CurveName:          common.CurveName(curves.K256().Name),
		Secret:             testSecret,
		EphemeralSecretKey: ephemeralKeypair.PrivateKey.Bytes(),
		EphemeralPublicKey: ephemeralKeypair.PublicKey.ToAffineCompressed(),
		Dealer:             testDealer.Details(),
		NewCommitteeParams: defaultSetup.NewCommitteeParams,
	}
	return msg
}

/*
Function: Process

Testcase: the receiving node is in New comittee (while it should be in Old comittee)

Expectations:
- early return. In particular no messages are broadcast
*/
func TestNodeInNewCommittee(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	// Node is NOT in old committee (as it should be)
	nodeNewCommittee := defaultSetup.GetSingleNewNodeFromTestSetup()
	transport := nodeNewCommittee.Transport

	ephemeralKeypair := common.GenerateKeyPair(curves.K256())

	// Create a DualCommitteeACSSShareMessage
	msg := getTestMsg(nodeNewCommittee, defaultSetup, ephemeralKeypair)

	// Call the process on the msg
	msg.Process(nodeNewCommittee.Details(), nodeNewCommittee)

	// CHECKS
	// 1. Check No msg were broadcasted; early return expected
	transport.AssertNoMsgsBroadcast(t)
}

/*
Function: Process

Testcase: the message comes from another node than self (sender neq receiver)

Expectations:
- early return. In particular no messages are broadcast
*/
func TestSenderNotSelf(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	oldNode1, oldNode2 := defaultSetup.GetTwoOldNodesFromTestSetup()
	transport := oldNode1.Transport

	ephemeralKeypair := common.GenerateKeyPair(curves.K256())

	// Create a DualCommitteeACSSShareMessage
	msg := getTestMsg(oldNode1, defaultSetup, ephemeralKeypair)

	// Call the process on the msg
	// The sender is not equal to the "self"(receiver)
	msg.Process(oldNode2.Details(), oldNode1)

	// CHECKS
	// 1. Check No msg were broadcasted; early return expected
	transport.AssertNoMsgsBroadcast(t)
}

/*
Function: Process

Testcase: DualAcssStarted in the node's state is true (meaning the process has already started)

Expectations:
- early return. In particular no messages are broadcast
*/
func TestDualACSSAlreadyStarted(t *testing.T) {
	defaultSetup := testutils.DefaultTestSetup()
	testDealer := defaultSetup.GetSingleOldNodeFromTestSetup()
	transport := testDealer.Transport

	ephemeralKeypair := common.GenerateKeyPair(curves.K256())

	// Create a DualCommitteeACSSShareMessage
	msg := getTestMsg(testDealer, defaultSetup, ephemeralKeypair)

	testDealer.State().DualAcssStarted = true

	// Manually set DualAcssStarted to true
	assert.True(t, testDealer.State().DualAcssStarted)

	// Call the process on the msg
	msg.Process(testDealer.Details(), testDealer)

	// CHECKS
	// 1. Check No msg were broadcasted; early return expected
	transport.AssertNoMsgsBroadcast(t)
}
