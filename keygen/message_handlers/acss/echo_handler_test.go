package acss

import (
	"math/big"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vivint/infectious"
)


/*
Function: Process
Case: Node 0 will send a valid EchoMessage to all nodes
Expects:
  - in receiving node: state.keygen.ReceivedEcho for the sending node is set to True
  - in receiving node: c.EC (echo count) is set to 1
  - NO ready msg is sent (since all nodes have only received 1 echo msg)
*/
func TestReceiveFirstEchoMessage(t *testing.T) {
	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)

	nodes, transport := setupNodes(n, 0)
	node0 := nodes[0]
	node3 := nodes[3]

	// Node 3 is the dealer of this round and is the one that sends shares to all nodes
	// Then, all nodes echo "having received a share"
	// In this test only node0 is going to echo
	round := common.RoundDetails{
		ADKGID: id,
		Dealer: node3.ID(),
		Kind:   "acss",
	}
	test_secret := acss.GenerateSecret(c)

	n, k, _ := node3.Params()

	// Node3 generates commitments and shares for all nodes from test_secret
	commitments, shares, _ := acss.GenerateCommitmentAndShares(test_secret,
		uint32(k), uint32(n), c)
	compressedCommitments := acss.CompressCommitments(commitments)

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

	proposeData, _ := messageData.Serialize()
	fec, _ := infectious.NewFEC(k, n)

	hash := common.HashByte(proposeData)
	encodedShares, _ := acss.Encode(fec, proposeData)

	// We skip the part where node0 "really" receives the share, we immediately send an echo to all other nodes
	for _, nodeDetails := range node0.Nodes() {
		receiverEcho := nodes[nodeDetails.Index-1]
		// Create empty keygen state
		defaultKeygen := &common.SharingStore{
			RoundID: round.ID(),
			State: common.RBCState{
				Phase:         common.Initial,
				ReceivedReady: make(map[int]bool),
				ReceivedEcho:  make(map[int]bool),
			},
			CStore: make(map[string]*common.CStore),
		}

		keygen, _ := receiverEcho.State().KeygenStore.GetOrSetIfNotComplete(round.ID(), defaultKeygen)
		state := receiverEcho.State().KeygenStore
		state_before, _ := state.Get(round.ID())
		// in the default keygen state `ReceivedEcho` for node0 must be false
		assert.False(t, state_before.State.ReceivedEcho[node0.Details().Index])

		msgToSend, _ := NewAcssEchoMessage(round.ID(), encodedShares[nodeDetails.Index-1], hash, common.CurveName(c.Name))
		echoMsg := EchoMessage{
			round.ID(),
			EchoMessageType,
			common.CurveName(c.Name),
			encodedShares[nodeDetails.Index-1],
			hash,
		}
		// receiverNode gets direct Echo message from node0
		receiverEcho.ReceiveMessage(node0.Details(), *msgToSend)

		// Check: receivedEcho for node0 is now set to true
		state_after, _ := state.Get(round.ID())
		assert.True(t, state_after.State.ReceivedEcho[node0.Details().Index])

		cid := echoMsg.Fingerprint()
		c := common.GetCStore(keygen, cid)

		// Check: Echo Count in receiving node is now set to 1
		assert.Equal(t, c.EC, 1)
	}

	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, len(broadcastedMessages), 0, "No `Ready` messages should have been broadcasted")
}

/*
Function: Process
Case: a node receives >= (2*f)+1) Echo's
Expects: the node broadcast a "Ready" message
*/
func TestReadyMsgSentCase1(t *testing.T) {
	// Node0 needs to receive 2*3+1=7 echoes; an echo from each node
	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)

	nodes, transport := setupNodes(n, 0)
	node3 := nodes[3]
	node0 := nodes[0]

	// Node 3 is the dealer of this round and is the one that sends shares to all nodes
	round := common.RoundDetails{
		ADKGID: id,
		Dealer: node3.ID(),
		Kind:   "acss",
	}
	test_secret := acss.GenerateSecret(c)

	n, k, _ := node3.Params()

	// Node3 generates commitments and shares for all nodes from test_secret
	commitments, shares, _ := acss.GenerateCommitmentAndShares(test_secret,
		uint32(k), uint32(n), c)
	compressedCommitments := acss.CompressCommitments(commitments)

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

	proposeData, _ := messageData.Serialize()
	fec, _ := infectious.NewFEC(k, n)

	hash := common.HashByte(proposeData)
	encodedShares, _ := acss.Encode(fec, proposeData)

	// All nodes will echo to node0
	for _, echoingNode := range nodes {
		broadcastedMessages := transport.GetBroadcastedMessages()
		assert.Equal(t, 0, len(broadcastedMessages), "No message should be broadcasted yet")
		msgToSend, _ := NewAcssEchoMessage(round.ID(), encodedShares[node0.id-1], hash, common.CurveName(c.Name))
		node0.ReceiveMessage(echoingNode.Details(), *msgToSend)
	}

	// Give some time for all echoes to be processed.
	time.Sleep(500 * time.Millisecond)

	// After receiving 2*f+1 (=7) echoes, node0 broadcasts a Ready message
	// (none of the other nodes should broadcast)
	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, 1, len(broadcastedMessages), "Node0 should send the Ready message")
}

/*
// TODO
// - if this node has received >= f+1 Ready msg's AND >= f+1 echo's
func TestReadyMsgSentCase2(t *testing.T) {

}

// return early is keygen process is already complete
func TestKeygenAlreadyComplete(t *testing.T) {

}

// return early if echo was already received
func TestEchoAlreadyReceived(t *testing.T) {

}
*/
