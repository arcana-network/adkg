package keygen

import (
	"encoding/hex"
	"math/big"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/arcana-network/dkgnode/common"
)

func TestBigIntToPoint(t *testing.T) {
	x := big.NewInt(32)
	y := big.NewInt(54)
	z := BigIntToPoint(x, y)
	if z.X.Cmp(x) != 0 || z.Y.Cmp(y) != 0 {
		t.Fatal("Should be able to generate a point")
	}
}

func TestECDSASign(t *testing.T) {
	privateKeyHex := "6a6594b9673a8c8987113327a93bf4a216d6cbe53e5f31bbaa8c5228a1664591"

	privateKeyECDSA, err := ethcrypto.HexToECDSA(privateKeyHex)
	if err != nil {
		t.Fatal(err)
	}
	signature := ECDSASign("foobar", privateKeyECDSA.X)
	if hex.EncodeToString(signature) != "53ea7ab816dd1ef2e712c74348751f781c068aeb48815b7de6a63d2bb91e02e006919a144b3db7da4e66eb65d691068615e52fc1580f5dead4acac6c52fde92e00" {
		t.Fatal("Should be able to sign a message")
	}
}

func TestECDSAVerify(t *testing.T) {
	privateKeyHex := "6a6594b9673a8c8987113327a93bf4a216d6cbe53e5f31bbaa8c5228a1664591"

	privateKeyECDSA, err := ethcrypto.HexToECDSA(privateKeyHex)
	if err != nil {
		t.Fatal(err)
	}
	point := common.Point{
		X: *privateKeyECDSA.X,
		Y: *privateKeyECDSA.Y,
	}
	msg := "foobar"
	signatureBinary, err := ethcrypto.Sign(common.Keccak256([]byte(msg)), privateKeyECDSA)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	if !ECDSAVerify("foobar", &point, signatureBinary) {
		t.Fatal("Should be able to verify a signature")
	}
}

func TestECDSASignBytes(t *testing.T) {
	privateKeyHex := "6a6594b9673a8c8987113327a93bf4a216d6cbe53e5f31bbaa8c5228a1664591"

	privateKeyECDSA, err := ethcrypto.HexToECDSA(privateKeyHex)
	if err != nil {
		t.Fatal(err)
	}
	msg := "foobar"
	bytes := ECDSASignBytes([]byte(msg), privateKeyECDSA.X)
	if hex.EncodeToString(bytes) != "53ea7ab816dd1ef2e712c74348751f781c068aeb48815b7de6a63d2bb91e02e006919a144b3db7da4e66eb65d691068615e52fc1580f5dead4acac6c52fde92e00" {
		t.Fatal("Should be able to encode bytes")
	}
}
