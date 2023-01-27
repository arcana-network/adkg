package cache

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/arcana-network/dkgnode/common"

	"github.com/patrickmn/go-cache"
)

type syncMap struct {
	sync.Map
}

func (s *syncMap) get(key string) *cache.Cache {
	valueInterface, ok := s.Map.Load(key)
	if !ok {
		return nil
	}

	value, ok := valueInterface.(*cache.Cache)
	if !ok {
		return nil
	}

	return value
}

func (s *syncMap) set(key string, value *cache.Cache) {
	s.Map.Store(key, value)
}
func (s *syncMap) exists(key string) (ok bool) {
	_, ok = s.Map.Load(key)
	return
}

func New() *CacheService {
	cacheService := CacheService{}
	return &cacheService
}

type CacheService struct {
	tokenCache    *syncMap
	signerCache   *cache.Cache
	verifierCache *cache.Cache
}

func (*CacheService) ID() string {
	return common.CACHE_SERVICE_NAME
}

func (c *CacheService) Start() error {
	c.tokenCache = &syncMap{}
	c.signerCache = cache.New(cache.NoExpiration, time.Minute)
	c.verifierCache = cache.New(cache.NoExpiration, time.Minute)
	return nil
}

func (c *CacheService) Stop() error {
	return nil
}
func (c *CacheService) IsRunning() bool {
	return true
}

func (c *CacheService) Call(method string, args ...interface{}) (interface{}, error) {
	switch method {
	case "token_commit_exists":

		var provider, hash string
		_ = common.CastOrUnmarshal(args[0], &provider)
		_ = common.CastOrUnmarshal(args[1], &hash)

		exists := c.tokenCommitExists(provider, hash)
		return exists, nil
	case "get_token_commit_key":

		var args0, args1 string
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)

		pubKey := c.getTokenCommitKey(args0, args1)
		return pubKey, nil
	case "record_token_commit":

		var args0, args1 string
		var args2 common.Point
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)
		_ = common.CastOrUnmarshal(args[2], &args2)

		c.recordTokenCommit(args0, args1, args2)
		return nil, nil
	case "signer_sig_exists":

		var args0 string
		_ = common.CastOrUnmarshal(args[0], &args0)
		exists := c.signerSigExists(args0)
		return exists, nil
	case "record_signer_sig":

		var args0 string
		_ = common.CastOrUnmarshal(args[0], &args0)
		return nil, c.recordSignerSig(args0)
	case "store_verifier_clientid":

		var args0 string
		var args1 string
		var args2 string
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)
		_ = common.CastOrUnmarshal(args[2], &args2)
		c.StoreVerifierClientID(args0, args1, args2)
		return nil, nil
	case "store_app_partition":

		var args0 string
		var args1 bool
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)
		c.StoreAppPartition(args0, args1)
		return nil, nil
	case "retrieve_verifier_clientid":

		var args0 string
		var args1 string
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)
		return c.RetrieveVerifierClientID(args0, args1), nil
	case "retrieve_app_partition":

		var args0 string
		_ = common.CastOrUnmarshal(args[0], &args0)
		return c.RetrieveAppPartition(args0)
	}
	return nil, fmt.Errorf("cache service method %v not found", method)
}

func (c *CacheService) StoreVerifierClientID(appID, verifier, clientID string) {
	key := strings.Join([]string{appID, verifier}, common.Delimiter1)
	c.verifierCache.Set(key, clientID, time.Minute*5)
}
func (c *CacheService) RetrieveVerifierClientID(appID, verifier string) string {
	key := strings.Join([]string{appID, verifier}, common.Delimiter1)

	value, found := c.verifierCache.Get(key)
	if !found {
		return ""
	}
	clientID := value.(string)
	return clientID
}

func (c *CacheService) StoreAppPartition(appID string, partitioned bool) {
	c.verifierCache.Set(appID, partitioned, time.Minute*5)
}
func (c *CacheService) RetrieveAppPartition(appID string) (bool, error) {
	value, found := c.verifierCache.Get(appID)
	if !found {
		return false, errors.New("not found")
	}
	partitioned := value.(bool)
	return partitioned, nil
}

func (c *CacheService) signerSigExists(signature string) (exists bool) {
	_, exists = c.signerCache.Get(signature)
	return
}

func (c *CacheService) recordSignerSig(signature string) error {
	return c.signerCache.Add(signature, true, time.Minute)
}

func (c *CacheService) tokenCommitExists(verifier string, tokenCommitment string) (exists bool) {
	if !c.tokenCache.exists(verifier) {
		c.tokenCache.set(verifier, cache.New(cache.NoExpiration, 10*time.Minute))
	}
	_, found := c.tokenCache.get(verifier).Get(tokenCommitment)
	return found
}

func (c *CacheService) getTokenCommitKey(verifier string, tokenCommitment string) (pubKey common.Point) {
	if !c.tokenCache.exists(verifier) {
		c.tokenCache.set(verifier, cache.New(cache.NoExpiration, 10*time.Minute))
	}
	item, found := c.tokenCache.get(verifier).Get(tokenCommitment)
	if found {
		tokenCommitmentData := item.(TokenCommitmentData)
		return tokenCommitmentData.PubKey
	}
	return common.Point{}
}

func (c *CacheService) recordTokenCommit(verifier string, tokenCommitment string, pubKey common.Point) {
	if !c.tokenCache.exists(verifier) {
		c.tokenCache.set(verifier, cache.New(cache.NoExpiration, 10*time.Minute))
	}
	tokenCommitmentData := TokenCommitmentData{Exists: true, PubKey: pubKey}
	c.tokenCache.get(verifier).Set(tokenCommitment, tokenCommitmentData, 90*time.Minute)
}

type TokenCommitmentData struct {
	Exists bool
	PubKey common.Point
}
