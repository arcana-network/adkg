package db

import (
	"fmt"
	"math/big"
	"strings"

	eth "github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/torusresearch/bijson"

	"github.com/arcana-network/dkgnode/common"
)

type DBWrapper struct {
	db *leveldb.DB
}

type KeygenStarted struct {
	Started bool
}

type completedShare struct {
	Si      big.Int `json:"si"`
	SiPrime big.Int `json:"si_prime"`
}

var pubkeyToKeyIndexBytes = []byte("f")
var commitmentBytes = []byte("co")
var TBytes = []byte("t")
var completedPSSShareBytes = []byte("b")
var completedShareCountBytes = []byte("c")
var keygenIDBytes = []byte("g")
var connectionDetailsBytes = []byte("i")
var pssCommitmentMatrixBytes = []byte("e")
var nodePubKeyBytes = []byte("j")

func NewDB(path string) (*DBWrapper, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &DBWrapper{db: db}, nil
}
func (t *DBWrapper) RetrieveCompletedShare(keyIndex big.Int) (*big.Int, *big.Int, error) {
	keyIndexBytes := keyIndex.Bytes()
	completedShareKey := append(completedPSSShareBytes, keyIndexBytes...)
	res := t.Get(completedShareKey)
	if res != nil {
		var retrievedShare completedShare
		err := bijson.Unmarshal(res, &retrievedShare)
		if err != nil {
			return nil, nil, err
		}
		return &retrievedShare.Si, &retrievedShare.SiPrime, nil
	}
	return nil, nil, errors.New("Share not found!")
}

func (w *DBWrapper) StoreCompletedPSSShare(keyIndex big.Int, si big.Int, siprime big.Int) error {
	keyIndexBytes := keyIndex.Bytes()
	completedShareKey := append(completedPSSShareBytes, keyIndexBytes...)
	marshalledShare, err := bijson.Marshal(completedShare{
		Si:      si,
		SiPrime: siprime,
	})
	if err != nil {
		return err
	}
	w.Set(completedShareKey, marshalledShare)
	return nil
}

func (w *DBWrapper) StoreCommitment(keyIndex big.Int, T []int, metadata map[string][]common.Point) error {
	keyIndexBytes := keyIndex.Bytes()
	commitmentKey := append(commitmentBytes, keyIndexBytes...)

	log.Debugf("coverted-meta=%v", metadata)
	marshalledCommitment, err := bijson.Marshal(metadata)
	if err != nil {
		return err
	}

	log.Debugf("set-metadata=%v", marshalledCommitment)
	w.Set(commitmentKey, marshalledCommitment)

	// Storing T
	tkey := append(TBytes, keyIndexBytes...)
	log.Debugf("set-metadata-t=%v", T)

	tVal, _ := bijson.Marshal(T)
	w.Set(tkey, tVal)

	return nil
}

// func (w *DBWrapper) FetchCommitment(keyIndex big.Int) map[string][]common.Point {
// 	keyIndexBytes := keyIndex.Bytes()
// 	commitmentKey := append(commitmentBytes, keyIndexBytes...)
// 	tkey := append(TBytes, keyIndexBytes...)
// 	tVal := w.Get(tkey)
// 	var T []int
// 	_ = bijson.Unmarshal(tVal, &T)
// 	log.Infof("get-metadata-T=%v", T)

// 	val := w.Get(commitmentKey)
// 	log.Infof("get-metadata=%v", val)

// 	unmarshalledVal := make(map[string][]common.Point)
// 	_ = bijson.Unmarshal(val, &unmarshalledVal)
// 	return unmarshalledVal
// }

func (w *DBWrapper) RetrieveNodePubKey(nodeAddress eth.Address) (pubKey common.Point, err error) {
	key := append(nodePubKeyBytes, nodeAddress[:]...)
	data := w.Get(key)
	if data == nil {
		return pubKey, fmt.Errorf("could not find pubkey for nodeAddress %s", nodeAddress.String())
	}
	var pubKeyHex *common.HexPoint
	err = bijson.Unmarshal(data, &pubKeyHex)
	pubKey = pubKeyHex.ToPoint()
	return
}

func (w *DBWrapper) GetKeygenStarted(keygenID string) bool {
	key := append(keygenIDBytes, []byte(keygenID)...)
	data := w.Get(key)
	if data == nil {
		return false
	}
	var keygenStarted KeygenStarted
	err := bijson.Unmarshal(data, &keygenStarted)
	if err != nil {
		log.WithError(err).Error("Could not unmarshal get keygen started")
	}
	return keygenStarted.Started
}

func (w *DBWrapper) SetKeygenStarted(keygenID string, started bool) {
	key := append(keygenIDBytes, []byte(keygenID)...)
	data, err := bijson.Marshal(KeygenStarted{Started: started})
	if err != nil {
		log.WithError(err).Error("Could not marshal set keygen started")
	}
	w.Set(key, data)
}

func (t *DBWrapper) RetrievePublicKeyToKeyIndex(publicKey common.Point) (*big.Int, error) {
	b, err := bijson.Marshal(publicKey)
	if err != nil {
		return nil, err
	}
	key := append(pubkeyToKeyIndexBytes, b...)
	var keyIndex big.Int
	keyIndexBytes := t.Get(key)
	keyIndex.SetBytes(keyIndexBytes)
	return &keyIndex, nil
}

func (w *DBWrapper) Has(key []byte) bool {
	return w.Get(key) != nil
}
func (w *DBWrapper) KeyIndexToPublicKeyExists(keyIndex big.Int) bool {
	key := append(keyIndexToPubKeyBytes, keyIndex.Bytes()...)
	return w.Has(key)
}

func (t *DBWrapper) StorePublicKeyToKeyIndex(publicKey common.Point, keyIndex big.Int) error {
	b, err := bijson.Marshal(publicKey)
	if err != nil {
		return err
	}
	// store pubkey -> key index
	pkkey := append(pubkeyToKeyIndexBytes, b...)
	t.Set(pkkey, keyIndex.Bytes())

	// store key index -> pubkey
	kikey := append(keyIndexToPubKeyBytes, keyIndex.Bytes()...)
	t.Set(kikey, b)

	return nil
}

func (t *DBWrapper) StoreNodePubKey(nodeAddress eth.Address, pubKey common.Point) error {
	key := append(nodePubKeyBytes, nodeAddress[:]...)
	pubKeyHex := pubKey.ToHex()
	data, err := bijson.Marshal(pubKeyHex)
	if err != nil {
		return err
	}
	t.Set(key, data)
	return nil
}

func (t *DBWrapper) StoreConnectionDetails(nodeAddress eth.Address, tmP2PConnection string, p2pConnection string) error {
	connectionDetailsKey := append(connectionDetailsBytes, nodeAddress[:]...)
	connectionData := strings.Join([]string{tmP2PConnection, p2pConnection}, common.Delimiter1)
	t.Set(connectionDetailsKey, []byte(connectionData))
	return nil
}

func (t *DBWrapper) RetrieveConnectionDetails(nodeAddress eth.Address) (tmP2PConnection string, p2pConnection string, err error) {
	connectionDetailsKey := append(connectionDetailsBytes, nodeAddress[:]...)
	res := t.Get(connectionDetailsKey)
	if res != nil {
		substrs := strings.Split(string(res), common.Delimiter1)
		if len(substrs) != 2 {
			return "", "", errors.New("unexpected number of substrs in connection details stored data")
		}
		tmP2PConnection = substrs[0]
		p2pConnection = substrs[1]
		return
	}
	return "", "", errors.New("could not get data from db for connection details")
}

func (t *DBWrapper) StorePSSCommitmentMatrix(keyIndex big.Int, c [][]common.Point) error {
	keyIndexBytes := keyIndex.Bytes()
	commitmentMatrixKey := append(pssCommitmentMatrixBytes, keyIndexBytes...)
	b, err := bijson.Marshal(c)
	if err != nil {
		log.WithField("c", c).WithField("keyIndex", keyIndex).Debug("could not store commitment matrix")
		return err
	}
	t.Set(commitmentMatrixKey, b)
	return nil
}

func (w *DBWrapper) Set(key []byte, value []byte) {
	key = nonNilBytes(key)
	value = nonNilBytes(value)
	err := w.db.Put(key, value, nil)
	if err != nil {
		log.WithError(err).Fatal()
	}
}

func (w *DBWrapper) Get(key []byte) []byte {
	key = nonNilBytes(key)
	res, err := w.db.Get(key, nil)
	if err != nil {
		if err == errors.ErrNotFound {
			return nil
		}
		panic(err)
	}
	return res
}

func nonNilBytes(bz []byte) []byte {
	if bz == nil {
		return []byte{}
	}
	return bz
}
