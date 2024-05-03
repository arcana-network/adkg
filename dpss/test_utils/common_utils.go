package testutils

import (
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
	"github.com/vivint/infectious"
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

func CreateShardAndHash(
	dealerNode common.PSSParticipant,
	ephemeralKeypairDealer common.KeyPair,
) ([]infectious.Share, []byte, error) {
	// Creates the Reed-Solomon shards for the message.
	n, k, _ := dealerNode.Params()
	secret := sharing.GenerateSecret(curves.K256())
	commitments, shares, err := sharing.GenerateCommitmentAndShares(
		secret,
		uint32(k),
		uint32(n),
		curves.K256(),
	)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	shards, hashMsg, err := ComputeReedSolomonShardsAndHash(
		commitments,
		dealerNode,
		shares,
		ephemeralKeypairDealer,
	)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	return shards, hashMsg, nil
}

// Computes the Reed-Solomon shards and hash of a given commitment and shares.
func ComputeReedSolomonShardsAndHash(
	commitment *sharing.FeldmanVerifier,
	dealer common.PSSParticipant,
	shares []*sharing.ShamirShare,
	dealerEphemeralKey common.KeyPair,
) ([]infectious.Share, []byte, error) {
	n, _, t := dealer.Params()
	compressedCommitments := sharing.CompressCommitments(commitment)
	shareMap := make(map[string][]byte, n)
	for _, share := range shares {
		nodePublicKey := dealer.GetPublicKeyFor(int(share.Id), dealer.IsNewNode())
		if nodePublicKey == nil {
			log.Errorf("Couldn't obtain public key for node with id=%v", share.Id)
			return []infectious.Share{}, []byte{}, errors.New("Public key is nil")
		}

		cipherShare, hmacTag, err := sharing.EncryptSymmetricCalculateKey(
			share.Bytes(),
			nodePublicKey,
			dealerEphemeralKey.PrivateKey,
		)

		if err != nil {
			log.Errorf("Error while encrypting secret share, err=%v", err)
			return []infectious.Share{}, []byte{}, errors.New("Can't been able to encrypt the shares")
		}
		log.Debugf("CIPHER_SHARE=%v, HMAC=%v", cipherShare, hmacTag)

		// combining the encrypted shares and hmac tag
		cipherShare = sharing.Combine(cipherShare, hmacTag)
		pubkeyHex := common.PointToHex(nodePublicKey)
		shareMap[pubkeyHex] = cipherShare
	}

	msgData := common.AcssData{
		Commitments:           compressedCommitments,
		ShareMap:              shareMap,
		DealerEphemeralPubKey: hex.EncodeToString(dealerEphemeralKey.PrivateKey.Bytes()),
	}

	msgBytes, err := bijson.Marshal(msgData)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	msgHash := common.HashByte(msgBytes)

	fec, err := infectious.NewFEC(t+1, n)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	shards, err := acss.Encode(fec, msgBytes)
	if err != nil {
		return []infectious.Share{}, []byte{}, err
	}

	return shards, msgHash, nil
}
