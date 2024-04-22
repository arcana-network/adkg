package sharing

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"

	// FIXME this import has to be replaced. Kryptology is deprecated
	"github.com/coinbase/kryptology/pkg/core/curves"
)

// TODO check this impl
// TODO tests

// Combine the keys to create a shared key and encrypt the message (symmetric encryption)
// Also returns the hmac of the msg
func EncryptSymmetricCalculateKey(msg []byte, public curves.Point, priv curves.Scalar) ([]byte, []byte, error) {
	key, err := CalculateSharedKey(public, priv)
	if err != nil || key == nil {
		return nil, nil, err
	}
	cipher, err := EncryptSymmetric(msg, key)

	if err != nil {
		return nil, nil, err
	}
	finalMAC, err := GetHmacTag(cipher, key.ToAffineCompressed())

	if err != nil {
		return nil, nil, err
	}
	return cipher, finalMAC, nil
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

// generates hmac
// takes msg bytes and hash bytes of the symmetric key
func GetHmacTag(msg, key []byte) ([]byte, error) {

	//create hmac
	// hamc using sha256 and hash of the symmetric key
	mac := hmac.New(sha256.New, sha256.New().Sum(key))

	if mac.Size() != sha256.New().Size() {
		return nil, errors.New("size not equal")
	}

	if mac.BlockSize() != sha256.New().BlockSize() {
		return nil, errors.New("BlockSize not equal")
	}

	mac.Write(msg)
	finalMAC := mac.Sum(nil)

	return finalMAC, nil
}

// The combine and Extract function is needed to combine and extract the encrypted shares and the hmacTag

// combines two byte array of arbitrary length
func Combine(arr1, arr2 []byte) []byte {

	// Create buffer to store the combined arrays
	var buf bytes.Buffer

	// Write the length of the first array as a 4-byte integer
	binary.Write(&buf, binary.LittleEndian, uint32(len(arr1)))

	// Write the first array
	buf.Write(arr1)

	// Write the length of the second array as a 4-byte integer
	binary.Write(&buf, binary.LittleEndian, uint32(len(arr2)))

	// Write the second array
	buf.Write(arr2)

	// Return the combined array
	return buf.Bytes()
}

// extracts the byte arrays from the combined array
func Extract(combined []byte) ([]byte, []byte) {
	// Create reader from combined array
	reader := bytes.NewReader(combined)

	// Read the length of the first array
	var len1 uint32
	binary.Read(reader, binary.LittleEndian, &len1)

	// Read the first array
	arr1 := make([]byte, len1)
	reader.Read(arr1)

	// Read the length of the second array
	var len2 uint32
	binary.Read(reader, binary.LittleEndian, &len2)

	// Read the second array
	arr2 := make([]byte, len2)
	reader.Read(arr2)

	// Return the extracted arrays
	return arr1, arr2
}
