package manager

import (
	"fmt"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/eventbus"
	log "github.com/sirupsen/logrus"
)

const (
	MSG_PREFIX     = "MSG_TO_MANAGER:"
	MSG_DPSS_START = "DPSS_START"
	MSG_DPSS_END   = "DPSS_END"
	MSG_SHUT_DOWN  = "KILL_NODE"
)

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
	//go stdin_listener(m.broker)
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
		// print a message with a MSG_PREFIX
		var msg string
		err := common.CastOrUnmarshal(args[0], &msg)
		if err != nil {
			return nil, fmt.Errorf("send_to_manager msg cast error")
		}
		fmt.Printf("%s%s\n", MSG_PREFIX, msg)
		return nil, nil
	case "send_dpss_start":
		// print dpss starting message
		fmt.Printf("%s%s\n", MSG_PREFIX, MSG_DPSS_START)
		return nil, nil
	case "send_dpss_end":
		// print dpss ending message
		fmt.Printf("%s%s\n", MSG_PREFIX, MSG_DPSS_END)
		return nil, nil
	case "send_kill_process":
		// print kill process message
		fmt.Printf("%s%s\n", MSG_PREFIX, MSG_SHUT_DOWN)
		return nil, nil

	default:
		return nil, fmt.Errorf("manager service method %v not found", method)
	}

}

//TODO - stdin listener unused for now, consider delete it
// func stdin_listener(brocker *common.MessageBroker) {
// 	scanner := bufio.NewScanner(os.Stdin)
// 	for scanner.Scan() {
// 		msg := scanner.Text()

// 		fmt.Printf("child received: %s \n", msg)
// 		// TODO send pssService msg from manager
// 		// if msg == "Some pss trigger"{
// 		// brocker.PSSMethods().TriggerPss(msg)}
// 		fmt.Printf("child returns: %s \n", "Hello from child")
// 	}

// }
