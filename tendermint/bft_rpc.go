package tendermint

import (
	"context"
	"crypto/rand"
	"fmt"
	"reflect"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/secp256k1"
	"github.com/arcana-network/dkgnode/tendermint/messageq"

	"github.com/arcana-network/dkgnode/eventbus"
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/torusresearch/bijson"
)

type BFTRPC struct {
	client.Client
	broker      *common.MessageBroker
	BftMsgQueue *messageq.MessageQueue
}

type BFTTxWrapper interface {
	PrepareBFTTx() ([]byte, error)
	DecodeBFTTx([]byte) error
	GetSerializedBody() []byte
}

type DefaultBFTTxWrapper struct {
	BFTTx     []byte       `json:"bft_tx,omitempty"`
	Nonce     uint32       `json:"nonce,omitempty"`
	PubKey    common.Point `json:"pub_key,omitempty"`
	MsgType   byte         `json:"msg_type,omitempty"`
	Signature []byte       `json:"signature,omitempty"`
}

type AssignmentTx struct {
	Provider string
	UserID   string
	AppID    string
}

// mapping of name of struct to id
var txTypeMap = map[string]byte{
	getType(AssignmentTx{}):      byte(1),
	getType(common.DKGMessage{}): byte(2),
}

func (wrapper *DefaultBFTTxWrapper) PrepareBFTTx(bftTx interface{}, broker *common.MessageBroker) ([]byte, error) {
	// type byte
	msgType, ok := txTypeMap[getType(bftTx)]
	if !ok {
		return nil, fmt.Errorf("msg type does not exist for BFT: %s ", getType(bftTx))
	}
	wrapper.MsgType = msgType
	nonce, err := rand.Int(rand.Reader, secp256k1.GeneratorOrder)
	if err != nil {
		return nil, fmt.Errorf("could not generate random number")
	}
	wrapper.Nonce = uint32(nonce.Int64())
	pk := broker.ChainMethods().GetSelfPublicKey()
	wrapper.PubKey.X = pk.X
	wrapper.PubKey.Y = pk.Y
	bftRaw, err := bijson.Marshal(bftTx)
	if err != nil {
		return nil, err
	}
	wrapper.BFTTx = bftRaw

	// sign message data
	data := wrapper.GetSerializedBody()
	wrapper.Signature = broker.ChainMethods().SelfSignData(data)

	rawMsg, err := bijson.Marshal(wrapper)
	if err != nil {
		return nil, err
	}
	return rawMsg, nil
}

func (wrapper *DefaultBFTTxWrapper) DecodeBFTTx(data []byte) error {
	err := bijson.Unmarshal(data, &wrapper.BFTTx)
	if err != nil {
		return err
	}
	return nil
}

func (wrapper DefaultBFTTxWrapper) GetSerializedBody() []byte {
	wrapper.Signature = nil
	bin, err := bijson.Marshal(wrapper)
	if err != nil {
		log.Errorf("could not GetSerializedBody bfttx, %v", err)
	}
	return bin
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

// BroadcastTxSync Wrapper (input should be fastjsoned) to tendermint.
// blocks until message is sent
func (bftrpc BFTRPC) Broadcast(bftTx interface{}) (*common.Hash, error) {

	var wrapper DefaultBFTTxWrapper
	preparedTx, err := wrapper.PrepareBFTTx(bftTx, bftrpc.broker)
	if err != nil {
		return nil, err
	}

	res := bftrpc.BftMsgQueue.Add(preparedTx)
	log.WithField("Tx", bftTx).Info("BFTRPC:Broadcast()")

	response, ok := res.(*tmtypes.ResultBroadcastTx)
	if !ok {
		return nil, fmt.Errorf("return type for broadcast was not *tmtypes.ResultBroadcastTx: %v", res)
	}
	log.Debugf("TENDERBFT RESPONSE code %v : %v", response.Code, response)
	log.Debugf("TENDERBFT LOG: %s", response.Log)

	if response.Code != 0 {
		log.Errorf("BFTBroadcast():responseCode=%d, log=%s", response.Code, response.Log)
		return nil, fmt.Errorf("Could not broadcast, ErrorCode: %d", response.Code)
	}

	return &common.Hash{HexBytes: response.Hash.Bytes()}, nil
}

// Retrieves tx from the bft and gives back results.
func (bftrpc BFTRPC) Retrieve(hash []byte, txStruct BFTTxWrapper) (err error) {
	result, err := bftrpc.Tx(context.Background(), hash, false)
	if err != nil {
		return err
	}
	if result.TxResult.Code != 0 {
		log.Debugf("Transaction not accepted %v", result.TxResult.Code)
	}

	err = (txStruct).DecodeBFTTx(result.Tx)
	if err != nil {
		return err
	}

	return nil
}

func NewBFTRPC(tmclient client.Client, bus eventbus.Bus) *BFTRPC {
	var rpc BFTRPC
	rpc.Client = tmclient
	rpc.BftMsgQueue = messageq.NewMessageQueue(rpc.messageRunFunc)
	rpc.BftMsgQueue.RunMsgEngine(50)
	rpc.broker = common.NewServiceBroker(bus, "bft-rpc")
	return &rpc
}

func (bftrpc *BFTRPC) messageRunFunc(msg []byte) (interface{}, error) {
	response, err := bftrpc.BroadcastTxSync(context.Background(), msg)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Retrying submission of bfttx... posssible pressure/overloaded tm")
		return nil, err
	}
	return response, nil
}

type Semaphore struct {
	sem chan struct{}
}

type ReleasedState struct {
	Released bool
}

// ReleaseFunc - releases after acquire
type ReleaseFunc func()

func (s *Semaphore) Acquire() ReleaseFunc {
	log.Info("acquiredSemaphores")
	s.sem <- struct{}{}
	return func() {
		log.Info("ReleasedSemaphores")
		<-s.sem
	}
}

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{
		sem: make(chan struct{}, n),
	}
}
