package common

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/avast/retry-go"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	log "github.com/sirupsen/logrus"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmp2p "github.com/tendermint/tendermint/p2p"
	"github.com/torusresearch/bijson"

	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/bytes"

	"github.com/arcana-network/dkgnode/secp256k1"
)

type MessageBroker struct {
	bus    eventbus.Bus
	caller string
}

type KeygenNodeDetails struct {
	Index  int
	PubKey Point
}

type VerifierData struct {
	Ok                   bool
	Verifier, VerifierID string
	KeyIndexes           []big.Int
	Err                  error
}

type Iterator struct {
	RandomID string
	CallNext func() bool
	Val      interface{}
	Err      error
}

func (iterator *Iterator) Next() bool {
	return iterator.CallNext()
}
func (iterator *Iterator) Value() (val interface{}) {
	return iterator.Val
}

type EpochInfo struct {
	Id        big.Int
	N         big.Int
	K         big.Int
	T         big.Int
	PrevEpoch big.Int
	NextEpoch big.Int
}

func (n *KeygenNodeDetails) FromNodeDetailsID(nodeDetailsID NodeDetailsID) {
	s := string(nodeDetailsID)
	substrings := strings.Split(s, Delimiter1)

	if len(substrings) != 3 {
		return
	}
	index, err := strconv.Atoi(substrings[0])
	if err != nil {
		return
	}
	n.Index = index
	pubkeyX, ok := new(big.Int).SetString(substrings[1], 16)
	if !ok {
		return
	}
	n.PubKey.X = *pubkeyX
	pubkeyY, ok := new(big.Int).SetString(substrings[2], 16)
	if !ok {
		return
	}
	n.PubKey.Y = *pubkeyY
}
func NewServiceBroker(bus eventbus.Bus, caller string) *MessageBroker {
	return &MessageBroker{
		bus:    bus,
		caller: caller,
	}
}

func (broker *MessageBroker) ChainMethods() *ChainMethods {
	return &ChainMethods{
		bus:     broker.bus,
		caller:  broker.caller,
		service: CHAIN_SERVICE_NAME,
	}
}
func (broker *MessageBroker) ServerMethods() *ServerMethods {
	return &ServerMethods{
		bus:     broker.bus,
		caller:  broker.caller,
		service: SERVER_SERVICE_NAME,
	}
}

func (broker *MessageBroker) DBMethods() *DBMethods {
	return &DBMethods{
		bus:     broker.bus,
		caller:  broker.caller,
		service: DB_SERVICE_NAME,
	}
}
func (broker *MessageBroker) KeygenMethods() *KeygenMethods {
	return &KeygenMethods{
		bus:     broker.bus,
		caller:  broker.caller,
		service: KEYGEN_SERVICE_NAME,
	}
}

func (broker *MessageBroker) TendermintMethods() *TendermintMethods {
	return &TendermintMethods{
		bus:     broker.bus,
		caller:  broker.caller,
		service: TENDERMINT_SERVICE_NAME,
	}
}
func (broker *MessageBroker) ABCIMethods() *ABCIMethods {
	return &ABCIMethods{
		bus:     broker.bus,
		caller:  broker.caller,
		service: ABCI_SERVICE_NAME,
	}
}

func (broker *MessageBroker) CacheMethods() *CacheMethods {
	return &CacheMethods{
		bus:     broker.bus,
		caller:  broker.caller,
		service: CACHE_SERVICE_NAME,
	}
}

func (broker *MessageBroker) VerifierMethods() *VerifierMethods {
	return &VerifierMethods{
		bus:     broker.bus,
		caller:  broker.caller,
		service: VERIFIER_SERVICE_NAME,
	}
}
func (broker *MessageBroker) KeystoreMethods() *KeystoreMethods {
	return &KeystoreMethods{
		bus:     broker.bus,
		caller:  broker.caller,
		service: KEYSTORE_SERVICE_NAME,
	}
}

type KeystoreMethods struct {
	bus     eventbus.Bus
	caller  string
	service string
}

func (ksm *KeystoreMethods) StoreShare(id string, share []byte) (err error) {
	methodResponse := ServiceMethod(ksm.bus, ksm.caller, ksm.service, "store", id, share)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	if err != nil {
		return err
	}
	return
}

func (ksm *KeystoreMethods) RetrieveShare(id string) (share []byte, err error) {
	methodResponse := ServiceMethod(ksm.bus, ksm.caller, ksm.service, "retrieve", id)
	if methodResponse.Error != nil {
		return share, methodResponse.Error
	}
	var data []byte
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return data, err
	}
	share = data
	return
}

type ServerMethods struct {
	bus     eventbus.Bus
	caller  string
	service string
}

func (sm *ServerMethods) RequestConnectionDetails(endpoint string) (connectionDetails ConnectionDetails, err error) {
	methodResponse := ServiceMethod(sm.bus, sm.caller, sm.service, "request_connection_details", endpoint)
	if methodResponse.Error != nil {
		return connectionDetails, methodResponse.Error
	}
	var data ConnectionDetails
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return connectionDetails, err
	}
	connectionDetails = data
	return
}

type VerifierMethods struct {
	bus     eventbus.Bus
	caller  string
	service string
}

func (vm *VerifierMethods) Verify(rawMessage *bijson.RawMessage) (valid bool, userID string, err error) {
	methodResponse := ServiceMethod(vm.bus, vm.caller, vm.service, "verify", rawMessage)
	if methodResponse.Error != nil {
		err = methodResponse.Error
		return
	}
	var data struct {
		Valid  bool
		UserID string
	}
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return valid, userID, err
	}
	valid = data.Valid
	userID = data.UserID
	return
}

func (vm *VerifierMethods) CleanToken(verifierIdentifier string, idtoken string) (cleanedToken string, err error) {
	methodResponse := ServiceMethod(vm.bus, vm.caller, vm.service, "clean_token", verifierIdentifier, idtoken)
	if methodResponse.Error != nil {
		err = methodResponse.Error
		return
	}
	var data string
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return cleanedToken, err
	}
	cleanedToken = data
	return
}

type ABCIMethods struct {
	bus     eventbus.Bus
	caller  string
	service string
}

type KeyAssignmentPublic struct {
	Index     big.Int
	PublicKey Point
	Verifiers map[string][]string // Verifier => VerifierID
}

func (am *ABCIMethods) RetrieveKeyMapping(keyIndex big.Int) (keyDetails KeyAssignmentPublic, err error) {
	methodResponse := ServiceMethod(am.bus, am.caller, am.service, "retrieve_key_mapping", keyIndex)
	if methodResponse.Error != nil {
		return keyDetails, methodResponse.Error
	}
	var data KeyAssignmentPublic
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return keyDetails, err
	}
	keyDetails = data
	return
}

func (am *ABCIMethods) GetIndexesFromVerifierID(verifier, verifierID, appID string) (keyIndexes []big.Int, err error) {
	methodResponse := ServiceMethod(am.bus, am.caller, am.service, "get_indexes_from_verifier_id", verifier, verifierID, appID)
	if methodResponse.Error != nil {
		return keyIndexes, methodResponse.Error
	}
	var data []big.Int
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return keyIndexes, err
	}
	keyIndexes = data
	return
}

type ChainMethods struct {
	bus     eventbus.Bus
	caller  string
	service string
}

func (cm *ChainMethods) GetTMP2PConnection() (tmp2pconnection string) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_tm_p2p_connection")
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data string
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		tmp2pconnection = data
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get node details by epoch and index")
	}
	return
}
func (cm *ChainMethods) GetP2PConnection() (p2pconnection string) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_p2p_connection")
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data string
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		p2pconnection = data
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get node details by epoch and index")
	}
	return
}

func (cm *ChainMethods) ValidateEpochPubKey(nodeAddress ethCommon.Address, pubK Point) (valid bool) {
	methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service,
		"validate_epoch_pub_key", nodeAddress, pubK)

	if methodResponse.Error != nil {
		return false
	}
	var data bool
	err := CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return false
	}
	return data
}

func (cm *ChainMethods) GetSelfAddress() (address ethCommon.Address) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_self_address")
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data ethCommon.Address
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		address = data
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get self address")
	}
	return
}

func (cm *ChainMethods) AwaitNodesConnected(epoch int) {
	methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "await_nodes_connected", epoch)
	if methodResponse.Error != nil {
		log.WithError(methodResponse.Error).Fatal("await nodes connected returned error")
	}
}

func (cm *ChainMethods) KeyBuffer() int {
	methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_key_buffer")
	if methodResponse.Error != nil {
		log.WithError(methodResponse.Error).Info("could not get key buffer")
	}
	var data int
	err := CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		log.WithError(err).Error("could not castOrUnmarshal data")
	}
	return data
}

func (cm *ChainMethods) SetSelfIndex(index int) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "set_self_index", index)
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not set self index")
	}
}

func (cm *ChainMethods) GetPreviousEpoch() (epoch int, err error) {
	methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_previous_epoch", nil)
	if methodResponse.Error != nil {
		return 0, methodResponse.Error
	}
	var data int
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		log.WithError(err).Error("could not castOrUnmarshal data")
		return
	}
	epoch = data
	return
}
func (cm *ChainMethods) GetNextEpoch() (epoch int, err error) {
	methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_next_epoch")
	if methodResponse.Error != nil {
		return 0, methodResponse.Error
	}
	var data int
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		log.WithError(err).Error("could not castOrUnmarshal data")
		return
	}
	epoch = data
	return
}

func (cm *ChainMethods) GetCurrentEpoch() (epoch int) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_current_epoch")
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data int
		log.WithField("GetCurentEpoch", methodResponse.Data).Debug("ServiceMapper")
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		if data == 0 {
			return errors.New("could not get current epoch")
		}
		epoch = data
		return nil
	})
	if err != nil || epoch == 0 {
		log.WithError(err).Fatal("could not get current epoch")
	}
	return
}

func (cm *ChainMethods) GetEpochInfo(epoch int, skipCache bool) (eInfo EpochInfo, err error) {
	methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_epoch_info", epoch, skipCache)
	if methodResponse.Error != nil {
		return eInfo, methodResponse.Error
	}
	var data EpochInfo
	log.WithField("get_curent_epoch_info", methodResponse.Data).Info("ServiceMapper")
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		log.WithError(err).Error("could not castOrUnmarshal data")
		return
	}
	if data.Id.Cmp(big.NewInt(0)) == 0 {
		return eInfo, errors.New("data is invalid, epochID is 0")
	}
	eInfo = data
	return
}

func (cm *ChainMethods) GetNodeDetailsByEpochAndIndex(epoch int, index int) (nodeRef NodeReference) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_node_details_by_epoch_and_index", epoch, index)
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data SerializedNodeReference
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		nodeRef = NodeReference{}.Deserialize(data)
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get node details by epoch and index")
	}
	return
}

func (cm *ChainMethods) GetNodeDetailsByAddress(address ethCommon.Address) (nodeRef NodeReference) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_node_details_by_address", address)
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data SerializedNodeReference
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		nodeRef = NodeReference{}.Deserialize(data)
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get node details by address")
	}
	return
}
func (cm *ChainMethods) GetSelfIndex() (index int) {
	methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_self_index")
	if methodResponse.Error != nil {
		log.Fatalf("Get self index returned error, should be blocking until success")
	}
	var data int
	err := CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		log.WithError(err).Error("could not castOrUnmarshal data")
	}
	index = data
	return
}
func (cm *ChainMethods) GetSelfPublicKey() (pubKey Point) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_self_public_key")
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data Point
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			log.Error(err)
			return err
		}
		pubKey = data
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get self public key")
	}
	return
}
func (cm *ChainMethods) GetSelfPrivateKey() (privKey big.Int) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_self_private_key")
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data big.Int
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		privKey = data
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get self private key")
	}
	return
}
func (cm *ChainMethods) VerifyDataWithNodelist(pk Point, sig []byte, input []byte) (senderDetails KeygenNodeDetails, err error) {
	methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "verify_data_with_nodelist", pk, sig, input)
	var data KeygenNodeDetails
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return senderDetails, err
	}
	return data, nil
}

func (cm *ChainMethods) Connect() error {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "connect_to_nodes")
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get node list")
		return err
	}
	return nil
}
func (cm *ChainMethods) GetNodeList(epoch int) (nodeRefs []NodeReference) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_node_list", epoch)
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data []SerializedNodeReference
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		var deserializedData []NodeReference
		for i := 0; i < len(data); i++ {
			deserializedData = append(deserializedData, NodeReference{}.Deserialize(data[i]))
		}
		nodeRefs = deserializedData
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get node list")
	}
	return
}

func (cm *ChainMethods) VerifyDataWithEpoch(pk Point, sig []byte, input []byte, epoch int) (senderDetails KeygenNodeDetails, err error) {
	methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "verify_data_with_epoch", pk, sig, input, epoch)
	var data KeygenNodeDetails
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return senderDetails, err
	}
	return data, nil
}
func (cm *ChainMethods) GetClientIDViaVerifier(appID, verifier string) (*VerifierParams, error) {
	methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_params_by_verifier", appID, verifier)
	var params VerifierParams
	err := CastOrUnmarshal(methodResponse.Data, &params)
	if err != nil {
		return nil, err
	}
	return &params, err
}

func (cm *ChainMethods) GetPartitionForApp(appID string) (partitioned bool, err error) {
	methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "get_app_partition", appID)
	err = CastOrUnmarshal(methodResponse.Data, &partitioned)
	if err != nil {
		return
	}
	return
}

func (cm *ChainMethods) AwaitCompleteNodeList(epoch int) (nodeRefs []NodeReference) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "await_complete_node_list", epoch)
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data []SerializedNodeReference
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		var deserializedData []NodeReference
		for i := 0; i < len(data); i++ {
			deserializedData = append(deserializedData, NodeReference{}.Deserialize(data[i]))
		}
		nodeRefs = deserializedData
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get complete node list")
	}
	return
}

func (cm *ChainMethods) SelfSignData(input []byte) (rawSig []byte) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cm.bus, cm.caller, cm.service, "self_sign_data", input)
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data []byte
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		rawSig = data
		return nil
	})
	if err != nil {
		log.Fatalf("Could not set self sign data, %v", err.Error())
	}
	return
}

type MethodRequest struct {
	Caller  string
	Service string
	Method  string
	ID      string
	Data    []interface{}
}

type MethodResponse struct {
	Request MethodRequest
	Error   error
	Data    interface{}
}

func ServiceMethod(eventBus eventbus.Bus, caller string, service string, method string, data ...interface{}) MethodResponse {
	nonce, err := rand.Int(rand.Reader, secp256k1.GeneratorOrder)
	if err != nil {
		return MethodResponse{
			Error: errors.New("could not generate random nonce"),
			Data:  nil,
		}
	}
	nonceStr := nonce.Text(16)
	nonce = nil
	responseCh := AwaitTopic(eventBus, nonceStr)
	eventBus.Publish("method", MethodRequest{
		Caller:  caller,
		Service: service,
		Method:  method,
		ID:      nonceStr,
		Data:    data,
	})
	methodResponseInter := <-responseCh
	methodResponse, ok := methodResponseInter.(MethodResponse)
	if !ok {
		return MethodResponse{
			Error: errors.New("method response was not of MethodResponse type"),
			Data:  nil,
		}
	}
	return methodResponse
}

func CastOrUnmarshal(dataInter interface{}, v interface{}, flags ...bool) (err error) {
	var silent bool
	if len(flags) >= 1 && flags[0] {
		silent = true
	}
	defer func() {
		if r := recover(); r != nil {
			if !silent {
				log.WithField("recover", r).WithField("stack", string(debug.Stack())).Info("could not cast in castOrUnmarshal")
			}
			err = errors.New("could not cast in castOrUnmarshal")
		}
	}()

	data, ok := dataInter.(EventBusBytes)
	if ok {
		err = bijson.Unmarshal(data, v)
		if err != nil {
			log.WithField("data", data).WithError(err).Info("could not unmarshal in castOrUnmarshal")
		}
	} else {
		lhs := reflect.ValueOf(dataInter)
		rhs := reflect.ValueOf(v)
		if lhs.Kind() == reflect.Ptr {
			el := lhs.Elem()
			if !el.IsValid() {
				log.Printf("LHS: %#v, RHS: %#v\n", dataInter, v)
				return errors.New("LHS' element is invalid and may not be casted to the RHS")
			}

			rhs.Elem().Set(el)
		} else {
			rhs.Elem().Set(lhs)
		}
	}
	return
}

func AwaitTopic(eventBus eventbus.Bus, topic string) <-chan interface{} {
	responseCh := make(chan interface{})
	err := eventBus.SubscribeOnceAsync(topic, func(res interface{}) {
		responseCh <- res
		close(responseCh)
	})
	if err != nil {
		log.WithError(err).Error("could not subscribe async")
	}
	return responseCh
}

type P2PMethods struct {
	bus     eventbus.Bus
	caller  string
	service string
}

func (broker *MessageBroker) P2PMethods() *P2PMethods {
	return &P2PMethods{
		bus:     broker.bus,
		caller:  broker.caller,
		service: P2P_SERVICE_NAME,
	}
}

func (pm *P2PMethods) ID() (peerID peer.ID) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(pm.bus, pm.caller, pm.service, "id")
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data peer.ID
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return fmt.Errorf("could not castOrUnmarshal %v", err.Error())
		}
		peerID = data
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get id")
	}
	return
}
func (pm *P2PMethods) AuthenticateMessage(p2pBasicMsg P2PBasicMsg) (err error) {
	methodResponse := ServiceMethod(pm.bus, pm.caller, pm.service, "authenticate_message", p2pBasicMsg)
	return methodResponse.Error
}

func (pm *P2PMethods) NewP2PMessage(messageId string, gossip bool, payload []byte, msgType string) (newMsg P2PBasicMessage) {
	methodResponse := ServiceMethod(pm.bus, pm.caller, pm.service, "new_p2p_message", messageId, gossip, payload, msgType)
	if methodResponse.Error != nil {
		log.WithError(methodResponse.Error).Fatal()
	}
	var data P2PBasicMessage
	err := CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		log.WithField("Data", methodResponse.Data).Error("could not castOrUnmarshal")
	}
	newMsg = data
	return
}

func (pm *P2PMethods) ConnectToP2PNode(nodeP2PConnection string, nodePeerID peer.ID) error {
	methodResponse := ServiceMethod(pm.bus, pm.caller, pm.service, "connect_to_p2p_node", nodeP2PConnection, nodePeerID)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}

func (pm *P2PMethods) SignP2PMessage(message P2PMessage) (signature []byte, err error) {
	methodResponse := ServiceMethod(pm.bus, pm.caller, pm.service, "sign_p2p_message", message)
	if methodResponse.Error != nil {
		return signature, methodResponse.Error
	}
	var data []byte
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return signature, fmt.Errorf("could not castOrUnmarshal data %v %v", methodResponse.Data, err.Error())
	}
	return data, nil
}
func (pm *P2PMethods) SendP2PMessage(id peer.ID, p protocol.ID, msg P2PMessage) error {
	methodResponse := ServiceMethod(pm.bus, pm.caller, pm.service, "send_p2p_message", id, p, msg)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}
func (pm *P2PMethods) SetStreamHandler(proto string, handler func(StreamMessage)) error {
	eventBus := pm.bus
	if eventBus.HasCallback("p2p:forward:" + proto) {
		return fmt.Errorf("cannot call setStreamHandler on proto %v as it already has a handler", proto)
	}
	err := eventBus.SubscribeAsync("p2p:forward:"+proto, func(inter interface{}) {
		var data P2PBasicMessage
		err := CastOrUnmarshal(inter, &data)
		if err != nil {
			log.WithField("inter", inter).Error("could not castOrUnmarshal")
		}
		log.WithFields(log.Fields{
			"protocol": proto,
		}).Debug("received p2p")
		handler(StreamMessage{
			Protocol: proto,
			Message:  data,
		})
	}, false)
	if err != nil {
		return err
	}
	methodResponse := ServiceMethod(pm.bus, pm.caller, pm.service, "set_stream_handler", proto)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}

func (pm *P2PMethods) GetHostAddress() (hostAddress string) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(pm.bus, pm.caller, pm.service, "get_host_address")
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data string
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		hostAddress = data
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get host address")
	}
	return
}

type DBMethods struct {
	bus     eventbus.Bus
	caller  string
	service string
}

func (dbm *DBMethods) RetrieveNodePubKey(nodeAddress ethCommon.Address) (pubKey Point, err error) {
	methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "retrieve_node_pub_key", nodeAddress)
	if methodResponse.Error != nil {
		return pubKey, nil
	}
	var data Point
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return pubKey, err
	}
	return data, nil
}

func (dbm *DBMethods) StoreConnectionDetails(nodeAddress ethCommon.Address, connectionDetails ConnectionDetails) error {
	methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "store_connection_details", nodeAddress, connectionDetails)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}

func (dbm *DBMethods) StoreNodePubKey(nodeAddress ethCommon.Address, pubKey Point) error {
	methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "store_node_pub_key", nodeAddress, pubKey)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}

func (dbm *DBMethods) RetrieveConnectionDetails(nodeAddress ethCommon.Address) (connectionDetails ConnectionDetails, err error) {
	methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "retrieve_connection_details", nodeAddress)
	if methodResponse.Error != nil {
		return connectionDetails, methodResponse.Error
	}
	var data ConnectionDetails
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return connectionDetails, err
	}
	return data, nil
}

func (dbm *DBMethods) StorePSSCommitmentMatrix(keyIndex big.Int, c [][]Point) error {
	methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "store_PSS_commitment_matrix", keyIndex, c)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}

func (dbm *DBMethods) RetrieveCommitmentMatrix(keyIndex big.Int) (c [][]Point, err error) {
	methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "retrieve_commitment_matrix", keyIndex)
	if methodResponse.Error != nil {
		return c, methodResponse.Error
	}
	var data [][]Point
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return c, err
	}
	return data, nil
}
func (dbm *DBMethods) RetrieveCompletedShare(keyIndex big.Int) (Si big.Int, Siprime big.Int, err error) {
	methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "retrieve_completed_share", keyIndex)
	if methodResponse.Error != nil {
		err = methodResponse.Error
		return
	}
	var data struct {
		Si      big.Int
		Siprime big.Int
	}
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return Si, Siprime, err
	}
	Si = data.Si
	Siprime = data.Siprime
	return
}
func (dbm *DBMethods) SetKeygenStarted(keygenID string, started bool) error {
	methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "set_keygen_started", keygenID, started)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}

func (dbm *DBMethods) StoreCompletedPSSShare(keyIndex big.Int, si big.Int, siprime big.Int) error {
	methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "store_completed_PSS_share", keyIndex, si, siprime)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}

func (dbm *DBMethods) StoreCommitment(keyIndex big.Int, T []int, metadata map[string][]Point) error {
	methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "store_sharing_commitment", keyIndex, T, metadata)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}

func (dbm *DBMethods) StorePublicKeyToIndex(publicKey Point, keyIndex big.Int) error {
	methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "store_public_key_to_index", publicKey, keyIndex)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}

func (dbm *DBMethods) GetKeygenStarted(keygenID string) (started bool) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "get_keygen_started", keygenID)
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data bool
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		started = data
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get keygen started")
	}
	return
}

func (dbm *DBMethods) IndexToPublicKeyExists(keyIndex big.Int) (exists bool) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "index_to_public_key_exists", keyIndex)
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data bool
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		exists = data
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get index_to_public_key_exists")
	}
	return
}

func (dbm *DBMethods) RetrievePublicKeyToIndex(publicKey Point) (keyIndex big.Int, err error) {
	methodResponse := ServiceMethod(dbm.bus, dbm.caller, dbm.service, "retrieve_public_key_to_index", publicKey)
	if methodResponse.Error != nil {
		return keyIndex, methodResponse.Error
	}
	var data big.Int
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return keyIndex, err
	}
	return data, nil
}

type KeygenMethods struct {
	caller  string
	bus     eventbus.Bus
	service string
}

func (km *KeygenMethods) ReceiveMessage(keygenMessage DKGMessage) error {
	methodResponse := ServiceMethod(km.bus, km.caller, km.service, "receive_message", keygenMessage)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}
func (km *KeygenMethods) Cleanup(id ADKGID) error {
	methodResponse := ServiceMethod(km.bus, km.caller, km.service, "cleanup", id)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}

type TendermintMethods struct {
	caller  string
	bus     eventbus.Bus
	service string
}

type Hash struct {
	bytes.HexBytes
}

func createHandler(respChannel chan []byte) func(inter interface{}) {
	return func(inter interface{}) {
		var methodResponse MethodResponse
		err := CastOrUnmarshal(inter, &methodResponse)
		if err != nil {
			log.WithError(err).WithField("methodResponse", methodResponse).Error("could not castOrUnmarshal")
			return
		}
		if methodResponse.Error != nil {
			log.WithError(methodResponse.Error).Error("could not query tendermint, got error")
			return
		}
		var data []byte
		err = CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			log.WithError(err).WithField("Data", methodResponse.Data).Error("could not castOrUnmarshal")
		}
		respChannel <- data
	}
}
func (tm *TendermintMethods) RegisterQuery(query string) (respChannel chan []byte, err error) {
	respChannel = make(chan []byte)
	eventBus := tm.bus
	if eventBus.HasCallback("tendermint:forward:" + query) {
		err = fmt.Errorf("cannot call RegisterQuery on query %v as it already has a handler", query)
		return
	}
	log.WithField("query", query).Info("RegisterQuery")

	handler := createHandler(respChannel)
	err = eventBus.SubscribeOnceAsync("tendermint:forward:"+query, handler)
	if err != nil {
		return nil, err
	}
	methodResponse := ServiceMethod(eventBus, tm.caller, tm.service, "register_query", query)
	if methodResponse.Error != nil {
		err = methodResponse.Error
	}
	return
}

func (tm *TendermintMethods) DeregisterQuery(query string, handler interface{}) error {
	err := tm.bus.Unsubscribe("tendermint:forward:"+query, handler)
	if err != nil {
		return err
	}
	methodResponse := ServiceMethod(tm.bus, tm.caller, tm.service, "deregister_query", query)
	if methodResponse.Error != nil {
		return methodResponse.Error
	}
	return nil
}

func (tm *TendermintMethods) GetNodeKey() (nodeKey tmp2p.NodeKey) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(tm.bus, tm.caller, tm.service, "get_node_key")
		if methodResponse.Error != nil {
			return methodResponse.Error
		}

		var data []byte
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		newKey := &tmp2p.NodeKey{
			PrivKey: ed25519.PrivKey{},
		}
		err = tmjson.Unmarshal(data, newKey)
		if err != nil {
			return err
		}
		nodeKey = *newKey
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not get nodeKey")
	}
	return
}

func (tm *TendermintMethods) Broadcast(tx interface{}) (txHash Hash, err error) {
	methodResponse := ServiceMethod(tm.bus, tm.caller, tm.service, "broadcast", tx)
	if methodResponse.Error != nil {
		return txHash, methodResponse.Error
	}
	var data Hash
	err = CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return txHash, err
	}
	txHash = data
	return
}
func (tm *TendermintMethods) TxStatus(hash []byte) (bool, error) {
	methodResponse := ServiceMethod(tm.bus, tm.caller, tm.service, "tx_status", hash)
	if methodResponse.Error != nil {
		return false, methodResponse.Error
	}
	var data bool
	err := CastOrUnmarshal(methodResponse.Data, &data)
	if err != nil {
		return false, err
	}
	return data, nil
}

type CacheMethods struct {
	caller  string
	bus     eventbus.Bus
	service string
}

func (cam *CacheMethods) TokenCommitExists(verifier string, tokenCommitment string) (exists bool) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cam.bus, cam.caller, cam.service, "token_commit_exists", verifier, tokenCommitment)
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data bool
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		exists = data
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not check if token commit exists")
	}
	return
}

func (cam *CacheMethods) RecordTokenCommit(verifier string, tokenCommitment string, pubKey Point) {
	methodResponse := ServiceMethod(cam.bus, cam.caller, cam.service, "record_token_commit", verifier, tokenCommitment, pubKey)
	if methodResponse.Error != nil {
		log.WithError(methodResponse.Error).Error("could not record token commit")
	}
}

func (cam *CacheMethods) StoreVerifierToClientID(appID, verifier string, params *VerifierParams) {
	methodResponse := ServiceMethod(cam.bus, cam.caller, cam.service, "store_verifier_params", appID, verifier, params)
	if methodResponse.Error != nil {
		log.WithError(methodResponse.Error).Error("StoreVerifierToClientID")
	}
}
func (cam *CacheMethods) StorePartitionForApp(appID string, partitioned bool) {
	methodResponse := ServiceMethod(cam.bus, cam.caller, cam.service, "store_app_partition", appID, partitioned)
	if methodResponse.Error != nil {
		log.WithError(methodResponse.Error).Error("StorePartitionForApp")
	}
}

func (cam *CacheMethods) RetrieveClientIDFromVerifier(appID, verifier string) *VerifierParams {
	methodResponse := ServiceMethod(cam.bus, cam.caller, cam.service, "retrieve_verifier_params", appID, verifier)
	if methodResponse.Error != nil {
		log.WithError(methodResponse.Error).Error("RetrieveClientIDFromVerifier")
	}
	var params VerifierParams
	err := CastOrUnmarshal(methodResponse.Data, &params)
	if err != nil {
		log.WithError(err).Error("RetrieveClientIDFromVerifier")
		return nil
	}
	return &params
}
func (cam *CacheMethods) GetPartitionForApp(appID string) (partitioned bool, err error) {
	methodResponse := ServiceMethod(cam.bus, cam.caller, cam.service, "retrieve_app_partition", appID)
	if methodResponse.Error != nil {
		log.WithError(methodResponse.Error).Error("GetPartitionForApp")
		err = methodResponse.Error
		return
	}
	err = CastOrUnmarshal(methodResponse.Data, &partitioned)
	if err != nil {
		log.WithError(err).Error("GetPartitionForApp")
		return
	}
	return
}

func (cam *CacheMethods) GetTokenCommitKey(verifier string, tokenCommitment string) (pubKey Point) {
	err := retry.Do(func() error {
		methodResponse := ServiceMethod(cam.bus, cam.caller, cam.service, "get_token_commit_key", verifier, tokenCommitment)
		if methodResponse.Error != nil {
			return methodResponse.Error
		}
		var data Point
		err := CastOrUnmarshal(methodResponse.Data, &data)
		if err != nil {
			return err
		}
		pubKey = data
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("could not check if token commit exists")
	}
	return
}
