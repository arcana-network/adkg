package acss

import (
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

func TestKZGVerifier_Verify(t *testing.T) {
	curve := curves.K256()
	secret := GenerateSecret(curve)
	var n = uint32(7)
	var k = uint32(3)

	verifier, _, err := GenerateCommitmentAndShares(secret, k, n, curve)
	if err != nil {
		t.Fatal(err)
	}
	compressedCommitments := CompressCommitments(verifier)
	keypair := GenerateKeyPair(curve)
	_, _, verified := Check(keypair.PrivateKey.Bytes(), secret.Bytes(), compressedCommitments, int(k), curve)
	if !verified {
		t.Fatal("Check() does not verify")
	}
}
