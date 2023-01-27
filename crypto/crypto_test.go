package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/arcana-network/dkgnode/common"
)

func TestHexToSig(t *testing.T) {
	sig := HexToSig("ROSGVVfIaH5n6INxwgT82mnCaozLJaPzORj7rYzNNLS8tfV3G560i71sIVjefbP4Lzi2RrijpakhLlbrpMaNv75wCzD3LOVWqVEPgEZ8ofXcAaBhtDmI9G0x8qBFWLFN3Ov")
	if string(sig.R[:]) != "ROSGVVfIaH5n6INxwgT82mnCaozLJaPz" {
		t.Fatal("Should be able to generate a signature")
	}
	if string(sig.S[:]) != "Lzi2RrijpakhLlbrpMaNv75wCzD3LOVW" {
		t.Fatal("Should be able to generate a signature")
	}
}

func TestSigToHex(t *testing.T) {
	ecdsaSig := Signature{}
	copy(ecdsaSig.R[:], []byte("ROSGVVfIaH5n6INxwgT82mnCaozLJaPz")[:32])
	copy(ecdsaSig.S[:], []byte("Lzi2RrijpakhLlbrpMaNv75wCzD3LOVW")[:32])
	hex := SigToHex(ecdsaSig)
	if hex != "524f5347565666496148356e36494e7877675438326d6e43616f7a4c4a61507a4c7a69325272696a70616b684c6c6272704d614e76373577437a44334c4f5657" {
		t.Fatal("Should be able to convert a signature into a hex")
	}
}

func TestSignData(t *testing.T) {
	privateKeyHex := "6a6594b9673a8c8987113327a93bf4a216d6cbe53e5f31bbaa8c5228a1664591"

	privateKeyECDSA, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		t.Fatal(err)
	}
	signature := SignData([]byte("ROSGVVfIaH5n6INxwgT82mnCaozLJaPzORj7rYzNNLS8tfV3G560i71sIVjefbP4Lzi2RrijpakhLlbrpMaNv75wCzD3LOVWqVEPgEZ8ofXcAaBhtDmI9G0x8qBFWLFN3Ov"), privateKeyECDSA)
	if bytes.Compare(signature.Raw, []byte{38, 13, 234, 31, 255, 134, 106, 82, 9, 210, 194, 0, 254, 182, 177, 131, 175, 134, 242, 0, 172, 48, 222, 171, 195, 49, 104, 26, 179, 109, 153, 26, 93, 200, 52, 88, 212, 65, 21, 176, 144, 12, 234, 83, 255, 228, 160, 84, 155, 243, 228, 222, 146, 202, 38, 51, 123, 249, 185, 4, 246, 158, 41, 155, 1}) != 0 {
		t.Fatal("Should be able to sign data")
	}
	if bytes.Compare(signature.Hash[:], []byte{134, 104, 147, 11, 57, 201, 108, 59, 184, 139, 243, 65, 40, 64, 253, 36, 238, 110, 178, 135, 3, 132, 81, 20, 90, 140, 191, 3, 248, 78, 206, 5}) != 0 {
		t.Fatal("Should be able to sign data")
	}
	if bytes.Compare(signature.R[:], []byte{38, 13, 234, 31, 255, 134, 106, 82, 9, 210, 194, 0, 254, 182, 177, 131, 175, 134, 242, 0, 172, 48, 222, 171, 195, 49, 104, 26, 179, 109, 153, 26}) != 0 {
		t.Fatal("Should be able to sign data")
	}
	if bytes.Compare(signature.S[:], []byte{93, 200, 52, 88, 212, 65, 21, 176, 144, 12, 234, 83, 255, 228, 160, 84, 155, 243, 228, 222, 146, 202, 38, 51, 123, 249, 185, 4, 246, 158, 41, 155}) != 0 {
		t.Fatal("Should be able to sign data")
	}
	if signature.V != 28 {
		t.Fatal("Should be able to sign data")
	}
}

func TestIsValidSignature(t *testing.T) {
	privateKeyHex := "6a6594b9673a8c8987113327a93bf4a216d6cbe53e5f31bbaa8c5228a1664591"

	privateKeyECDSA, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		t.Fatal(err)
	}
	pubKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("cannot convert to pubkey")
	}

	signature := SignData([]byte("ROSGVVfIaH5n6INxwgT82mnCaozLJaPzORj7rYzNNLS8tfV3G560i71sIVjefbP4Lzi2RrijpakhLlbrpMaNv75wCzD3LOVWqVEPgEZ8ofXcAaBhtDmI9G0x8qBFWLFN3Ov"), privateKeyECDSA)

	if !IsValidSignature(*publicKeyECDSA, signature) {
		t.Fatal("Should be able to verify a signature")
	}
}

func TestBigIntToECDSAPrivateKey(t *testing.T) {
	x, _ := ethcrypto.GenerateKey()
	y := BigIntToECDSAPrivateKey(*x.D)
	if y.D.Cmp(x.D) != 0 {
		t.Fatal("Should be able to convert a bigint to a private key")
	}
}

func TestPointToEthAddress(t *testing.T) {
	privateKeyHex := "6a6594b9673a8c8987113327a93bf4a216d6cbe53e5f31bbaa8c5228a1664591"
	privateKeyECDSA, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		t.Fatal(err)
	}

	point := common.Point{
		X: *privateKeyECDSA.X,
		Y: *privateKeyECDSA.Y,
	}
	address := PointToEthAddress(point)
	if address.String() != "0x1a6bd03C12F763D42126872aC42662B08A5423B3" {
		t.Fatal("Should be able to convert to an eth address")
	}
}

func TestVerifyPtFromRaw(t *testing.T) {
	msg := []byte("nonce")

	privateKeyHex := "6a6594b9673a8c8987113327a93bf4a216d6cbe53e5f31bbaa8c5228a1664591"
	privateKeyECDSA, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		t.Fatal(err)
	}

	point := common.Point{
		X: *privateKeyECDSA.X,
		Y: *privateKeyECDSA.Y,
	}
	signature, err := ethcrypto.Sign(common.Keccak256(msg), privateKeyECDSA)
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPtFromRaw(msg, point, signature) {
		t.Fatal("Should be able to verify a signature from a point")
	}
}
