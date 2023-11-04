package db

import (
	"fmt"
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"
)

type DBService struct {
	dbInstance *DBWrapper
}

func New() *DBService {
	return &DBService{}
}

func (*DBService) ID() string {
	return common.DB_SERVICE_NAME
}
func (service *DBService) Start() error {
	dbPath := fmt.Sprintf("%s/keygendb", config.GlobalConfig.BasePath)
	db, err := NewDB(dbPath)
	if err != nil {
		return err
	}
	service.dbInstance = db
	return nil
}

func (service *DBService) Stop() error {
	return nil
}
func (service *DBService) IsRunning() bool {
	return true
}

func (d *DBService) Call(method string, args ...interface{}) (interface{}, error) {

	switch method {
	case "set_keygen_started":

		var args0 string
		var args1 bool
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)

		d.dbInstance.SetKeygenStarted(args0, args1)
		return nil, nil
	case "get_keygen_started":

		var args0 string
		_ = common.CastOrUnmarshal(args[0], &args0)

		return d.dbInstance.GetKeygenStarted(args0), nil

	case "retrieve_public_key_to_index":

		var args0 common.Point
		_ = common.CastOrUnmarshal(args[0], &args0)

		keyIndex, err := d.dbInstance.RetrievePublicKeyToKeyIndex(args0)
		if err != nil || keyIndex == nil {
			return big.NewInt(-1), err
		}
		return *keyIndex, err
	case "index_to_public_key_exists":

		var args0 big.Int
		var args1 common.CurveName
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)

		exists := d.dbInstance.KeyIndexToPublicKeyExists(args0, args1)
		return exists, nil
	case "store_public_key_to_index":

		var args0 common.Point
		var args1 big.Int
		var args2 common.CurveName
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)

		err := d.dbInstance.StorePublicKeyToKeyIndex(args0, args1, args2)
		return nil, err
	case "retrieve_completed_share":

		var args0 big.Int
		var curve common.CurveName

		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &curve)

		rs := new(struct {
			Si      big.Int
			Siprime big.Int
		})
		si, sip, err := d.dbInstance.RetrieveCompletedShare(args0, curve)
		if si == nil {
			si = big.NewInt(0)
		}
		if sip == nil {
			sip = big.NewInt(0)
		}
		rs.Si = *si
		rs.Siprime = *sip
		return *rs, err
	case "store_PSS_commitment_matrix":

		var args0 big.Int
		var args1 [][]common.Point
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)

		err := d.dbInstance.StorePSSCommitmentMatrix(args0, args1)
		return nil, err
	case "store_completed_PSS_share":

		var args0, args1, args2 big.Int
		var curve common.CurveName
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)
		_ = common.CastOrUnmarshal(args[2], &args2)
		_ = common.CastOrUnmarshal(args[3], &curve)

		err := d.dbInstance.StoreCompletedPSSShare(args0, args1, args2, curve)
		return nil, err
	case "store_sharing_commitment":

		var args0 big.Int
		var args1 []int
		var args2 map[string][]common.Point
		var curve common.CurveName

		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)
		_ = common.CastOrUnmarshal(args[2], &args2)
		_ = common.CastOrUnmarshal(args[3], &curve)

		err := d.dbInstance.StoreCommitment(args0, args1, args2, curve)
		return nil, err
	case "store_connection_details":

		var args0 eth.Address
		var args1 common.ConnectionDetails
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)

		err := d.dbInstance.StoreConnectionDetails(args0, args1.TMP2PConnection, args1.P2PConnection)
		return nil, err
	case "retrieve_connection_details":
		var args0 eth.Address
		_ = common.CastOrUnmarshal(args[0], &args0)

		tmP2PConnection, P2PConnection, err := d.dbInstance.RetrieveConnectionDetails(args0)
		return common.ConnectionDetails{
			TMP2PConnection: tmP2PConnection,
			P2PConnection:   P2PConnection,
		}, err
	case "retrieve_node_pub_key":

		var args0 eth.Address
		_ = common.CastOrUnmarshal(args[0], &args0)

		return d.dbInstance.RetrieveNodePubKey(args0)
	case "store_node_pub_key":

		var args0 eth.Address
		var args1 common.Point
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)

		err := d.dbInstance.StoreNodePubKey(args0, args1)
		return nil, err
	}
	return "", nil

}
