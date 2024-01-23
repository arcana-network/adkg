package acss

import (
	"math/big"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

/*
Function: Process
Case: happy path; leader of round is sender & shares and commitments check out
Expects: Echo msgs are sent to all nodes
*/
func TestProcessProposeMessage(t *testing.T) {
	// Testcase: Node0 sends shares to node1

	// Setup:
	// - Node0 is leader of the round
	// - shareMap contains all the -encrypted- shares
	// - compressedCommitments is also needed for the messageData
	transport, node0, node1, round, shares, compressedCommitments, shareMap := processTestSetup()

	// Create message data for ProposeMsg
	msg := createProposeMessage(shares, node0, shareMap, compressedCommitments, round)

	// Node0 sends proposeMessage to node1, with all the encrypted shares and commitments
	node1.ReceiveMessage(node0.Details(), *msg)
	// Add a small pause so all messages can be sent and received
	time.Sleep(1 * time.Second)

	echoMessagesSent := countSentEchoMessages(transport)

	// Verify that the node sent an echo msg to all other nodes
	assert.Equal(t, echoMessagesSent, n, "This node should have sent 7 EchoMsgs")
}

/*
Function: Process
Case: leader of round not equal to sender of msg
Expects: early return (no messages are sent)
*/
func TestProcessSenderNotLeader(t *testing.T) {
	transport, node0, node1, round, shares, compressedCommitments, shareMap := processTestSetup()

	// Create message data for ProposeMsg
	msg := createProposeMessage(shares, node0, shareMap, compressedCommitments, round)

	// Node0 is the round leader, but Node1 is the sender
	node1.ReceiveMessage(node1.Details(), *msg)
	// Add a small pause so all messages can be sent and received
	time.Sleep(1 * time.Second)

	echoMessagesSent := countSentEchoMessages(transport)

	// Check for early return
	assert.Equal(t, echoMessagesSent, 0, "Nothing should have been done")
}

/*
Function: Process
Case: predicate fails because shares are encrypted with wrong key
Expects: early return (no messages are sent)
*/
func TestProcessInvalidPredicate(t *testing.T) {
	transport, node0, node1, round, shares, compressedCommitments, shareMap := processTestSetup()

	wrongKeyPair := acss.GenerateKeyPair(c)
	// encrypt each share with node the WRONG key, add to share map
	for _, share := range shares {

		cipherShare, err := acss.Encrypt(share.Bytes(), wrongKeyPair.PublicKey,
			wrongKeyPair.PrivateKey)
		if err != nil {
			log.Errorf("acss.Encrypt():err=%v", err)
			return
		}
		shareMap[share.Id] = cipherShare
	}

	// Create message data for ProposeMsg
	messageData := &messages.MessageData{
		Commitments: compressedCommitments,
		ShareMap:    shareMap,
	}

	data, _ := messageData.Serialize()

	msg, _ := NewAcssProposeMessage(
		round.ID(),
		data,
		common.SECP256K1,
	)

	// Node0 sends proposeMessage to node1, with all the encrypted shares and commitments
	node1.ReceiveMessage(node0.Details(), *msg)
	// Add a small pause so all messages can be sent and received
	time.Sleep(1 * time.Second)

	echoMessagesSent := countSentEchoMessages(transport)

	// Verify that no msgs have been sent, since the predicate won't verify
	assert.Equal(t, echoMessagesSent, 0, "No msgs should have been sent")
}

/*
Function: Process
Case: getting leader from RoundID fails
Expects: early return (no messages are sent)
*/
func TestProcessInvalidRoundLeader(t *testing.T) {
	transport, node0, node1, _, shares, compressedCommitments, shareMap := processTestSetup()

	// Create message data for ProposeMsg
	for _, share := range shares {
		nodePublicKey := node0.PublicKey(int(share.Id))

		cipherShare, _ := acss.Encrypt(share.Bytes(), nodePublicKey,
			node0.PrivateKey())

		shareMap[share.Id] = cipherShare
	}

	messageData := &messages.MessageData{
		Commitments: compressedCommitments,
		ShareMap:    shareMap,
	}

	data, _ := messageData.Serialize()

	msg, _ := NewAcssProposeMessage(
		"wrong Round ID", // This is the invalid round ID we're sending to trigger early return
		data,
		common.SECP256K1,
	)
	// Node0 sends proposeMessage to node1, with all the encrypted shares and commitments
	node1.ReceiveMessage(node0.Details(), *msg)
	// Add a small pause so all messages can be sent and received
	time.Sleep(1 * time.Second)

	echoMessagesSent := countSentEchoMessages(transport)

	// Check for early return
	assert.Equal(t, 0, echoMessagesSent, "Cannot get leader from RoundID; no messages should be sent")
}

/*
Function: Process
Case: invalid Pubkey of round leader (sender of message)
Expects: early return (no messages are sent)
*/
func TestProcessInvalidDealerPubkey(t *testing.T) {
	transport, node0, node1, round, shares, compressedCommitments, shareMap := processTestSetup()

	// Create message data for ProposeMsg
	msg := createProposeMessage(shares, node0, shareMap, compressedCommitments, round)

	// Invalid pubkey of dealer
	x := big.NewInt(12345)
	y := big.NewInt(67890)
	invalidPubkey := common.Point{X: *x, Y: *y}
	node1.ReceiveMessage(common.KeygenNodeDetails{Index: node0.id, PubKey: invalidPubkey}, *msg)
	// Add a small pause so all messages can be sent and received
	time.Sleep(1 * time.Second)

	echoMessagesSent := countSentEchoMessages(transport)

	// Check for early return
	assert.Equal(t, echoMessagesSent, 0, "Round leader has invalid pubkey; no messages should be sent")
}

/*
Function: Process
Case: invalid message data in ProposeMessage
Expects: early return (no messages are sent)
*/
func TestProcessInvalidMessageData(t *testing.T) {

	transport, node0, node1, round, _, _, _ := processTestSetup()

	// Create INVALID message data for ProposeMsg
	msg, _ := NewAcssProposeMessage(
		round.ID(),
		[]byte{1},
		common.SECP256K1,
	)

	// Node0 sends proposeMessage to node1, with all the encrypted shares and commitments
	node1.ReceiveMessage(node0.Details(), *msg)
	// Add a small pause so all messages can be sent and received
	time.Sleep(1 * time.Second)

	echoMessagesSent := countSentEchoMessages(transport)

	assert.Equal(t, echoMessagesSent, 0, "Message data couldn't be deserialized; no messages should be sent")
}

/*
Function: Process
Case: invalid parameter values to create FEC
Expects: early return (no messages are sent)
*/
func TestProcessFECError(t *testing.T) {

	transport, node0, node1, round, shares, compressedCommitments, shareMap := processTestSetup()

	// Create message data for ProposeMsg
	msg := createProposeMessage(shares, node0, shareMap, compressedCommitments, round)
	// Adjust parameter n to 0
	node1.AdjustParamN(0)
	node1.ReceiveMessage(node0.Details(), *msg)
	// Add a small pause so all messages can be sent and received
	time.Sleep(1 * time.Second)

	echoMessagesSent := countSentEchoMessages(transport)

	assert.Equal(t, echoMessagesSent, 0, "Invalid parameters; no messages should be sent")
}

// HELPER FUNCTIONS

// returns standard setup for testing; n=7 nodes, no malicious actors, node0 is leader for the current round
func processTestSetup() (*MockTransport, *Node, *Node, common.RoundDetails, []sharing.ShamirShare, []byte, map[uint32][]byte) {
	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)

	nodes, transport := setupNodes(n, 0)

	node0 := nodes[0]
	node1 := nodes[1]

	round := common.RoundDetails{
		ADKGID: id,
		Dealer: node0.ID(),
		Kind:   "acss", // Still in the ACSS part of the process
	}

	// Generate test secret
	test_secret := acss.GenerateSecret(c)

	/*
		n -> total number of nodes
		t = f -> number of max malicious nodes
		k = f + 1 > reconstruction threshold
	*/
	n, k, _ := node0.Params()

	// Generate commitments and shares for all nodes from test_secret
	commitments, shares, _ := acss.GenerateCommitmentAndShares(test_secret,
		uint32(k), uint32(n), c)
	compressedCommitments := acss.CompressCommitments(commitments)

	shareMap := make(map[uint32][]byte, n)
	return transport, node0, node1, round, shares, compressedCommitments, shareMap
}

// returns a ProposeMessage with the commitments and sharesMap
func createProposeMessage(shares []sharing.ShamirShare, node0 *Node, shareMap map[uint32][]byte, compressedCommitments []byte, round common.RoundDetails) *common.DKGMessage {
	// encrypt each share with node respective generated symmetric key, add to share map
	for _, share := range shares {
		nodePublicKey := node0.PublicKey(int(share.Id))

		cipherShare, _ := acss.Encrypt(share.Bytes(), nodePublicKey,
			node0.PrivateKey())

		shareMap[share.Id] = cipherShare
	}

	messageData := &messages.MessageData{
		Commitments: compressedCommitments,
		ShareMap:    shareMap,
	}

	data, _ := messageData.Serialize()

	msg, _ := NewAcssProposeMessage(
		round.ID(),
		data,
		common.SECP256K1,
	)
	return msg
}

// returns the amount of EchoMessages that have been sent with the "Send" function over the MockTransport
func countSentEchoMessages(transport *MockTransport) int {
	sentMessages := transport.GetSentMessages()
	filteredMessages := make([]common.DKGMessage, 0)

	for _, msg := range sentMessages {
		if msg.Method == EchoMessageType {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return len(filteredMessages)
}
