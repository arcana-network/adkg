package sharing

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	// FIXME this import has to be replaced. Kryptology is deprecated
	"github.com/coinbase/kryptology/pkg/core/curves"
)

// TODO check this impl
// TODO tests

// Combine the keys to create a shared key and encrypt the message (symmetric encryption)
func EncryptSymmetricCalculateKey(msg []byte, public curves.Point, priv curves.Scalar) ([]byte, error) {
	key, err := CalculateSharedKey(public, priv)
	if err != nil || key == nil {
		return nil, err
	}
	return EncryptSymmetric(msg, key)
}

func CalculateSharedKey(public curves.Point, priv curves.Scalar) (curves.Point, error) {
	// TODO check if the public key is valid and possibly throw error

	key := public.Mul(priv)
	return key, nil
}

// Encrypts the message using AES-GCM with the given elliptic curve point as the key
func EncryptSymmetric(msg []byte, key curves.Point) ([]byte, error) {
	// Serialize the elliptic curve point to bytes.
	keyBytes := key.ToAffineCompressed()

	// Hash the serialized point to derive a symmetric key.
	hashedKey := sha256.Sum256(keyBytes)

	// Create a new AES cipher block using the hashed key.
	block, err := aes.NewCipher(hashedKey[:])
	if err != nil {
		return nil, err
	}

	// Create a new GCM cipher mode instance.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate a nonce for AES-GCM.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the message using AES-GCM.
	encrypted := gcm.Seal(nonce, nonce, msg, nil) // Prepend the nonce to the ciphertext.

	return encrypted, nil
}

// Decrypts the message using AES-GCM with the given elliptic curve point as the key
func DecryptSymmetricCalculateKey(encryptedMsg []byte, public curves.Point, priv curves.Scalar) ([]byte, error) {
	key, _ := CalculateSharedKey(public, priv)

	// Serialize the elliptic curve point to bytes (the shared key) and hash it to derive the symmetric key
	keyBytes := key.ToAffineCompressed()
	hashedKey := sha256.Sum256(keyBytes)

	return Decrypt(hashedKey, encryptedMsg)
}

func Decrypt(hashedKey [32]byte, encryptedMsg []byte) ([]byte, error) {
	// Create a new AES cipher block using the hashed key
	block, err := aes.NewCipher(hashedKey[:])
	if err != nil {
		return nil, err
	}

	// Create a new GCM cipher mode instance
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// The nonce size should be the same as was used during encryption
	nonceSize := gcm.NonceSize()
	if len(encryptedMsg) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Extract the nonce and the actual ciphertext from the encrypted message
	nonce, ciphertext := encryptedMsg[:nonceSize], encryptedMsg[nonceSize:]

	// Decrypt the message using AES-GCM
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
