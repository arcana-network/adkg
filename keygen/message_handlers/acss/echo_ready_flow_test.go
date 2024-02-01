package acss

import (
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vivint/infectious"
)

// Combined tests sending `Echo` and `Ready` messages.

/*
1. send f+1 Echo messages to node0 (those will come from different nodes, they echo the share for node0 to node0)
2. send f+1 Ready messages to node0 from other nodes (not node0 itself)
3. this should trigger node0 to broadcast a Ready message
*/
func TestSendReadyInReadyHandler(t *testing.T) {
	// SETUP
	id := common.GenerateADKGID(*big.NewInt(int64(1)))

	log.SetLevel(log.InfoLevel)
	nodes, transport := setupNodes(n, 0)
	node0 := nodes[0]
	node3 := nodes[3]

	// Node3 generates commitments and shares for all nodes from test_secret
	round := common.RoundDetails{
		ADKGID: id,
		Dealer: node3.ID(),
		Kind:   "acss",
	}
	test_secret := acss.GenerateSecret(c)

	n, k, _ := node3.Params()

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

	// TEST

	echoMsg, _ := NewAcssEchoMessage(round.ID(), encodedShares[node0.id-1], hash, common.CurveName(c.Name))
	// 1. f+1 other nodes echo node0's share to node0
	for i := 1; i<=f+1; i++ {
		node0.ReceiveMessage(nodes[i].Details(), *echoMsg)
	}

	broadcastedMessages := transport.GetBroadcastedMessages()
	assert.Equal(t, len(broadcastedMessages), 0, "No `Ready` messages should have been broadcasted")

	// 2. f+1 other nodes send a ready message to node0
	for i := 1; i<=f+1; i++ {
		senderNode := nodes[i]
		readyMsg, _ := NewReadyMessage(round.ID(), encodedShares[senderNode.id-1], hash, common.CurveName(c.Name))
		node0.ReceiveMessage(senderNode.Details(), *readyMsg)
	}
	
	// After receiving f+1 Echo messages and f+1 Ready messages, 
	// node0 should broadcast a Ready message
	broadcastedMessages = transport.GetBroadcastedMessages()
	assert.Equal(t, 1, len(broadcastedMessages), "No `Ready` messages should have been broadcasted")

}