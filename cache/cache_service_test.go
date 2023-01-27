package cache

import (
	"math/big"
	"testing"

	"github.com/arcana-network/dkgnode/common"
)

func TestCacheService_StoreAppPartition(t *testing.T) {
	c := New()
	c.Start()

	c.StoreAppPartition("appid", false)

	partition, err := c.RetrieveAppPartition("appid")
	if err != nil {
		t.Fatal(err)
	}
	if partition {
		t.Fatal("Should be able to retrieve an partition")
	}

	c.StoreAppPartition("appid2", true)

	partition2, err := c.RetrieveAppPartition("appid2")
	if err != nil {
		t.Fatal(err)
	}
	if !partition2 {
		t.Fatal("Should be able to retrieve an artition")
	}
}

func TestCacheService_StoreVerifierClientID(t *testing.T) {
	c := New()
	c.Start()

	c.StoreVerifierClientID("appid", "verifier", "clientid")
	id := c.RetrieveVerifierClientID("appid", "verifier")
	if id != "clientid" {
		t.Fatal("Should be able to retrieve a verifier client id")
	}
}

func TestCacheService_Call_token_commit_exists(t *testing.T) {
	c := New()
	c.Start()

	call, err := c.Call("token_commit_exists", "providerverifier", "hash")
	if err != nil {
		t.Fatal(err)
	}
	exists, ok := call.(bool)
	if !ok {
		t.Fatal("Should return a bool")
	}
	if exists {
		t.Fatal("Should not return true for token_commit_exists")
	}
	c.recordTokenCommit("providerverifier", "tokencommitment", common.Point{})
	call, err = c.Call("token_commit_exists", "providerverifier", "tokencommitment")
	if err != nil {
		t.Fatal(err)
	}

	exists, ok = call.(bool)
	if !ok {
		t.Fatal("Should return a bool")
	}
	if !exists {
		t.Fatal("Should return true for token_commit_exists")
	}
}

func TestCacheService_Call_get_token_commit_key(t *testing.T) {
	c := New()
	c.Start()

	call, err := c.Call("get_token_commit_key", "providerverifier", "hash")
	if err != nil {
		t.Fatal(err)
	}
	retrievedPubkey, ok := call.(common.Point)
	if !ok {
		t.Fatal("Should return a common.Point")
	}

	var x, y big.Int
	x.SetInt64(90)
	y.SetUint64(39)
	pubKey := common.Point{X: x, Y: y}
	c.recordTokenCommit("providerverifier", "tokencommitment", pubKey)
	call, err = c.Call("get_token_commit_key", "providerverifier", "tokencommitment")
	if err != nil {
		t.Fatal(err)
	}

	retrievedPubkey, ok = call.(common.Point)
	if !ok {
		t.Fatal("Should return a bool")
	}
	if retrievedPubkey.X.Cmp(&pubKey.X) != 0 || retrievedPubkey.Y.Cmp(&pubKey.Y) != 0 {
		t.Fatal("Should return the correct pubkey")
	}
}

func TestCacheService_Call_record_token_commit(t *testing.T) {
	c := New()
	c.Start()

	call, err := c.Call("record_token_commit", "providerverifier", "hash")
	if err != nil {
		t.Fatal(err)
	}
	retrievedPubkey, ok := call.(common.Point)
	if !ok {
		t.Fatal("Should return a common.Point")
	}

	var x, y big.Int
	x.SetInt64(90)
	y.SetUint64(39)
	pubKey := common.Point{X: x, Y: y}
	c.recordTokenCommit("providerverifier", "tokencommitment", pubKey)
	call, err = c.Call("get_token_commit_key", "providerverifier", "tokencommitment")
	if err != nil {
		t.Fatal(err)
	}

	retrievedPubkey, ok = call.(common.Point)
	if !ok {
		t.Fatal("Should return a bool")
	}
	if retrievedPubkey.X.Cmp(&pubKey.X) != 0 || retrievedPubkey.Y.Cmp(&pubKey.Y) != 0 {
		t.Fatal("Should return the correct pubkey")
	}
}

func TestCacheService_Call_signer_sig_exists(t *testing.T) {
	c := New()
	c.Start()

	call, err := c.Call("signer_sig_exists", "signature")
	if err != nil {
		t.Fatal(err)
	}
	exists, ok := call.(bool)
	if !ok {
		t.Fatal("Should return a bool")
	}
	if exists {
		t.Fatal("Should not retrieve a non-existing signature")
	}
	c.recordSignerSig("signature")
	call, err = c.Call("signer_sig_exists", "signature")
	if err != nil {
		t.Fatal(err)
	}

	exists, ok = call.(bool)
	if !ok {
		t.Fatal("Should return a bool")
	}
	if !exists {
		t.Fatal("Should retrieve an existing signature")
	}
}

func TestCacheService_Call_record_signer_sig(t *testing.T) {
	c := New()
	c.Start()

	_, err := c.Call("record_signer_sig", "signature")
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Call("record_signer_sig", "signature")
	if err == nil {
		t.Fatal("Should give an error for an already registered signature")
	}
}

func TestCacheService_Call_store_verifier_clientid_retrieve_verifier_clientid(t *testing.T) {
	c := New()
	c.Start()

	_, err := c.Call("store_verifier_clientid", "appid", "verifier", "clientid")
	if err != nil {
		t.Fatal(err)
	}
	call, err := c.Call("retrieve_verifier_clientid", "appid", "verifier")
	if err != nil {
		t.Fatal(err)
	}
	clientID := call.(string)
	if clientID != "clientid" {
		t.Fatal("Should be able to retrieve a client id")
	}

}

func TestCacheService_Call_store_app_artition_retrieve_app_partition(t *testing.T) {
	c := New()
	c.Start()

	_, err := c.Call("store_app_partition", "appid", true)
	if err != nil {
		t.Fatal(err)
	}
	call, err := c.Call("retrieve_app_partition", "appid")
	if err != nil {
		t.Fatal(err)
	}
	partition := call.(bool)
	if !partition {
		t.Fatal("Should be able to retrieve an partition")
	}

}
