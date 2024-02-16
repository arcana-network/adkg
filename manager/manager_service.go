package manager

import (
	"bufio"
	"fmt"
	"os"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/eventbus"
	log "github.com/sirupsen/logrus"
)

const MSG_PREFIX = "TO_MANAGER"

type ManagerService struct {
	bus    eventbus.Bus
	broker *common.MessageBroker
}

func New(bus eventbus.Bus) *ManagerService {
	m := &ManagerService{
		bus:    bus,
		broker: common.NewServiceBroker(bus, common.MANAGER_SERVICE_NAME)}
	return m
}

func (m *ManagerService) ID() string {
	return common.MANAGER_SERVICE_NAME
}

func (m *ManagerService) Start() error {
	go stdin_listener(m.broker)
	log.Info("Manager service running.")
	return nil
}

func (m *ManagerService) Stop() error {
	return nil
}
func (m *ManagerService) IsRunning() bool {
	return true
}

func (m *ManagerService) Call(method string, args ...interface{}) (result interface{}, err error) {
	switch method {
	case "send_to_manager":
		var msg string
		err := common.CastOrUnmarshal(args[0], &msg)
		if err != nil {
			return nil, fmt.Errorf("send_to_manager msg cast error")
		}
		fmt.Printf("%s:%s", MSG_PREFIX, msg)
		return nil, nil

	default:
		return nil, fmt.Errorf("manager service method %v not found", method)
	}

}

func stdin_listener(brocker *common.MessageBroker) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()

		fmt.Printf("child received: %s \n", msg)
		// TODO send pssService msg from manager
		// if msg == "Some pss trigger"{
		// brocker.PSSMethods().TriggerPss(msg)}
		fmt.Printf("child returns: %s \n", "Hello from child")
	}

}
