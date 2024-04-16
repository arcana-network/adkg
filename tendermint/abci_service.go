package tendermint

import (
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/arcana-network/dkgnode/common"

	"github.com/arcana-network/dkgnode/eventbus"
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/server"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/service"
)

type ABCIService struct {
	bus          eventbus.Bus
	ABCI         *ABCI
	broker       *common.MessageBroker
	socketServer service.Service
}

// TODO test if this works on all system
// must remove old socket file
func cleanupSockFile(p string) {
	sock_file := strings.Split(p, "//")[1]
	if common.DoesFileExist(sock_file) {
		_ = os.Remove(sock_file)
	}
}

func NewABCI(bus eventbus.Bus) *ABCIService {
	abciService := ABCIService{
		bus:    bus,
		broker: common.NewServiceBroker(bus, common.ABCI_SERVICE_NAME),
	}
	return &abciService
}

func (s *ABCIService) ID() string {
	return common.ABCI_SERVICE_NAME
}

func (s *ABCIService) Start() error {
	s.ABCI = s.ABCI.NewABCI(s.broker)
	socketAddr := common.GetSocketAddress()
	cleanupSockFile(socketAddr)
	s.socketServer = server.NewSocketServer(socketAddr, s.ABCI)
	logger := tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout))

	s.socketServer.SetLogger(logger.With("module", "abci-service"))
	log.Info("Starting ABCI server")

	if err := s.socketServer.Start(); err != nil {
		log.WithError(err).Error("ABCI.SocketServer.Start()")
		return err
	}
	return nil
}

func (service *ABCIService) IsRunning() bool {
	return true
}

func (service *ABCIService) Stop() error {
	return service.socketServer.Stop()
}

func (a *ABCIService) Call(method string, args ...interface{}) (interface{}, error) {

	switch method {
	case "last_created_index":
		return a.ABCI.state.LastCreatedIndex, nil
	case "last_unassigned_index":
		return a.ABCI.state.LastUnassignedIndex, nil
	case "last_c25519_created_index":
		return a.ABCI.state.C25519State.LastCreatedIndex, nil
	case "last_c25519_unassigned_index":
		return a.ABCI.state.C25519State.LastUnassignedIndex, nil
	case "retrieve_key_mapping":
		var keyIndex big.Int
		var curve common.CurveName
		_ = common.CastOrUnmarshal(args[0], &keyIndex)
		_ = common.CastOrUnmarshal(args[1], &curve)

		keyDetails, err := a.ABCI.retrieveKeyMapping(keyIndex, curve)
		if err != nil {
			return nil, err
		}
		return *keyDetails, err
	case "get_indexes_from_verifier_id":
		var provider, userID, appID string
		var curve common.CurveName
		_ = common.CastOrUnmarshal(args[0], &provider)
		_ = common.CastOrUnmarshal(args[1], &userID)
		_ = common.CastOrUnmarshal(args[2], &appID)
		_ = common.CastOrUnmarshal(args[3], &curve)

		keyIndexes, err := a.ABCI.getIndexesFromVerifierID(provider, userID, appID, curve)
		return keyIndexes, err
	}

	return nil, fmt.Errorf("ABCI service method %v not found", method)
}
