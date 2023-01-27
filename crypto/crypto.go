package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	"log"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/torusresearch/bijson"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/secp256k1"
)

type Signature struct {
	Raw  []byte
	Hash [32]byte
	R    [32]byte
	S    [32]byte
	V    uint8
}

type signObject struct {
	Body    []byte `json:"body"`
	PubKeyX string `json:"pubkey_x"`
	PubKeyY string `json:"pubkey_y"`
}

type SignatureHashless struct {
	Raw []byte
	R   [32]byte
	S   [32]byte
	V   uint8
}

func bytes32(bytes []byte) [32]byte {
	tmp := [32]byte{}
	copy(tmp[:], bytes)
	return tmp
}

func SigToHex(ecdsaSig Signature) string {
	return hex.EncodeToString(ecdsaSig.R[:]) + hex.EncodeToString(ecdsaSig.S[:]) + hex.EncodeToString(big.NewInt(int64(ecdsaSig.V)).Bytes())
}

func HexToSig(hexString string) SignatureHashless {
	hexR := hexString[:64]
	hexS := hexString[64:128]
	hexV := hexString[128:130]

	R, _ := hex.DecodeString(hexR)
	S, _ := hex.DecodeString(hexS)

	Vbytes, _ := hex.DecodeString(hexV)
	V := new(big.Int).SetBytes(Vbytes)
	Vuint8 := uint8(V.Int64())

	var signature []byte
	signature = append(signature, R...)
	signature = append(signature, S...)
	signature = append(signature, V.Bytes()...)

	var (
		R32byte [32]byte
		S32byte [32]byte
	)
	copy(R32byte[:], R[:32])
	copy(S32byte[:], S[:32])

	return SignatureHashless{
		signature,
		R32byte,
		S32byte,
		Vuint8,
	}
}

// SignData returns a hex-encoded signature of the passed in data
func SignData(data []byte, ecdsaKey *ecdsa.PrivateKey) Signature {
	// to get data []byte from string, do secp256k1.Keccak256([]byte(messageString))
	hashRaw := secp256k1.Keccak256(data)
	signature, err := ethcrypto.Sign(hashRaw, ecdsaKey)
	if err != nil {
		log.Fatal(err)
	}

	return Signature{
		signature,
		bytes32(hashRaw),
		bytes32(signature[:32]),
		bytes32(signature[32:64]),
		uint8(int(signature[64])) + 27, // Yes add 27, weird Ethereum quirk
	}
}

func signDataWithPubKey(body []byte, pubKeyX string, pubKeyY string, privateKey *ecdsa.PrivateKey) Signature {
	objectToSign := signObject{
		body,
		pubKeyX,
		pubKeyY,
	}

	rawData, err := bijson.Marshal(objectToSign)
	if err != nil {
		log.Fatal(err)
	}

	return SignData(rawData, privateKey)
}

func IsValidSignature(ecdsaPubKey ecdsa.PublicKey, ecdsaSignature Signature) bool {
	r := new(big.Int)
	s := new(big.Int)
	r.SetBytes(ecdsaSignature.R[:])
	s.SetBytes(ecdsaSignature.S[:])

	return ecdsa.Verify(
		&ecdsaPubKey,
		ecdsaSignature.Hash[:],
		r,
		s,
	)
}

func verifyFromRaw(msg []byte, ecdsaPubKey ecdsa.PublicKey, signature []byte) bool {
	r := new(big.Int)
	s := new(big.Int)
	tempR := bytes32(signature[:32])
	tempS := bytes32(signature[32:64])
	r.SetBytes(tempR[:])
	s.SetBytes(tempS[:])
	bytesHash := bytes32(secp256k1.Keccak256(msg))

	return ecdsa.Verify(
		&ecdsaPubKey,
		bytesHash[:],
		r,
		s,
	)
}

func VerifyPtFromRaw(msg []byte, pubKeyPt common.Point, signature []byte) bool {
	ecdsaPubKey := ecdsa.PublicKey{
		Curve: secp256k1.Curve,
		X:     &pubKeyPt.X,
		Y:     &pubKeyPt.Y,
	}

	return verifyFromRaw(msg, ecdsaPubKey, signature)
}

func BigIntToECDSAPrivateKey(x big.Int) *ecdsa.PrivateKey {
	ecdsaPrivKey, err := ethcrypto.ToECDSA(x.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	return ecdsaPrivKey
}

func pointToECDSAPublicKey(point common.Point) *ecdsa.PublicKey {
	return &ecdsa.PublicKey{
		Curve: ethcrypto.S256(),
		X:     &point.X,
		Y:     &point.Y,
	}
}

// PointToEthAddress converts a common point to ETH address format borrowing ethcrypto functions
func PointToEthAddress(point common.Point) *ethcommon.Address {
	addr := ethcrypto.PubkeyToAddress(*pointToECDSAPublicKey(point))
	return &addr
}
