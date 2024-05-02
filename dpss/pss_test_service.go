//go:build test

package dpss

import (
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/dacss"
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

var secpShareNum uint = 10   //hardcoded secp256k1 share number
var c25519ShareNum uint = 10 //hardcoded ed25519 share number

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
		batchSize := uint(5)
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
		var epochOfNode int
		if isNewCommittee {
			epochOfNode = newEpoch
		} else {
			epochOfNode = oldEpoch
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
			epochOfNode)
		if err != nil {
			log.Errorf("Could not create pssNode in trigger_pss %s", err.Error())
			return nil, err
		}
		service.SetNode(pssNode)

		// calculate batch needed
		ceil := (secpShareNum % batchSize) != 0
		secpBatchNum := secpShareNum / batchSize
		if ceil {
			secpBatchNum += 1
		}

		// calculate batch needed
		ceil = (c25519ShareNum % batchSize) != 0
		c25519BatchNum := c25519ShareNum / batchSize
		if ceil {
			c25519BatchNum += 1
		}

		//TODO - check if we need this
		// To make sure honest nodes have finished creating PssNode
		time.Sleep(10 * time.Second)

		if !isNewCommittee {
			// only nodes in old committee need to initiate DPSS
			go service.BatchRunDPSS(secpBatchNum,
				c25519BatchNum,
				batchSize,
				secpShareNum,
				c25519ShareNum,
				int(newEpochInfo.N.Int64()),
				int(newEpochInfo.K.Int64()),
				int(newEpochInfo.T.Int64()))
		}
	}
	return nil, nil
}

func (service *PssService) BatchRunDPSS(secpBatchNum uint, c25519BatchNum uint, batchSize uint, secpShareNum uint, c25519ShareNum uint,
	new_N int, new_K int, new_T int) {

	// TODO make more generalized and call for both secp and c25519
	// as of now only in secp256k1 the message handlers are triggered

	id := service.broker.ChainMethods().GetSelfIndex()
	// set the seed so we can recreate the random secret in the new committee
	secpRand := rand.New(rand.NewSource(0))
	// secp256k1 shares
	for currentBatch := uint(0); currentBatch < secpBatchNum; currentBatch++ {
		// FIXME oldShares needs to be accompanied by the userId
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

			// for benchmark
			// create random scalar as the secret share
			// this method is better for benchmarking since it is faster
			// however it isn't good for testing since it will be hard to interpolate and compare the secret
			// si := curves.K256().Scalar.Random(secpRand).BigInt()
			// share := sharing.ShamirShare{
			// 	Id:    uint32(id),
			// 	Value: si.Bytes(),
			// }

			// for testing
			// this method is slower since we have to run shamir sharing everytime
			// but we can easily get the secret in the new committee
			// by regenerate the random values with the correct seed
			shamir, _ := sharing.NewShamir(uint32(service.pssNode.OldCommitteeNodes.K), uint32(service.pssNode.OldCommitteeNodes.N), curves.K256())
			secret := curves.K256().Scalar.Random(secpRand)
			fmt.Printf("secp %v", secret)
			shares, _ := shamir.Split(secret, secpRand)
			share := *shares[id-1]

			oldShares = append(oldShares, share)
		}

		if len(oldShares) > 0 {
			log.WithFields(log.Fields{
				"type":  "secp256k1",
				"batch": currentBatch,
			}).Info("Running DPSS")

			// FIXME placeholder
			roundDetails := common.PSSRoundDetails{
				PssID:  common.NewPssID(*big.NewInt(int64(currentBatch))),
				Dealer: service.pssNode.NodeDetails,
			}

			// FIXME replace with `oldShares` once it has the right type
			shares := make([]common.PrivKeyShare, len(oldShares))

			// FIXME placeholder
			ephemeralKeypair := common.GenerateKeyPair(common.CurveFromName(common.SECP256K1))

			// FIXME how do we know the new comittee params
			createdMsgBytes, err := dacss.NewInitMessage(
				roundDetails,
				shares,
				common.SECP256K1,
				ephemeralKeypair,
				common.CommitteeParams{N: new_N, K: new_K, T: new_T},
			)

			if err != nil {
				// TODO
				log.Errorf("Couldn't create dacss Init message %s", err.Error())
			}

			go service.pssNode.PssNodeTransport.Receive(service.pssNode.NodeDetails, *createdMsgBytes)

			// block until the batch has finished
			//<-service.batchFinChannel
			log.WithFields(log.Fields{
				"type":  "secp256k1",
				"batch": currentBatch,
			}).Info("DPSS finished")
		}

	}

	// set the seed so we can recreate the random secret in the new committee
	ed25519Rand := rand.New(rand.NewSource(1))

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

			// for benchmark
			// create random scalar as the secret share
			// this method is for benchmarking since it is faster
			// but it isn't good for testing since it will be hard to interpolate the secret
			// si := curves.ED25519().Scalar.Random(ed25519Rand).BigInt()
			// share := sharing.ShamirShare{
			// 	Id:    uint32(id),
			// 	Value: si.Bytes(),
			// }

			// for testing
			// this method is slower since we have to run shamir sharing everytime
			// but we can easily get the secret in the new committee
			// by regenerate the random values with the correct seed
			shamir, _ := sharing.NewShamir(uint32(service.pssNode.OldCommitteeNodes.K), uint32(service.pssNode.OldCommitteeNodes.N), curves.ED25519())
			secret := curves.ED25519().Scalar.Random(ed25519Rand)
			fmt.Printf("ed25519 %v", secret)
			shares, _ := shamir.Split(secret, ed25519Rand)
			share := *shares[id-1]

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

// func get_next_share(k, n uint32, curve *curves.Curve) {
// 	shamir, _ := sharing.NewShamir(k, n, curve)
// 	shamir.
// }

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
