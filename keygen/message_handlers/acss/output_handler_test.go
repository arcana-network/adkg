package acss

import (
	"math/big"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/keyset"
	"github.com/arcana-network/dkgnode/keygen/messages"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

/*
Function: Process
Case: a node receives 1 `Output` message from itself with a share that fulfills the predicate
Expects:
- no `keyset.Init` message sent
- share is store in sessionStore.S
- in sessionStore.TPrime bit for dealer of that round is set
- in sessionStore.C commitments are added for that dealer
*/
func TestReceiveFirstOutputMessage(t *testing.T) {
	// Node 3 = dealer round
	_, node0, node3, round, shares, verifier, msgToSend := outputHandlerTestSetup()
	node0.ReceiveMessage(node0.Details(), *msgToSend)

	adkgid, _ := common.ADKGIDFromRoundID(round.ID())
	sessionStore, _ := node0.State().SessionStore.GetOrSetIfNotComplete(adkgid, common.DefaultADKGSession())

	// Checks
	// 1. no additional messages were sent
	assert.True(t, node0.messageCount == 1) // (this is the initial message itself)
	// 2. share is store in sessionStore.S
	// nodeIds start from 1, therefore we need to subtract 1 to get the right index for shares[uint32(node0.id)-1]
	assert.Equal(t, shares[uint32(node0.id)-1], sessionStore.S[int(node3.id)])
	// 3. in sessionStore.TPrime bit for dealer of that round is set
	assert.True(t, kcommon.HasBit(sessionStore.TPrime, int(node3.id)))
	// 4. in sessionStore.C commitments are added for that dealer
	assert.True(t, pointsEqual(verifier.Commitments, sessionStore.C[int(node3.id)]))
}

/*
Function: Process
Case: a node receives invalid share
Expects: early return
- no `keyset.Init` message sent
- nothing stored (shares & commitments)
- bit for dealer of round is not set
*/
func TestInvalidShare(t *testing.T) {
	// Node 3 = dealer round
	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)
	nodes, _ := setupNodes(n, 0)
	node0 := nodes[0]
	node3 := nodes[3]

	round := common.RoundDetails{
		ADKGID: id,
		Dealer: node3.ID(),
		Kind:   "acss",
	}
	test_secret := acss.GenerateSecret(c)

	n, k, _ := node3.Params()

	verifier, shares, _ := acss.GenerateCommitmentAndShares(test_secret,
		uint32(k), uint32(n), c)
	compressedCommitments := acss.CompressCommitments(verifier)

	shareMap := make(map[uint32][]byte, n)
	for _, share := range shares {
		// Instead of storing the actual shares, we store zeroes
		shareMap[share.Id] = make([]byte, 81)
	}

	messageData := &messages.MessageData{
		Commitments: compressedCommitments,
		ShareMap:    shareMap,
	}

	data, _ := messageData.Serialize()
	msgToSend, _ := NewOutputMessage(round.ID(), data, common.CurveName(c.Name))

	node0.ReceiveMessage(node0.Details(), *msgToSend)

	adkgid, _ := common.ADKGIDFromRoundID(round.ID())
	sessionStore, _ := node0.State().SessionStore.GetOrSetIfNotComplete(adkgid, common.DefaultADKGSession())

	// Checks
	// 1. no additional messages were sent
	assert.True(t, node0.messageCount == 1) // (this is the initial message itself)
	// 2. no shares are stored in sessionStore.S
	assert.True(t, len(sessionStore.S) == 0)
	// 3. bit for dealer of that round is not set in sessionStore.TPrime
	assert.False(t, kcommon.HasBit(sessionStore.TPrime, int(node3.id)))
	// 4. no commitments are stores in sessionStore.C
	assert.True(t, len(sessionStore.C) == 0)
}

/*
Function: Process
Case: 'Output` message received from other node
Expects: early return. 
- no `keyset.Init` message sent
- nothing stored (shares & commitments)
- bit for dealer of round is not set
*/
func TestSenderNotSelf(t *testing.T) {
	// Node 3 = dealer round
	nodes, node0, node3, round, _, _, msgToSend := outputHandlerTestSetup()

	node0.ReceiveMessage(nodes[1].Details(), *msgToSend)

	adkgid, _ := common.ADKGIDFromRoundID(round.ID())
	sessionStore, _ := node0.State().SessionStore.GetOrSetIfNotComplete(adkgid, common.DefaultADKGSession())

	// Checks
	// 1. no additional messages were sent
	assert.True(t, node0.messageCount == 1) // (this is the initial message itself)
	// 2. no shares are stored in sessionStore.S
	assert.True(t, len(sessionStore.S) == 0)
	// 3. bit for dealer of that round is not set in sessionStore.TPrime
	assert.False(t, kcommon.HasBit(sessionStore.TPrime, int(node3.id)))
	// 4. no commitments are stores in sessionStore.C
	assert.True(t, len(sessionStore.C) == 0)
}

/*
Function: Process
Case: a node receives the (f+1)-th `Output` message from itself with a share that fulfills the predicate
Expects:
- 1 `keyset.Init` message sent and it contains the Int value of the TPrime
- sessionStore.TPrime must have been stored in sessionStore.T[nodeId]
*/
func TestEnoughSharesReady(t *testing.T) {
	// Node 3 = dealer round
	_, node0, node3, round, _, _, msgToSend := outputHandlerTestSetup()

	adkgid, _ := common.ADKGIDFromRoundID(round.ID())
	sessionStore, _ := node0.State().SessionStore.GetOrSetIfNotComplete(adkgid, common.DefaultADKGSession())
	// by setting f(=2) bits, we simulate f shares have been set
	// the next share should trigger keyset init message
	testTPrime := (1 << f) - 1
	sessionStore.TPrime = testTPrime
	// update this value the way it will be updated during the `Process` message
	testTPrime = kcommon.SetBit(sessionStore.TPrime, node3.id)
	expectedOutput := kcommon.IntToByteValue(testTPrime)
	newRoundId := common.CreateRound(adkgid, node0.ID(), "keyset")
	expectedInitMsg, _ := keyset.NewInitMessage(newRoundId, expectedOutput, common.CurveName(c.Name))

	node0.ReceiveMessage(node0.Details(), *msgToSend)
	time.Sleep(500 * time.Millisecond)

	// Checks
	// 1. 1 InitMessage must have been sent & received by node0
	sentMsgs := node0.GetReceivedMessages(keyset.InitMessageType)
	assert.Equal(t, 1, len(sentMsgs))
	// 2. check InitMsg is as expected
	assert.Equal(t, sentMsgs[0], *expectedInitMsg)
	// 3. sessionStore.T[nodeId] should contain expected TPrime value
	assert.Equal(t, sessionStore.T[node0.id], testTPrime) 
}

/*
Function: Process
Case: receives a double share;
a node receives the (f+1)-th `Output` message from itself with a share that fulfills the predicate, 
but nothing should happen since the sent share was received already
Expects:
- no `keyset.Init` message sent
*/
func TestReceivesDoubleShare(t *testing.T) {
	// Node 3 = dealer round
	_, node0, node3, round, _, _, msgToSend := outputHandlerTestSetup()

	adkgid, _ := common.ADKGIDFromRoundID(round.ID())
	sessionStore, _ := node0.State().SessionStore.GetOrSetIfNotComplete(adkgid, common.DefaultADKGSession())
	// f(=2) received shares, from round of node0 and round of node3
	testTPrime := 0
	testTPrime = kcommon.SetBit(sessionStore.TPrime, node3.id)
	testTPrime = kcommon.SetBit(sessionStore.TPrime, node0.id)
	sessionStore.TPrime = testTPrime

	node0.ReceiveMessage(node0.Details(), *msgToSend)
	time.Sleep(500 * time.Millisecond)

	// Checks
	// 1. 1 InitMessage must have been sent & received by node0
	sentMsgs := node0.GetReceivedMessages(keyset.InitMessageType)
	assert.Equal(t, 0, len(sentMsgs))
}

// FIXME this test fails, as the check is not for keygen, but for session in `output_handler`

/*
Function: Process
Case: keygen already completed
Expects: early return. 
- no `keyset.Init` message sent
- nothing stored (shares & commitments)
- bit for dealer of round is not set
*/
// func TestKeygenAlreadyCompleted(t *testing.T) {
// 	// Node 3 = dealer round
// 	_, node0, node3, round, _, _, msgToSend := outputHandlerTestSetup()
	
// 	// Set keygen phase to complete for node0 before sending message
// 	node0.State().KeygenStore.Complete(round.ID())

// 	node0.ReceiveMessage(node0.Details(), *msgToSend)

// 	adkgid, _ := common.ADKGIDFromRoundID(round.ID())
// 	sessionStore, _ := node0.State().SessionStore.GetOrSetIfNotComplete(adkgid, common.DefaultADKGSession())
	
// 	// Checks
// 	// 1. no additional messages were sent
// 	assert.True(t, node0.messageCount == 1) // (this is the initial message itself)
// 	// 2. no shares are stored in sessionStore.S
// 	assert.True(t, len(sessionStore.S) == 0)
// 	// 3. bit for dealer of that round is not set in sessionStore.TPrime
// 	assert.False(t, kcommon.HasBit(sessionStore.TPrime, int(node3.id)))
// 	// 4. no commitments are stores in sessionStore.C
// 	assert.True(t, len(sessionStore.C) == 0)
// }


/*
Pending testcase
- set of currently terminated ACSS processes can either be or not be a subset of an existing proposal
*/


func outputHandlerTestSetup() ([]*Node, *Node, *Node, common.RoundDetails, []sharing.ShamirShare, *sharing.FeldmanVerifier, *common.DKGMessage) {
	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)
	nodes, _ := setupNodes(n, 0)
	node0 := nodes[0]
	node3 := nodes[3]

	round := common.RoundDetails{
		ADKGID: id,
		Dealer: node3.ID(),
		Kind:   "acss",
	}
	test_secret := acss.GenerateSecret(c)

	n, k, _ := node3.Params()

	verifier, shares, _ := acss.GenerateCommitmentAndShares(test_secret,
		uint32(k), uint32(n), c)
	compressedCommitments := acss.CompressCommitments(verifier)

	shareMap := make(map[uint32][]byte, n)
	for _, share := range shares {
		nodePublicKey := node3.PublicKey(int(share.Id))

		cipherShare, _ := acss.Encrypt(share.Bytes(), nodePublicKey,
			node3.PrivateKey())

		shareMap[share.Id] = cipherShare
	}

	messageData := &messages.MessageData{
		Commitments: compressedCommitments,
		ShareMap:    shareMap,
	}

	data, _ := messageData.Serialize()
	msgToSend, _ := NewOutputMessage(round.ID(), data, common.CurveName(c.Name))
	return nodes, node0, node3, round, shares, verifier, msgToSend
}

func pointsEqual(a, b []curves.Point) bool {
	if len(a) != len(b) {
			return false
	}
	for i := range a {
			if !a[i].Equal(b[i]) {
					return false
			}
	}
	return true
}