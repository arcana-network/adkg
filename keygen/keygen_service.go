package keygen

import (
	"fmt"
	"sync"

	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/coinbase/kryptology/pkg/core/curves"

	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/common"
)

type KeygenService struct {
	sync.Mutex
	bus        eventbus.Bus
	broker     *common.MessageBroker
	KeygenNode *KeygenNode
}

func New(bus eventbus.Bus) *KeygenService {
	keygenService := &KeygenService{
		bus:    bus,
		broker: common.NewServiceBroker(bus, common.KEYGEN_SERVICE_NAME),
	}
	return keygenService
}

func (*KeygenService) ID() string {
	return common.KEYGEN_SERVICE_NAME
}

func (service *KeygenService) Start() error {
	ChainMethods := service.broker.ChainMethods()
	selfIndex := ChainMethods.GetSelfIndex()
	selfPubKey := ChainMethods.GetSelfPublicKey()
	currEpoch := ChainMethods.GetCurrentEpoch()
	currNodeList := ChainMethods.AwaitCompleteNodeList(currEpoch)
	currEpochInfo, err := ChainMethods.GetEpochInfo(currEpoch, true)
	if err != nil {
		return err
	}

	selfDetails := common.NodeDetails{
		Index:  selfIndex,
		PubKey: common.Point{X: selfPubKey.X, Y: selfPubKey.Y},
	}

	k := service.broker.ChainMethods().GetSelfPrivateKey()
	priv, err := curves.K256().NewScalar().SetBigInt(&k)
	if err != nil {
		return err
	}
	keygenNode, err := NewKeygenNode(
		service.broker,
		selfDetails,
		getCommonNodesFromNodeRefArray(currNodeList),
		service.bus,
		int(currEpochInfo.T.Int64()),
		int(currEpochInfo.K.Int64()),
		priv,
	)
	if err != nil {
		return err
	}

	service.KeygenNode = keygenNode
	return nil
}

func (service *KeygenService) Call(method string, args ...interface{}) (interface{}, error) {
	// dBMethods := service.broker.DBMethods()
	switch method {

	case "receive_message":

		var args0 common.DKGMessage
		err := common.CastOrUnmarshal(args[0], &args0)
		if err != nil {
			return nil, err
		}

		log.WithField("keygen_method", args0.Method).Debug("keygen_service_call")

		if args0.Method == "acss_share" {
			service.Lock()
			defer service.Unlock()
			id, err := common.ADKGIDFromRoundID(args0.RoundID)
			if err != nil {
				return nil, err
			}
			if !service.broker.DBMethods().GetKeygenStarted(string(id)) {
				err := service.broker.DBMethods().SetKeygenStarted(string(id), true)
				if err != nil {
					return nil, err
				}
				service.KeygenNode.tracker.Add(id)
			} else {
				return nil, nil
			}
		}

		pubKey := service.KeygenNode.Details().PubKey
		index := service.broker.ChainMethods().GetSelfIndex()

		details := common.NodeDetails{
			PubKey: pubKey,
			Index:  index,
		}

		log.WithFields(log.Fields{
			"index": args0.RoundID,
			"type":  args0.Method,
		}).Debug("Broker:ReceiveMessage()")
		return nil, service.KeygenNode.Transport.Receive(details, args0)
	case "cleanup":
		var adkgid common.ADKGID
		err := common.CastOrUnmarshal(args[0], &adkgid)
		if err != nil {
			return nil, err
		}
		service.KeygenNode.BFTDecided(adkgid)
		return nil, nil
	}
	return nil, fmt.Errorf("keygen service method %v not found", method)
}

func (service *KeygenService) Stop() error {
	log.Info("Stopping keygen service")
	return nil
}
func (service *KeygenService) IsRunning() bool {
	return true
}

type KeygenProtocolPrefix string
