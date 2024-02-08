package pss

import (
	"sync"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/eventbus"
	log "github.com/sirupsen/logrus"
)

type MessageType int32

const (
	RUNNING MessageType = 1
	PAUSED  MessageType = 0
	STOPPED MessageType = -1
)

type PssService struct {
	sync.Mutex
	bus        eventbus.Bus
	broker     *common.MessageBroker
	status      MessageType
}

func New(bus eventbus.Bus) *PssService {
	keygenService := &PssService{
		bus:    bus,
		broker: common.NewServiceBroker(bus, common.PSS_SERVICE_NAME),
		status: STOPPED,
	}
	return keygenService
}

func (*PssService) ID() string {
	return common.PSS_SERVICE_NAME
}

func (service *PssService) Start() error {
	// TODO what needs to be added here?

	ChainMethods := service.broker.ChainMethods()
	currEpoch := ChainMethods.GetCurrentEpoch()
	// We'll probably need the currEpochInfo that is retrieved here
	_, err := ChainMethods.GetEpochInfo(currEpoch, true)
	if err != nil {
		return err
	}

	service.status = RUNNING

	return nil
}

func (service *PssService) IsRunning() bool {
	return service.status == RUNNING
}

func (service *PssService) Stop() error {
	service.status = STOPPED
	log.Info("Stopping PSS service")
	return nil
}


func (service *PssService) Call(method string, args ...interface{}) (interface{}, error) {
	switch method {

	case "trigger_pss":
		// TODO add functionality

		/*
		1. Abort keygen
		2. Retrieve old & new committee from smart contract
		3. Create PssNode
			PssNode in startup process 
			- should figure out the type it is
			- must connect to the correct committee
		*/
	}
	return nil, nil
}