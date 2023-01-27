package secret

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

type SecretManager interface {
	Setup() error
	GetSecret(name string) ([]byte, error)
	SetSecret(name string, value []byte) error
}

const (
	NodeKey       = "node-key"
	TendermintKey = "tm-key"
)

func InitNodeKey(manager SecretManager) (string, string, error) {
	// Create private key
	key, err := crypto.GenerateKey()
	if err != nil {
		return "", "", err
	}

	keyBytes := crypto.FromECDSA(key)
	address := crypto.PubkeyToAddress(key.PublicKey)

	publicKey := crypto.FromECDSAPub(&key.PublicKey)[1:]

	err = manager.SetSecret(NodeKey, keyBytes)
	if err != nil {
		return "", "", err
	}

	return hex.EncodeToString(publicKey), address.Hex(), err
}

func InitTendermintKey(manager SecretManager) (string, error) {
	// ED25519 for tendermint key
	key := ed25519.GenPrivKey()
	err := manager.SetSecret(TendermintKey, key)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(key.PubKey().Bytes()), err
}

func GetResult(manager SecretManager) (result Result, err error) {
	tmKeyBytes, err := manager.GetSecret(TendermintKey)
	if err != nil {
		return
	}
	result.TendermintKey = hex.EncodeToString(ed25519.PrivKey(tmKeyBytes).PubKey().Bytes())

	nodeKeyBytes, err := manager.GetSecret(NodeKey)
	if err != nil {
		return
	}

	privKey, err := crypto.ToECDSA(nodeKeyBytes)
	if err != nil {
		return
	}

	address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	publicKey := crypto.FromECDSAPub(&privKey.PublicKey)[1:]

	result.NodeKey = hex.EncodeToString(publicKey)
	result.Address = address
	return
}

type Keys struct {
	NodePrivateKey []byte
	TmPrivateKey   []byte
}
