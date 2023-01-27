package tendermint

import (
	"fmt"
	"math/big"
	"os"

	"github.com/arcana-network/dkgnode/common"

	"github.com/arcana-network/dkgnode/eventbus"
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/server"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
)

type ABCIService struct {
	bus    eventbus.Bus
	ABCI   *ABCI
	broker *common.MessageBroker
}

func NewABCI(bus eventbus.Bus) *ABCIService {
	abciService := ABCIService{
		bus:    bus,
		broker: common.NewServiceBroker(bus, common.ABCI_SERVICE_NAME),
	}
	return &abciService
}

func (service *ABCIService) ID() string {
	return common.ABCI_SERVICE_NAME
}

func (service *ABCIService) Start() error {
	service.ABCI = service.ABCI.NewABCI(service.broker)
	socketAddr := "unix://dkg.sock"
	srv := server.NewSocketServer(socketAddr, service.ABCI)
	logger := tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout))

	srv.SetLogger(logger.With("module", "abci-service"))
	log.Info("Starting ABCI server")
	if err := srv.Start(); err != nil {
		log.WithField("err", err).Info()
		return err
	}
	go func() {
		tmos.TrapSignal(logger.With("module", "trap-signal"), func() {
			err := srv.Stop()
			if err != nil {
				log.WithError(err).Error("could not stop service")
			}
		})
	}()
	return nil
}

func (service *ABCIService) IsRunning() bool {
	return true
}

func (service *ABCIService) Stop() error {
	return nil
}

func (a *ABCIService) Call(method string, args ...interface{}) (interface{}, error) {

	switch method {
	case "last_created_index":
		return a.ABCI.state.LastCreatedIndex, nil
	case "last_unassigned_index":
		return a.ABCI.state.LastUnassignedIndex, nil
	case "retrieve_key_mapping":
		var keyIndex big.Int
		_ = common.CastOrUnmarshal(args[0], &keyIndex)

		keyDetails, err := a.ABCI.retrieveKeyMapping(keyIndex)
		if err != nil {
			return nil, err
		}
		return *keyDetails, err
	case "get_indexes_from_verifier_id":
		var provider, userID, appID string
		_ = common.CastOrUnmarshal(args[0], &provider)
		_ = common.CastOrUnmarshal(args[1], &userID)
		_ = common.CastOrUnmarshal(args[2], &appID)
		keyIndexes, err := a.ABCI.getIndexesFromVerifierID(provider, userID, appID)
		return keyIndexes, err
	}

	return nil, fmt.Errorf("ABCI service method %v not found", method)
}
