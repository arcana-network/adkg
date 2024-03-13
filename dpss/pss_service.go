package dpss

import (
	"sync"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/coinbase/kryptology/pkg/core/curves"
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
	bus       eventbus.Bus
	broker    *common.MessageBroker
	running   bool
	pssStatus MessageType
	pssNode   *PSSNode
}

func New(bus eventbus.Bus) *PssService {
	keygenService := &PssService{
		bus:       bus,
		broker:    common.NewServiceBroker(bus, common.PSS_SERVICE_NAME),
		running:   false,
		pssStatus: STOPPED,
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

	service.running = true

	return nil
}

func (service *PssService) IsRunning() bool {
	return service.running
}

func (service *PssService) Stop() error {
	service.running = false
	log.Info("Stopping PSS service")
	return nil
}

func (service *PssService) PssRunning() bool {
	return service.pssStatus == RUNNING
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
		service.pssStatus = RUNNING
		// get self info from ChainService
		chainMethods := service.broker.ChainMethods()
		selfIndex := chainMethods.GetSelfIndex()
		selfPubKey := chainMethods.GetSelfPublicKey()
		selfDetails := common.NodeDetails{
			Index:  selfIndex,
			PubKey: common.Point{X: selfPubKey.X, Y: selfPubKey.Y},
		}
		k := service.broker.ChainMethods().GetSelfPrivateKey()
		priv, err := curves.K256().NewScalar().SetBigInt(&k)
		if err != nil {
			log.Errorf("key error %s", err.Error())
			return nil, err
		}
		// get old epoch and epochInfo
		oldEpoch := chainMethods.GetCurrentEpoch()
		oldEpochInfo, err := chainMethods.GetEpochInfo(oldEpoch, false)
		if err != nil {
			log.Errorf("Could not get currEpochInfo in trigger_pss %s", err.Error())
			return nil, err
		}
		// get new epoch and epochInfo
		newEpoch := int(oldEpochInfo.NextEpoch.Int64())
		newEpochInfo, err := chainMethods.GetEpochInfo(newEpoch, false)
		if err != nil {
			log.Errorf("Could not get newEpochInfo in trigger_pss %s", err.Error())
			return nil, err
		}
		// get new and old node list
		oldNodeList := chainMethods.AwaitCompleteNodeList(oldEpoch)
		newNodeList := chainMethods.AwaitCompleteNodeList(newEpoch)

		// create a PSSNode
		pssNode, err := NewPSSNode(*service.broker,
			selfDetails,
			getCommonNodesFromNodeRefArray(oldNodeList),
			getCommonNodesFromNodeRefArray(newNodeList),
			service.bus,
			int(oldEpochInfo.T.Int64()),
			int(oldEpochInfo.K.Int64()),
			int(newEpochInfo.T.Int64()),
			int(newEpochInfo.K.Int64()),
			priv)
		if err != nil {
			log.Errorf("Could not create pssNode in trigger_pss %s", err.Error())
			return nil, err
		}
		service.pssNode = pssNode
	}

	// TODO add stop_pss
	return nil, nil
}

// TODO - same function exists in keygen_node, move both to common
func getCommonNodesFromNodeRefArray(nodeRefs []common.NodeReference) (commonNodes []common.NodeDetails) {
	for _, nodeRef := range nodeRefs {
		commonNodes = append(commonNodes, common.NodeDetails{
			Index: int(nodeRef.Index.Int64()),
			PubKey: common.Point{
				X: *nodeRef.PublicKey.X,
				Y: *nodeRef.PublicKey.Y,
			},
		})
	}
	return
}
