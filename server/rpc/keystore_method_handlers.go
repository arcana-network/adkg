package rpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/arcana-network/dkgnode/common"

	ecies "github.com/ecies/go/v2"
	eth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	fastjson "github.com/goccy/go-json"
	"github.com/osamingo/jsonrpc/v2"
)

type StoreParams struct {
	StoreSignatureSkeleton
	Signature string `json:"signature"`
}

type StoreResult struct {
	Ok bool `json:"ok"`
}

type StoreSignatureSkeleton struct {
	TxHash         string `json:"tx_hash"`
	EncryptedShare string `json:"encrypted_share"`
}

type TxDetails struct {
	Address string `json:"address"`
	DID     string `json:"did"`
}

func signHash(sigParams []byte) ethcommon.Hash {
	dataHash := ethcrypto.Keccak256Hash(sigParams)
	hexData := hexutil.Encode(dataHash.Bytes())
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(hexData), hexData)

	return ethcrypto.Keccak256Hash([]byte(msg))
}

var hexPrefix = "0x"

func (h StoreKeyRequestHandler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	var p StoreParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	// Get tx details from hash and get addr
	expectedData := TxDetails{}
	if err := getUploadTxDetails(p.TxHash, h.client, &expectedData); err != nil {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "error while fetching tx details"}
	}

	// Verify signature and get pub key
	sigParams, err := json.Marshal(StoreSignatureSkeleton{p.TxHash, p.EncryptedShare})
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "could not generate hash for signature"}
	}

	normalizedSignature, _ := hex.DecodeString(strings.TrimPrefix(p.Signature, hexPrefix))
	normalizedSignature[64] -= 27

	pk, err := ethcrypto.Ecrecover(signHash(sigParams).Bytes(), normalizedSignature)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "unable to verify signature"}
	}

	// Compared recovered PK to addr from tx
	ecdsaPK, _ := ethcrypto.UnmarshalPubkey(pk)
	addr := ethcrypto.PubkeyToAddress(*ecdsaPK)

	if !strings.EqualFold(strings.TrimPrefix(addr.Hex(), hexPrefix), expectedData.Address) {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "address did not match"}
	}

	// Decrypt share
	broker := common.NewServiceBroker(h.bus, "store_key_request_handler")
	privateKey := broker.ChainMethods().GetSelfPrivateKey()
	encryptedBytes, err := hex.DecodeString(p.EncryptedShare)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "invalid encrypted share"}
	}
	share, err := ecies.Decrypt(ecies.NewPrivateKeyFromBytes(privateKey.Bytes()), encryptedBytes)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "could not decrypt share"}
	}

	// Store share to storage
	err = broker.KeystoreMethods().StoreShare(expectedData.DID, share)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "error while storing share"}
	}
	return StoreResult{Ok: true}, nil
}

type RetrieveParams struct {
	RetrieveSignatureSkeleton
	Signature string `json:"signature"`
}

type RetrieveSignatureSkeleton struct {
	TxHash string `json:"tx_hash"`
}

type RetrieveResult struct {
	Share string `json:"share"`
}

func (h RetrieveKeyRequestHandler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	var p RetrieveParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	// Get tx details from hash and get addr
	expectedData := TxDetails{}
	if err := getDownloadTxDetails(p.TxHash, h.client, &expectedData); err != nil {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "error while fetching tx details"}
	}

	// Verify signature and get pub key
	sigParams, err := json.Marshal(RetrieveSignatureSkeleton{p.TxHash})
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "could not generate hash for signature"}
	}

	normalizedSignature, _ := hex.DecodeString(strings.TrimPrefix(p.Signature, hexPrefix))
	normalizedSignature[64] -= 27

	pk, err := ethcrypto.Ecrecover(signHash(sigParams).Bytes(), normalizedSignature)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "unable to verify signature"}
	}

	// Compared recovered PK to addr from tx
	ecdsaPK, _ := ethcrypto.UnmarshalPubkey(pk)
	addr := ethcrypto.PubkeyToAddress(*ecdsaPK)

	if !strings.EqualFold(strings.TrimPrefix(addr.Hex(), hexPrefix), expectedData.Address) {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "address did not match"}
	}

	// Retrieve share from storage
	broker := common.NewServiceBroker(h.bus, "retrieve_key_request_handler")
	share, err := broker.KeystoreMethods().RetrieveShare(expectedData.DID)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "error while storing share"}
	}

	pubKey, err := ecies.NewPublicKeyFromBytes(pk)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "invalid public key"}
	}

	ciphertext, err := ecies.Encrypt(pubKey, share)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "could not decrypt share"}
	}

	// Send back share
	return RetrieveResult{Share: hex.EncodeToString(ciphertext)}, nil
}

func getDownloadTxDetails(hash string, client *ethclient.Client, details *TxDetails) error {
	v, err := getTransactionDetails(hash, client)
	if err != nil {
		return err
	}
	details.Address = v[744:784]
	details.DID = v[592:656]
	return nil
}
func getUploadTxDetails(hash string, client *ethclient.Client, details *TxDetails) error {
	v, err := getTransactionDetails(hash, client)
	if err != nil {
		return err
	}
	details.Address = v[872:912]
	details.DID = v[592:656]
	return nil
}

func getTransactionDetails(hash string, client *ethclient.Client) (string, error) {
	receipt, _, err := client.TransactionByHash(context.Background(), eth.HexToHash(hash))
	if err != nil {
		fmt.Println("[ERROR]", err)
		return "", err
	}
	val := receipt.Data()
	v := hex.EncodeToString(val)
	return v, nil
}
