package dpss

import (
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

type MessageType int32

const (
	RUNNING   MessageType = 1
	PAUSED    MessageType = 0
	STOPPED   MessageType = -1
	BATCHSIZE int         = 500
)

type PssService struct {
	bus              eventbus.Bus
	broker           *common.MessageBroker
	running          bool
	pssStatus        MessageType
	pssNode          *PSSNode
	currentSecpBatch int
	currentC255Batch int
	secpBatchNum     int
	c25519BatchNum   int
	secpShareNum     int
	c25519ShareNum   int
	newEpochInfo     common.EpochInfo
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

func (pssService *PssService) SetNode(pssNode *PSSNode) {
	pssService.pssNode = pssNode
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
		log.Info("PSS WAS TRIGGERED")
		// TODO - check what to do for the new committee nodes
		service.pssStatus = RUNNING
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

		log.Info("CREATING PSS NODE")
		// create a PSSNode
		var epochForNode int
		if isNewCommittee {
			epochForNode = newEpoch
		} else {
			epochForNode = oldEpoch
		}
		pssNode, err := NewPSSNode(*service.broker,
			selfDetails,
			getCommonNodesFromNodeRefArray(oldNodeList),
			getCommonNodesFromNodeRefArray(newNodeList),
			service.bus,
			int(oldEpochInfo.T.Int64()),
			int(oldEpochInfo.K.Int64()),
			int(newEpochInfo.T.Int64()),
			int(newEpochInfo.K.Int64()),
			priv,
			epochForNode)
		if err != nil {
			log.Errorf("Could not create pssNode in trigger_pss %s", err.Error())
			return nil, err
		}
		service.SetNode(pssNode)

		// only nodes in old committee need to initiate DPSS
		if !isNewCommittee {
			// get total assigned share num for secp256k1
			secpShareNumUint, err := service.broker.ABCIMethods().LastUnassignedIndex()
			if err != nil {
				log.Errorf("Could not get share number %s", err.Error())
				return nil, err
			}
			secpShareNum := int(secpShareNumUint)
			// calculate batch needed
			ceil := (secpShareNum % BATCHSIZE) != 0
			secpBatchNum := secpShareNum / BATCHSIZE
			if ceil {
				secpBatchNum += 1
			}

			// get total assigned share num for c25519
			c25519ShareNumUint, err := service.broker.ABCIMethods().LastC25519UnassignedIndex()
			if err != nil {
				log.Errorf("Could not get share number %s", err.Error())
				return nil, err
			}
			c25519ShareNum := int(c25519ShareNumUint)
			// calculate batch needed
			ceil = (c25519ShareNum % BATCHSIZE) != 0
			c25519BatchNum := c25519ShareNum / BATCHSIZE
			if ceil {
				c25519BatchNum += 1
			}

			// store the batch info in PssService
			service.secpBatchNum = int(secpBatchNum)
			service.c25519BatchNum = int(c25519BatchNum)
			service.secpShareNum = int(secpShareNum)
			service.c25519ShareNum = int(c25519ShareNum)
			service.newEpochInfo = newEpochInfo

			// start the first batch
			service.StartNextPSSBatch()
		}
	}
	return nil, nil
}

// StartNextPSSBatch starts the next batch of PSS
// based on the current batch saved in PssService
// This function should be called by the DPSS
// message handlers to start the next batch
func (service *PssService) StartNextPSSBatch() {
	if service.currentSecpBatch <= service.secpBatchNum {
		// secp256K1 key shares
		// FIXME oldShares needs to be accompanied by the userId
		var oldShares []sharing.ShamirShare
		id := service.broker.ChainMethods().GetSelfIndex()
		// get old shares list of the batch
		for i := 0; i < BATCHSIZE; i++ {
			index := service.currentSecpBatch*BATCHSIZE + i
			if index >= service.secpShareNum {
				log.WithFields(log.Fields{
					"type":            "secp256k1",
					"last index":      index - 1,
					"total share num": service.secpShareNum,
				}).Debug("Last share added")
				break
			}
			si, _, err := service.broker.DBMethods().RetrieveCompletedShare(*big.NewInt(int64(index)), common.SECP256K1)
			if err != nil {
				log.Errorf("unable to ertrieve secp256k1 share of index %v: %s", index, err.Error())
			}
			share := sharing.ShamirShare{
				Id:    uint32(id),
				Value: si.Bytes(),
			}
			oldShares = append(oldShares, share)
		}

		if len(oldShares) > 0 {
			log.WithFields(log.Fields{
				"type":  "secp256k1",
				"batch": service.currentSecpBatch,
			}).Info("Running DPSS")

			// FIXME placeholder
			roundDetails := common.PSSRoundDetails{
				PssID:     common.NewPssID(*big.NewInt(int64(service.currentSecpBatch))),
				Dealer:    service.pssNode.NodeDetails,
				BatchSize: BATCHSIZE,
			}

			// FIXME replace with `oldShares` once it has the right type
			// TODO: Fill this array.
			shares := make([]common.PrivKeyShare, len(oldShares))

			// FIXME placeholder
			ephemeralKeypair := common.GenerateKeyPair(common.CurveFromName(common.SECP256K1))

			// FIXME how do we know the new comittee params
			createdMsgBytes, err := dacss.NewInitMessage(
				roundDetails,
				shares,
				common.SECP256K1,
				ephemeralKeypair,
				common.CommitteeParams{N: int(service.newEpochInfo.N.Int64()), K: int(service.newEpochInfo.K.Int64()), T: int(service.newEpochInfo.T.Int64())},
			)

			if err != nil {
				// TODO
			}

			service.pssNode.PssNodeTransport.Receive(service.pssNode.NodeDetails, *createdMsgBytes)

			log.WithFields(log.Fields{
				"type":  "secp256k1",
				"batch": service.currentSecpBatch,
			}).Info("DPSS finished")

			service.currentSecpBatch++

		}
	} else if service.currentC255Batch <= service.c25519BatchNum {
		//c25519 key shares
		var oldShares []sharing.ShamirShare
		// get old shares list of the batch
		for i := 0; i < BATCHSIZE; i++ {
			index := service.currentC255Batch*BATCHSIZE + i
			if index >= service.c25519ShareNum {
				log.WithFields(log.Fields{
					"type":            "ed25519",
					"last index":      index - 1,
					"total share num": service.c25519ShareNum,
				}).Debug("Last share added")
				break
			}
			si, _, err := service.broker.DBMethods().RetrieveCompletedShare(*big.NewInt(int64(index)), common.ED25519)
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
				"batch": service.currentC255Batch,
			}).Info("Running DPSS")
			// Todo: what message to send here?
			// dacss.NewInitMessage(,oldoldShares,)
			log.WithFields(log.Fields{
				"type":  "ed25519",
				"batch": service.currentC255Batch,
			}).Info("DPSS finished")

			service.currentSecpBatch++
		}

	}
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
