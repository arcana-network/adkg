package sharing

import (
	"crypto/rand"
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestEncryption(t *testing.T) {
	curve := curves.ED25519()
	DealerSecretKey := curve.Scalar.Random(rand.Reader)
	DealerPublicKey := curve.NewGeneratorPoint().Mul(DealerSecretKey)

	NodeSecretKey := curve.Scalar.Random(rand.Reader)
	NodePublicKey := curve.NewGeneratorPoint().Mul(NodeSecretKey)

	//Test for honest dealer
	for i := 0; i < 100; i++ {
		plainText, err := generateRandomPlaintext(i*i + i + 1)
		assert.Nil(t, err)

		//encrypting with dealer's private key and node's public key
		cipher, err := EncryptSymmetricCalculateKey(plainText, NodePublicKey, DealerSecretKey)
		assert.Nil(t, err)

		//decryption with dealer's public key and node's private key
		decryptCipher, err := DecryptSymmetricCalculateKey(cipher, DealerPublicKey, NodeSecretKey)
		assert.Nil(t, err)

		assert.Equal(t, plainText, decryptCipher)
	}

	//Test for malicious dealer
	for i := 0; i < 100; i++ {

		//malicious dealer giving random public key and encrypting with different secret key
		DealerPublicKey = curve.Point.Random(rand.Reader)

		plainText, err := generateRandomPlaintext(i*i*i + 1)
		assert.Nil(t, err)

		//encrypting with dealer's private key and node's public key
		cipher, err := EncryptSymmetricCalculateKey(plainText, NodePublicKey, DealerSecretKey)
		assert.Nil(t, err)

		//decryption with dealer's public key and node's private key
		_, err = DecryptSymmetricCalculateKey(cipher, DealerPublicKey, NodeSecretKey)
		assert.NotNil(t, err)
	}

	for i := 0; i < 100; i++ {

		// The dealer encrypting with random publicKey instead of the node's public key
		NodePublicKey = curve.Point.Random(rand.Reader)

		plainText, err := generateRandomPlaintext(i*i + i + 41)
		assert.Nil(t, err)

		//encrypting with dealer's private key and node's public key
		cipher, err := EncryptSymmetricCalculateKey(plainText, NodePublicKey, DealerSecretKey)
		assert.Nil(t, err)

		//decryption with dealer's public key and node's private key
		_, err = DecryptSymmetricCalculateKey(cipher, DealerPublicKey, NodeSecretKey)
		assert.NotNil(t, err)
	}

}

func generateRandomPlaintext(length int) ([]byte, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return []byte{}, err
	}

	return randomBytes, nil
}

func TestPredicate(t *testing.T) {

	//Checked for both K256 and ED25519
	curve := curves.K256()

	//generate random keys for all nodes
	NodesSecretKey := make([]curves.Scalar, 0)
	NodesPublicKey := make([]curves.Point, 0)

	for i := 0; i < 7; i++ {
		nodeSecret := curve.Scalar.Random(rand.Reader)
		NodesSecretKey = append(NodesSecretKey, nodeSecret)
		NodesPublicKey = append(NodesPublicKey, curve.NewGeneratorPoint().Mul(nodeSecret))
	}

	//dealer's Ephemeral key
	DealerSecretKey := curve.Scalar.Random(rand.Reader)
	DealerPublicKey := curve.NewGeneratorPoint().Mul(DealerSecretKey)

	//secret to be shared
	secret := curve.Scalar.Random(rand.Reader)
	commitments, shares, _ := GenerateCommitmentAndShares(secret, 3, 7, curve)

	// Compress commitments
	compressedCommitments := CompressCommitments(commitments)

	// Init share map
	shareMap := make(map[uint32][]byte, 7)

	// encrypt each share with node respective generated symmetric key using Ephemeral Private key and add to share map
	for i, share := range shares {

		if NodesPublicKey[i] == nil {
			log.Errorf("Couldn't obtain public key for node with id=%v", share.Id)
			return
		}

		cipherShare, err := EncryptSymmetricCalculateKey(share.Bytes(), NodesPublicKey[i], DealerSecretKey)

		if err != nil {
			log.Errorf("Error while encrypting secret share, err=%v", err)
			return
		}
		log.Debugf("CIPHER_SHARE=%v", cipherShare)
		shareMap[share.Id] = cipherShare
	}

	//test predicate
	for i := 0; i < 7; i++ {
		shares_node2 := shareMap[uint32(i+1)][:]
		//decrypting with dealer's public Ephemeral Key and node's private key
		symm_key2, _ := CalculateSharedKey(DealerPublicKey, NodesSecretKey[i])
		_, _, verified_old := Predicate(symm_key2, shares_node2, compressedCommitments, 3, curves.K256())
		assert.True(t, verified_old)
	}

}
