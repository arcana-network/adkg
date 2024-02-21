package acss

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	tronCrypto "github.com/TRON-US/go-eccrypto"
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/coinbase/kryptology/pkg/core/curves"
	kryptsharing "github.com/coinbase/kryptology/pkg/sharing"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

func EncodeEncrypted(ciphertext string, metadata *tronCrypto.EciesMetadata) ([]byte, error) {
	cipher, err := hex.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	iv, err := hex.DecodeString(metadata.Iv)
	if err != nil {
		return nil, err
	}
	mac, err := hex.DecodeString(metadata.Mac)
	if err != nil {
		return nil, err
	}

	pk, err := hex.DecodeString(metadata.EphemPublicKey)
	if err != nil {
		return nil, err
	}
	cpk, err := tronCrypto.NewPublicKeyFromBytes(pk)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 0)
	b = append(b, iv...)              // 16 bytes
	b = append(b, cpk.Bytes(true)...) // 33 bytes
	b = append(b, mac...)             // 32 bytes
	b = append(b, cipher...)          // variable bytes

	return b, nil
}

func DecodeEncrypted(b []byte) (string, *tronCrypto.EciesMetadata, error) {
	metadata := tronCrypto.EciesMetadata{
		Mode: tronCrypto.ENCRYPTION_MODE_1,
	}

	iv := b[:16]     // 0-15 bytes
	pk := b[16:49]   // 16-48 bytes
	mac := b[49:81]  // 49-80 bytes
	cipher := b[81:] // 81- bytes

	pubKey, err := tronCrypto.NewPublicKeyFromBytes(pk)
	if err != nil {
		return "", nil, err
	}

	upk := hex.EncodeToString(pubKey.Bytes(false))

	metadata.EphemPublicKey = upk
	metadata.Iv = hex.EncodeToString(iv)
	metadata.Mac = hex.EncodeToString(mac)

	return hex.EncodeToString(cipher), &metadata, nil
}

func Encrypt(share []byte, public curves.Point, priv curves.Scalar) ([]byte, error) {
	pkHex := hex.EncodeToString(public.ToAffineUncompressed())
	cipher, meta, err := tronCrypto.Encrypt(pkHex, share)
	if err != nil {
		return nil, err
	}
	return EncodeEncrypted(cipher, meta)
}

func Decrypt(privateHex string, cipher []byte) ([]byte, error) {
	c, meta, err := DecodeEncrypted(cipher)
	if err != nil {
		return nil, err
	}
	plaintext, err := tronCrypto.Decrypt(privateHex, c, meta)
	if err != nil {
		return nil, err
	}
	return []byte(plaintext), nil
}

func CompressCommitments(v *sharing.FeldmanVerifier) []byte {
	c := make([]byte, 0)
	for _, v := range v.Commitments {
		e := v.ToAffineCompressed() // 33 bytes
		c = append(c, e[:]...)
	}
	return c
}

func DecompressCommitments(k int, c []byte, curve *curves.Curve) ([]curves.Point, error) {
	commitment := make([]curves.Point, 0)
	for i := 0; i < k; i++ {
		length := 33
		if curve.Name == "ed25519" {
			length = 32
		}
		cI, err := curve.Point.FromAffineCompressed(c[i*length : (i*length)+length])
		if err == nil {
			commitment = append(commitment, cI)
		} else {
			return nil, err
		}
	}

	return commitment, nil
}

// for hbACSS batch commitments
// TODO: Testing not done
func BatchDecompressCommitments(k, B int, c []byte, curve *curves.Curve) ([][]curves.Point, error) {

	if len(c)%B != 0 {
		return nil, fmt.Errorf("not valid Batch compression")
	}

	Batchcommitment := make([][]curves.Point, 0)
	size := len(c) / B
	for i := 0; i < B; i++ {
		points, err := DecompressCommitments(k, c[i*size:(i*size)+size], curve)

		if err != nil {
			return nil, err
		}
		Batchcommitment = append(Batchcommitment, points)
	}

	return Batchcommitment, nil
}

func verifierFromCommits(k int, c []byte, curve *curves.Curve) (*sharing.FeldmanVerifier, error) {

	commitment, err := DecompressCommitments(k, c, curve)
	if err != nil {
		return nil, err
	}
	verifier := new(sharing.FeldmanVerifier)
	verifier.Commitments = commitment
	return verifier, nil
}

func GenerateKeyPair(curve *curves.Curve) common.KeyPair {
	g := curve.NewGeneratorPoint()
	privateKey := curve.NewScalar().Random(rand.Reader)
	publicKey := g.Mul(privateKey)
	return common.KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
}
func GenerateSecret(c *curves.Curve) curves.Scalar {
	secret := c.Scalar.Random(rand.Reader)
	return secret
}

func GenerateCommitmentAndShares(s curves.Scalar, k, n uint32, curve *curves.Curve) (*sharing.FeldmanVerifier, []sharing.ShamirShare, error) {
	f, err := sharing.NewFeldman(k, n, curve)
	if err != nil {
		return nil, nil, fmt.Errorf("gen_commitment_and_shares: %w", err)
	}

	feldcommit, shares, err := Split(s, f.Threshold, f.Limit, f.Curve, rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("gen_commitment_and_shares: %w", err)
	}
	return feldcommit, shares, nil
}

func Split(secret curves.Scalar, threshold, limit uint32, curve *curves.Curve, reader io.Reader) (*sharing.FeldmanVerifier, []sharing.ShamirShare, error) {

	shares, poly := getPolyAndShares(secret, threshold, limit, curve, reader)
	verifier := new(sharing.FeldmanVerifier)
	verifier.Commitments = make([]curves.Point, threshold)
	for i := range verifier.Commitments {
		base, _ := sharing.CurveParams(curve.Name)
		verifier.Commitments[i] = base.Mul(poly.Coefficients[i])
	}
	return verifier, shares, nil
}

func getPolyAndShares(
	secret curves.Scalar,
	threshold, limit uint32,
	curve *curves.Curve,
	reader io.Reader) ([]sharing.ShamirShare, *kryptsharing.Polynomial) {
	poly := new(kryptsharing.Polynomial).Init(secret, threshold, reader)
	shares := make([]sharing.ShamirShare, limit)
	for i := range shares {
		x := curve.Scalar.New(i + 1)
		shares[i] = sharing.ShamirShare{
			Id:    uint32(i + 1),
			Value: poly.Evaluate(x).Bytes(),
		}
	}
	return shares, poly
}

func SharedKey(priv curves.Scalar, dealerPublicKey curves.Point) [32]byte {
	key := dealerPublicKey.Mul(priv)
	keyHash := sha256.Sum256(key.ToAffineCompressed())
	return keyHash
}

// Predicate verifies if the share fits the polynomial commitments
func Predicate(key []byte, cipher []byte, commits []byte, k int, curve *curves.Curve) (*sharing.ShamirShare, *sharing.FeldmanVerifier, bool) {
	shareBytes, err := Decrypt(hex.EncodeToString(key), cipher)
	if err != nil {
		log.Errorf("Error while decrypting share: err=%s", err)
		return nil, nil, false
	}
	share := sharing.ShamirShare{Id: binary.BigEndian.Uint32(shareBytes[:4]), Value: shareBytes[4:]}
	log.Debugf("share: id=%d, val=%v", share.Id, share.Value)
	verifier, err := verifierFromCommits(k, commits, curve)
	if err != nil {
		log.Errorf("Error while getting verifier from commits=%s", err)
		return nil, nil, false
	}

	if err = verifier.Verify(&share); err != nil {
		log.Errorf("Error while verifying share=%s", err)
		return nil, nil, false
	}
	return &share, verifier, true
}

func Encode(encoder *infectious.FEC, msg []byte) ([]infectious.Share, error) {
	shares := make([]infectious.Share, encoder.Total())
	output := func(s infectious.Share) {
		shares[s.Number] = s.DeepCopy()
	}

	paddedMsg, err := pkcs7Pad(msg, encoder.Required())
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(paddedMsg, output)
	if err != nil {
		return nil, err
	}

	return shares, nil
}

func Decode(f *infectious.FEC, s []infectious.Share) ([]byte, error) {
	result, err := f.Decode(nil, s)
	if err != nil {
		return nil, err
	}

	unpaddedMsg, err := pkcs7Unpad(result, f.Required())
	return unpaddedMsg, err
}

var (
	// ErrInvalidBlockSize indicates hash blocksize <= 0.
	ErrInvalidBlockSize = errors.New("invalid blocksize")

	// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad.
	ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")

	// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails to bad input.
	ErrInvalidPKCS7Padding = errors.New("invalid padding on input")
)

func pkcs7Pad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

func pkcs7Unpad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	if len(b)%blocksize != 0 {
		return nil, ErrInvalidPKCS7Padding
	}
	c := b[len(b)-1]
	n := int(c)
	if n == 0 || n > len(b) {
		return nil, ErrInvalidPKCS7Padding
	}
	for i := 0; i < n; i++ {
		if b[len(b)-n+i] != c {
			return nil, ErrInvalidPKCS7Padding
		}
	}
	return b[:len(b)-n], nil
}
