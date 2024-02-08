package common

import (
	"crypto/ecdsa"
	"math/big"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/secp256k1"
)

func BigIntToPoint(x, y *big.Int) Point {
	return Point{X: *x, Y: *y}
}

func ECDSAVerify(str string, pubKey *Point, signature []byte) bool {
	r := new(big.Int)
	s := new(big.Int)
	r.SetBytes(signature[:32])
	s.SetBytes(signature[32:64])

	ecdsaPubKey := &ecdsa.PublicKey{
		Curve: secp256k1.Curve,
		X:     &(*pubKey).X,
		Y:     &(*pubKey).Y,
	}

	return ecdsa.Verify(
		ecdsaPubKey,
		secp256k1.Keccak256([]byte(str)),
		r,
		s,
	)
}

func ECDSASignBytes(b []byte, privKey *big.Int) []byte {
	return ECDSASign(string(b), privKey)
}

// ECDSASign creates signatures using Ethereum ECDSA (where randomness is deterministic)
func ECDSASign(s string, privKey *big.Int) []byte {
	ecdsaPrivKey, _ := ethCrypto.ToECDSA(privKey.Bytes())
	hashRaw := secp256k1.Keccak256([]byte(s))
	signature, err := ethCrypto.Sign(hashRaw, ecdsaPrivKey)
	if err != nil {
		log.Fatal(err)
	}
	return signature
}
