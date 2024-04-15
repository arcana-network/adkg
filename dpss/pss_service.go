package dpss

import (
	"math/big"
	"sync"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
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
	bus             eventbus.Bus
	broker          *common.MessageBroker
	running         bool
	pssStatus       MessageType
	pssNode         *PSSNode
	batchFinChannel chan struct{}
}

func New(bus eventbus.Bus) *PssService {
	keygenService := &PssService{
		bus:             bus,
		broker:          common.NewServiceBroker(bus, common.PSS_SERVICE_NAME),
		running:         false,
		pssStatus:       STOPPED,
		batchFinChannel: make(chan struct{}),
	}
	return keygenService
}

func (*PssService) ID() string {
	return common.PSS_SERVICE_NAME
}

func (service *PssService) Start() error {
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

// TODO - check if we need this here, currently using one in chain_service
func (service *PssService) PssRunning() bool {
	return service.pssStatus == RUNNING
}

func (service *PssService) Call(method string, args ...interface{}) (interface{}, error) {
	switch method {

	case "trigger_pss":
		// TODO - check what to do for the new committee nodes
		service.pssStatus = RUNNING
		batchSize := uint(500)
		isNewCommittee, err := service.broker.ChainMethods().IsNewCommittee()
		if err != nil {
			log.Errorf("Could not get isNewCommittee %s", err.Error())
			return nil, err
		}
		if !isNewCommittee {
			// send DpssStart msg to manager to create a new child process for new node
			err := service.broker.ManagerMethods().SendDpssStart()
			if err != nil {
				log.Errorf("unable to send DPSS start message: %s", err.Error())
			}

		}
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

		// get total assigned share num for secp256k1
		secpShareNum, err := service.broker.ABCIMethods().LastUnassignedIndex()
		if err != nil {
			log.Errorf("Could not get share numebr %s", err.Error())
			return nil, err
		}
		secpShareNum -= 1
		// calculate batch needed
		ceil := (secpShareNum % batchSize) != 0
		secpBatchNum := secpShareNum / batchSize
		if ceil {
			secpBatchNum += 1
		}

		// get total assigned share num for c25519
		c25519ShareNum, err := service.broker.ABCIMethods().LastC25519UnassignedIndex()
		if err != nil {
			log.Errorf("Could not get share numebr %s", err.Error())
			return nil, err
		}
		c25519ShareNum -= 1
		// calculate batch needed
		ceil = (c25519ShareNum % batchSize) != 0
		c25519BatchNum := c25519ShareNum / batchSize
		if ceil {
			c25519BatchNum += 1
		}

		if !isNewCommittee {
			// only nodes in old committee need to initiate DPSS
			go service.BatchRunDPSS(secpBatchNum, c25519BatchNum, batchSize, secpShareNum, c25519ShareNum)
		}
	}
	return nil, nil
}

func (service *PssService) BatchRunDPSS(secpBatchNum uint, c25519BatchNum uint, batchSize uint, secpShareNum uint, c25519ShareNum uint) {

	// secp256k1 shares
	for currentBatch := uint(0); currentBatch < secpBatchNum; currentBatch++ {
		var oldShares []sharing.ShamirShare
		// get old shares list of the batch
		for i := uint(0); i < batchSize; i++ {
			index := int64(currentBatch*batchSize + i)
			if index > int64(secpShareNum) {
				log.WithFields(log.Fields{
					"type":            "secp256k1",
					"last index":      index - 1,
					"total share num": secpShareNum,
				}).Debug("Last share added")
				break
			}
			si, _, err := service.broker.DBMethods().RetrieveCompletedShare(*big.NewInt(index), common.SECP256K1)
			if err != nil {
				log.Errorf("unable to ertrieve secp256k1 share of index %v: %s", index, err.Error())
			}
			id := service.broker.ChainMethods().GetSelfIndex()
			share := sharing.ShamirShare{
				Id:    uint32(id),
				Value: si.Bytes(),
			}
			oldShares = append(oldShares, share)
		}

		if len(oldShares) > 0 {
			log.WithFields(log.Fields{
				"type":  "secp256k1",
				"batch": currentBatch,
			}).Info("Running DPSS")
			// Todo: what message to send here?
			// dacss.NewInitMessage(,oldoldShares,)
			// block until the batch has finished
			<-service.batchFinChannel
			log.WithFields(log.Fields{
				"type":  "secp256k1",
				"batch": currentBatch,
			}).Info("DPSS finished")
		}

	}

	// ed25519 shares
	for currentBatch := uint(0); currentBatch < c25519BatchNum; currentBatch++ {
		var oldShares []sharing.ShamirShare
		// get old shares list of the batch
		for i := uint(0); i < batchSize; i++ {
			index := int64(currentBatch*batchSize + i)
			if index > int64(c25519ShareNum) {
				log.WithFields(log.Fields{
					"type":            "ed25519",
					"last index":      index - 1,
					"total share num": c25519ShareNum,
				}).Debug("Last share added")
				break
			}
			si, _, err := service.broker.DBMethods().RetrieveCompletedShare(*big.NewInt(index), common.ED25519)
			if err != nil {
				log.Errorf("unable to ertrieve ed25519 share of index %v: %s", index, err.Error())
			}
			id := service.broker.ChainMethods().GetSelfIndex()
			share := sharing.ShamirShare{
				Id:    uint32(id),
				Value: si.Bytes(),
			}
			oldShares = append(oldShares, share)
		}

		if len(oldShares) > 0 {

			log.WithFields(log.Fields{
				"type":  "ed25519",
				"batch": currentBatch,
			}).Info("Running DPSS")
			// Todo: what message to send here?
			// dacss.NewInitMessage(,oldoldShares,)
			// block until the batch has finished
			<-service.batchFinChannel
			log.WithFields(log.Fields{
				"type":  "ed25519",
				"batch": currentBatch,
			}).Info("DPSS finished")
		}

	}

}

// Todo add this in dpss flow
// function to indicate a dpss batch has finished
func (service *PssService) BatchFinCallBack() {
	service.batchFinChannel <- struct{}{}
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
