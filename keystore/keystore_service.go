package keystore

import (
	"fmt"

	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"
)

type KeystoreService struct {
	db  *leveldb.DB
	bus eventbus.Bus
}

func New(bus eventbus.Bus) *KeystoreService {
	k := &KeystoreService{bus: bus}
	return k
}

func (k *KeystoreService) ID() string {
	return common.KEYSTORE_SERVICE_NAME
}

func (k *KeystoreService) Start() error {
	db, err := leveldb.OpenFile(fmt.Sprintf("%s/sym_shares", config.GlobalConfig.BasePath), nil)
	if err != nil {
		return err
	}
	k.db = db
	return nil
}
func (k *KeystoreService) Stop() error {
	k.db.Close()
	return nil
}
func (k *KeystoreService) IsRunning() bool {
	return true
}
func (k *KeystoreService) Call(method string, args ...interface{}) (result interface{}, err error) {
	switch method {
	case "store":
		var id string
		var share []byte
		_ = common.CastOrUnmarshal(args[0], &id)
		_ = common.CastOrUnmarshal(args[1], &share)
		err := storeKeyShare(k.db, id, share)
		return nil, err
	case "retrieve":
		var id string
		_ = common.CastOrUnmarshal(args[0], &id)
		data, err := retrieveKeyShare(k.db, id)
		return data, err
	default:
		return nil, fmt.Errorf("keystore service method %v not found", method)
	}
}

var symmetricShareByte = []byte("ss")

func storeKeyShare(db *leveldb.DB, id string, share []byte) error {
	k := append(symmetricShareByte, []byte(id)...)
	err := db.Put(k, share, nil)
	return err
}

func retrieveKeyShare(db *leveldb.DB, id string) ([]byte, error) {
	k := append(symmetricShareByte, []byte(id)...)
	data, err := db.Get(k, nil)
	return data, err
}
