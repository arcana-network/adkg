package dacss

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	testutils "github.com/arcana-network/dkgnode/dpss/test_utils"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Nodes(Old and New both) receives a Propose msg from the dealer
// stores the state
// verifies the shares and commitment
// if everything passes, send an eco msg

func TestProcessProposeMessage(t *testing.T) {

	defaultSetup := testutils.DefaultTestSetup()

	SingleOldNode := defaultSetup.GetSingleOldNodeFromTestSetup()
	transport := SingleOldNode.Transport

	msgOldCommittee := getTestValidProposeMsg(SingleOldNode, defaultSetup, false)

	// Pre-check: in the node's state DualAcssStarted is false
	assert.False(t, SingleOldNode.State().DualAcssStarted)

	// Call the process on the msg
	msgOldCommittee.Process(SingleOldNode.Details(), SingleOldNode)

	sent_msg := transport.GetSentMessages()
	assert.Equal(t, len(sent_msg), defaultSetup.OldCommitteeParams.N)

	singleNewNode := defaultSetup.GetSingleNewNodeFromTestSetup()

	msgNewCommittee := getTestValidProposeMsg(singleNewNode, defaultSetup, true)
	msgNewCommittee.Process(singleNewNode.Details(), singleNewNode)

	sent_msg = transport.GetSentMessages()
	//total length of the transport msg = length of the old msgs + new msgs
	assert.Equal(t, len(sent_msg), defaultSetup.NewCommitteeParams.N+defaultSetup.OldCommitteeParams.N)

}

func getTestValidProposeMsg(SingleNode *testutils.PssTestNode, defaultSetup *testutils.TestSetup, newCommittee bool) AcssProposeMessage {

	id := big.NewInt(1)
	pssRoundDetails := common.PSSRoundDetails{
		PssID:  common.NewPssID(*id),
		Dealer: SingleNode.Details(),
	}
	acssRoundDetails := common.ACSSRoundDetails{
		PSSRoundDetails: pssRoundDetails,
		ACSSCount:       1,
	}

	k := 0
	n := 0

	if newCommittee {
		k = defaultSetup.NewCommitteeParams.K
		n = defaultSetup.NewCommitteeParams.K
	} else {
		k = defaultSetup.OldCommitteeParams.K
		n = defaultSetup.OldCommitteeParams.N
	}

	DealerEphemeralKey := common.GenerateKeyPair(curves.K256())
	testSecret := sharing.GenerateSecret(curves.K256())
	commitments, shares, _ := sharing.GenerateCommitmentAndShares(testSecret, uint32(k), uint32(n), curves.K256())
	compressedCommitments := sharing.CompressCommitments(commitments)
	shareMap := make(map[string][]byte, n)

	for _, share := range shares {

		nodePublicKey := SingleNode.GetPublicKeyFor(int(share.Id), newCommittee)
		if nodePublicKey == nil {
			log.Errorf("Couldn't obtain public key for node with id=%v", share.Id)
		}

		cipherShare, err := sharing.EncryptSymmetricCalculateKey(share.Bytes(), nodePublicKey, DealerEphemeralKey.PrivateKey)
		if err != nil {
			log.Errorf("Error while encrypting secret share, err=%v", err)
		}
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		pubkeyHex := hex.EncodeToString(nodePublicKey.ToAffineCompressed())
		shareMap[pubkeyHex] = cipherShare
	}
	msgData := common.AcssData{
		Commitments:           compressedCommitments,
		ShareMap:              shareMap,
		DealerEphemeralPubKey: hex.EncodeToString(DealerEphemeralKey.PublicKey.ToAffineCompressed()),
	}

	msg := AcssProposeMessage{
		ACSSRoundDetails:   acssRoundDetails,
		Kind:               AcssProposeMessageType,
		CurveName:          common.CurveName(curves.K256().Name),
		Data:               msgData,
		NewCommittee:       newCommittee,
		NewCommitteeParams: defaultSetup.NewCommitteeParams,
	}
	return msg
}
