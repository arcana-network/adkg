package sharing

import (
	"crypto/rand"
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
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
