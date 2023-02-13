package chain

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/arcana-network/dkgnode/appdata"
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"
	"github.com/arcana-network/dkgnode/crypto"
	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/arcana-network/dkgnode/nodelist"
	"github.com/arcana-network/dkgnode/secp256k1"

	"github.com/avast/retry-go"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
	tmp2p "github.com/tendermint/tendermint/p2p"
)

const DEFAULT_KEY_BUFFER = 50000

type NodeRegister struct {
	AllConnected bool
	NodeList     []*common.NodeReference
}
type ChainService struct {
	sync.Mutex
	bus             eventbus.Bus
	broker          *common.MessageBroker
	running         bool
	client          *ethclient.Client
	pubKey          *ecdsa.PublicKey
	privKey         *ecdsa.PrivateKey
	cachedEpochInfo *EpochCache
	addr            *ethCommon.Address
	nodeList        *nodelist.NodeList
	nodeRegisterMap map[int]*NodeRegister
	isWhitelisted   bool
	tmp2pConnection string
	p2pConnection   string
	isRegistered    bool
	currentEpoch    int
	index           int
}

type EpochCache struct {
	sync.Map
}

func (s *EpochCache) Get(epoch int) (e common.EpochInfo, found bool) {
	val, ok := s.Map.Load(epoch)
	if !ok {
		return e, false
	}
	return val.(common.EpochInfo), true
}

func (s *EpochCache) Set(epoch int, e common.EpochInfo) {
	s.Map.Store(epoch, e)
}

func New(bus eventbus.Bus) *ChainService {
	return &ChainService{
		bus:    bus,
		broker: common.NewServiceBroker(bus, common.CHAIN_SERVICE_NAME),
	}
}

func (service *ChainService) ID() string {
	return common.CHAIN_SERVICE_NAME
}

func (service *ChainService) Start() error {
	client, err := ethclient.Dial(config.GlobalConfig.EthConnection)
	if err != nil {
		return err
	}
	service.client = client

	privateKeyECDSA, err := ethCrypto.ToECDSA(config.GlobalConfig.PrivateKey)
	if err != nil {
		return err
	}
	service.privKey = privateKeyECDSA

	nodePublicKey, err := getPublicKey(privateKeyECDSA)
	if err != nil {
		return err
	}
	service.pubKey = nodePublicKey

	nodeAddress := ethCrypto.PubkeyToAddress(*nodePublicKey)
	service.addr = &nodeAddress

	nodeListAddress := ethCommon.HexToAddress(config.GlobalConfig.ContractAddress)
	NodeListContract, err := nodelist.NewNodeList(nodeListAddress, client)
	if err != nil {
		return err
	}
	service.nodeList = NodeListContract
	service.currentEpoch = 1
	service.index = 0
	service.running = true
	service.cachedEpochInfo = &EpochCache{}
	service.nodeRegisterMap = make(map[int]*NodeRegister)

	go whitelistMonitor(service)

	go registerNode(service)

	go currentNodesMonitor(service)

	return nil
}

func (chain *ChainService) getArcanaContract(appID string) (*appdata.Arcana, error) {
	appAddress := ethCommon.HexToAddress(appID)
	appContract, err := appdata.NewArcana(appAddress, chain.client)
	if err != nil {
		return nil, err
	}
	return appContract, nil
}
func (chain *ChainService) getKeyPartition(appID string) (unpartitioned bool, err error) {
	partitioned := true
	c, err := chain.getArcanaContract(appID)
	if err != nil {
		log.WithField("err", err).Error("ChainService:getKeyPartition")
		return partitioned, errors.New("error while connecting to contract")
	}
	unpartitioned, err = c.Unpartitioned(nil)
	if err != nil {
		return partitioned, err
	}
	partitioned = !unpartitioned
	log.Infof("unpartitioned value from contract: %v", unpartitioned)
	return partitioned, nil
}

type Creds struct {
	Provider string `json:"verifier"`
	ClientID string `json:"client_id"`
}

type GatewayResponse struct {
	Creds []Creds `json:"cred"`
}

type GatewayIDFromAddressResponse struct {
	ID int `json:"id"`
}

func GatewayUrl(path, query string) (*url.URL, error) {
	log.Info("GetGatewayUrl", log.Fields{
		"gatewayUrl": config.GlobalConfig.GatewayURL,
	})
	u, err := url.Parse(config.GlobalConfig.GatewayURL)
	if err != nil {
		return nil, err
	}
	u.Path = path
	u.RawQuery = query
	return u, nil
}

func fetchClientID(appID string, p string) (string, error) {
	u, err := GatewayUrl("/api/v1/get-app-config/",
		fmt.Sprintf("id=%s", appID))
	if err != nil {
		return "", err
	}
	log.Debug("FetchClientID", log.Fields{
		"url": u.String(),
	})
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("FetchClientID.client.Do()")
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("FetchClientID.ReadBody")
		return "", err
	}
	r := GatewayResponse{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("FetchClientID.Unmarshal")
		return "", err
	}
	for _, v := range r.Creds {
		if p == v.Provider {
			return v.ClientID, nil
		}
	}
	return "", errors.New("ClientID not found")
}

func ClientID(appID, provider string) (clientID string, err error) {
	allowedProviders := []string{
		"google", "discord", "twitch",
		"reddit", "twitter", "github",
		"passwordless",
	}
	if stringInSlice(provider, allowedProviders) {
		if provider == "passwordless" {
			return appID, nil
		}

		clientID, err = fetchClientID(appID, provider)
		log.Debug("FetchClientID", log.Fields{
			"clientID": clientID,
			"err":      err,
		})
		return clientID, err
	}
	return "", errors.New("Invalid verifier")
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func getPublicKey(privateKey *ecdsa.PrivateKey) (*ecdsa.PublicKey, error) {
	nodePublicKey := privateKey.Public()
	nodePublicKeyEC, ok := nodePublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting to Public Key")
	}
	return nodePublicKeyEC, nil
}

func (cm *ChainService) ChainID() (chainID *big.Int, err error) {
	ctx := context.Background()
	chainID, err = cm.client.ChainID(ctx)
	return
}

func (s *ChainService) RegisterNode(epoch int, declaredIP string, TMP2PConnection string, P2PConnection string) (*types.Transaction, error) {
	txOpts, err := s.createTransactionOpts()
	if err != nil {
		log.WithError(err).Error("ListNode()")
		return nil, err
	}
	log.WithFields(log.Fields{
		"declaredIP": declaredIP,
		"epoch":      epoch,
	}).Info("RegisterNode()")
	tx, err := s.nodeList.ListNode(txOpts, big.NewInt(int64(epoch)), declaredIP, s.pubKey.X, s.pubKey.Y, "", "")
	if err != nil {
		log.WithError(err).Error("ListNode()")
		return nil, err
	}
	return tx, nil
}

func (s *ChainService) createTransactionOpts() (*bind.TransactOpts, error) {
	nonce, err := s.client.PendingNonceAt(context.Background(), ethCrypto.PubkeyToAddress(*s.pubKey))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	chainID, err := s.ChainID()
	if err != nil {
		log.WithError(err).Error("NewKeyedTransactorWithChainID()")
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.privKey, chainID)
	if err != nil {
		log.WithError(err).Error("NewKeyedTransactorWithChainID()")
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = 0

	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		log.WithError(err).Error("SuggestGasPrice()")
		return nil, err
	}
	auth.GasPrice = gasPrice

	return auth, nil
}

func registerNode(e *ChainService) {
	for {
		if e.isWhitelisted {
			break
		}
		log.Info("Node is not whitelisted yet.")
		time.Sleep(10 * time.Second)
	}
	var registered bool
	err := retry.Do(func() error {
		res, err := e.IsSelfRegistered(e.currentEpoch)
		if err != nil {
			return fmt.Errorf("could not check if node was registered on node list, %v", err.Error())
		}
		registered = res
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal()
	}

	externalAddr := "tcp://" + config.GlobalConfig.IPAddress + ":" + strings.Split(config.GlobalConfig.TMP2PListenAddress, ":")[2]
	tmp2pNodeKey := e.broker.TendermintMethods().GetNodeKey()
	p2pHostAddress := e.broker.P2PMethods().GetHostAddress()
	splitP2PHostAddr := strings.Split(p2pHostAddress, "/")
	splitP2PHostAddr[2] = config.GlobalConfig.IPAddress
	hostP2PAddressWithIP := strings.Join(splitP2PHostAddr, "/")

	e.tmp2pConnection = tmp2p.IDAddressString(tmp2pNodeKey.ID(), externalAddr)
	e.p2pConnection = hostP2PAddressWithIP

	log.WithFields(log.Fields{
		"p2pConnection":   e.p2pConnection,
		"tmp2pConnection": e.tmp2pConnection,
		"externalAddr":    externalAddr,
	}).Info("before registering node")

	if !registered {
		port := config.GlobalConfig.HttpServerPort
		var endpoint string
		if len(config.GlobalConfig.Domain) > 0 {
			endpoint = config.GlobalConfig.Domain
		} else {
			endpoint = config.GlobalConfig.IPAddress + ":" + port
		}

		log.WithFields(log.Fields{
			"IPAddress":       config.GlobalConfig.IPAddress,
			"Port":            port,
			"IDAddressString": tmp2p.IDAddressString(tmp2pNodeKey.ID(), externalAddr),
			"PublicEndpoint":  endpoint,
		}).Info("Registering self on contract")

		_, err := e.RegisterNode(
			e.currentEpoch,
			endpoint,
			e.tmp2pConnection,
			e.p2pConnection,
		)
		if err != nil {
			log.WithError(err).Fatal()
		}
	}
	e.isRegistered = true
}

func whitelistMonitor(e *ChainService) {
	interval := time.NewTicker(10 * time.Second)
	for range interval.C {
		isWhitelisted, err := e.nodeList.IsWhitelisted(nil, big.NewInt(int64(e.currentEpoch)), *e.addr)
		if err != nil {
			log.WithError(err).Error("could not check ethereum whitelist")
		}
		if isWhitelisted {
			e.isWhitelisted = true
			break
		}
		log.Info("node is not whitelisted yet!")
	}
}
func (s *ChainService) IsSelfRegistered(epoch int) (bool, error) {
	opts := s.CallOpts()
	result, err := s.nodeList.NodeRegistered(opts, big.NewInt(int64(epoch)), *s.addr)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (s *ChainService) CallOpts() *bind.CallOpts {
	auth := bind.CallOpts{
		From: *s.addr,
	}
	return &auth
}

func (s *ChainService) Sign(data []byte) []byte {
	ecSig := crypto.SignData(data, s.privKey)
	return ecSig.Raw
}
func (chainService *ChainService) Stop() error {
	return nil
}

func (chainService *ChainService) IsRunning() bool {
	return chainService.running
}

func (chainService *ChainService) getBuffer() int {
	buffer, err := chainService.nodeList.BufferSize(nil)
	if err != nil {
		return DEFAULT_KEY_BUFFER
	}

	return int(buffer.Int64())
}

func (e *ChainService) verifyDataWithNodelist(pk common.Point, sig []byte, data []byte) (senderDetails common.KeygenNodeDetails, err error) {
	// Check if PubKey Exists in Nodelist
	nodeExists := false
	var foundNode *common.NodeReference
	e.Lock()
	for _, nodeRegister := range e.nodeRegisterMap {
		for _, nodeRef := range nodeRegister.NodeList {
			if nodeRef.PublicKey.X.Cmp(&pk.X) == 0 && nodeRef.PublicKey.Y.Cmp(&pk.Y) == 0 {
				foundNode = nodeRef
				nodeExists = true
			}
		}
	}
	e.Unlock()
	if !nodeExists {
		err = fmt.Errorf("node doesnt exist in node register map")
		return
	}
	// Check validity of signature
	valid := crypto.VerifyPtFromRaw(data, pk, sig)
	if !valid {
		err = fmt.Errorf("invalid ecdsa sig for data %v", data)
		return
	}
	return common.KeygenNodeDetails{
		Index: int(foundNode.Index.Int64()),
		PubKey: common.Point{
			X: *foundNode.PublicKey.X,
			Y: *foundNode.PublicKey.Y,
		},
	}, err
}

func (chainService *ChainService) Call(method string, args ...interface{}) (interface{}, error) {
	log.WithFields(log.Fields{
		"method": method,
	}).Debug("ChainService:Call()")
	switch method {
	case "get_self_address":

		chainService.Lock()
		defer chainService.Unlock()
		if chainService.addr == nil {
			return nil, errors.New("node address has not been initialized")
		}
		return *chainService.addr, nil
	case "set_self_index":

		chainService.Lock()
		defer chainService.Unlock()
		var args0 int
		fmt.Println("setselfindex args", args)
		_ = common.CastOrUnmarshal(args[0], &args0)

		chainService.index = args0
		return nil, nil
	case "get_tm_p2p_connection":
		return chainService.tmp2pConnection, nil
	case "get_p2p_connection":
		return chainService.p2pConnection, nil
	case "get_current_epoch":
		log.WithField("get_curent_epoch_return", chainService.currentEpoch).Debug("ChainService")
		return chainService.currentEpoch, nil
	case "validate_epoch_pub_key":
		var args0 ethCommon.Address
		var args1 common.Point
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)
		pubKey, err := common.NewServiceBroker(chainService.bus, "chain").DBMethods().RetrieveNodePubKey(args0)
		if err != nil {
			log.WithField("error", err).Info("ValidateEpochPubKey")
			return false, err
		}
		log.WithFields(log.Fields{
			"address": args0.String(),
			"input":   args1,
			"storage": pubKey,
		}).Debug("ValidateEpochPubKey")
		if pubKey.X.Cmp(&args1.X) == 0 && pubKey.Y.Cmp(&args1.Y) == 0 {
			log.WithField("comparison", "pubkey is valid").Debug("ValidateEpochPubKey")
			return true, nil
		}
		return false, errors.New("incorrect pubkey")
	case "get_previous_epoch":

		epochInfo, err := chainService.GetEpochInfo(chainService.currentEpoch, false)
		if err != nil {
			return nil, err
		}
		prevEpoch := int(epochInfo.PrevEpoch.Int64())
		return prevEpoch, nil
	// GetNextEpoch() (epoch int)
	case "get_next_epoch":
		epochInfo, err := chainService.GetEpochInfo(chainService.currentEpoch, false)
		if err != nil {
			return nil, err
		}
		nextEpoch := int(epochInfo.NextEpoch.Int64())
		return nextEpoch, nil
	case "get_self_index":
		chainService.Lock()
		defer chainService.Unlock()
		for {
			if chainService.index != 0 {
				return chainService.index, nil
			}
			chainService.Unlock()
			time.Sleep(1 * time.Second)
			chainService.Lock()
		}
	case "get_self_public_key":
		chainService.Lock()
		defer chainService.Unlock()
		return common.Point{
			X: *chainService.pubKey.X,
			Y: *chainService.pubKey.Y,
		}, nil
	case "get_address":
		return chainService.addr, nil
	case "get_self_private_key":
		return *chainService.privKey.D, nil
	case "self_sign_data":
		var args0 []byte
		_ = common.CastOrUnmarshal(args[0], &args0)
		rawSig := chainService.Sign(args0)
		return rawSig, nil
	case "get_key_buffer":
		buffer := chainService.getBuffer()
		return buffer, nil
	case "await_nodes_connected":
		var args0 int
		_ = common.CastOrUnmarshal(args[0], &args0)

		interval := time.NewTicker(1 * time.Second)
		if chainService.nodeRegisterMap[args0] != nil && len(chainService.nodeRegisterMap[args0].NodeList) > 0 {
			return nil, nil
		}
		for range interval.C {
			if chainService.nodeRegisterMap[args0] != nil && len(chainService.nodeRegisterMap[args0].NodeList) > 0 {
				return nil, nil
			}
			log.WithField("epoch", args0).Debug("waiting for nodes to be connected")
		}
	case "get_node_details_by_address":

		var args0 ethCommon.Address
		_ = common.CastOrUnmarshal(args[0], &args0)

		chainService.Lock()
		defer chainService.Unlock()
		for _, nodeRegister := range chainService.nodeRegisterMap {
			for _, nodeDetails := range nodeRegister.NodeList {
				if nodeDetails.Address.String() == args0.String() {
					return nodeDetails.Serialize(), nil
				}
			}
		}
		return nil, fmt.Errorf("node could not be found for address %v", args0)
	case "verify_data_with_nodelist":

		var args0 common.Point
		var args1, args2 []byte
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)
		_ = common.CastOrUnmarshal(args[2], &args2)

		return chainService.verifyDataWithNodelist(args0, args1, args2)
	case "await_complete_node_list":
		var args0 int
		_ = common.CastOrUnmarshal(args[0], &args0)

		nodeEpoch := args0
		if chainService.nodeList == nil {
			return nil, errors.New("nodelist contract is undefined")
		}
		first := true
		for {
			if !first {
				time.Sleep(10 * time.Second)
			}
			first = false
			log.Info("attempting to retrieve complete node list")
			if chainService.nodeRegisterMap[nodeEpoch] == nil {
				log.WithField("nodeRegisterMap", chainService.nodeRegisterMap).Error("could not get node list")
				continue
			}
			currEpochInfo, err := chainService.GetEpochInfo(nodeEpoch, true)
			if err != nil {
				log.WithError(err).Error("could not get current epoch info")
				continue
			}
			nodeList := chainService.nodeRegisterMap[nodeEpoch].NodeList
			if currEpochInfo.N.Cmp(big.NewInt(int64(len(nodeList)))) != 0 {
				log.WithFields(log.Fields{
					"nodeList":      len(nodeList),
					"expectedNodes": currEpochInfo.N.Int64(),
				}).Error("NodeList and expected not yet equal, waiting...")
			} else {
				break
			}
		}
		nodeReferences := make([]common.SerializedNodeReference, 0)
		for _, nodeDetails := range chainService.nodeRegisterMap[nodeEpoch].NodeList {
			nodeReferences = append(nodeReferences, nodeDetails.Serialize())
		}
		return nodeReferences, nil
	case "verify_data_with_epoch":

		var args0 common.Point
		var args1, args2 []byte
		var args3 int
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)
		_ = common.CastOrUnmarshal(args[2], &args2)
		_ = common.CastOrUnmarshal(args[3], &args3)

		return chainService.verifyDataWithEpoch(args0, args1, args2, args3)
	case "get_epoch_info":

		var args0 int
		var args1 bool

		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)

		eInfo, err := chainService.GetEpochInfo(args0, args1)
		log.WithField("get_curent_epoch_info_return", eInfo).Debug("ChainService")

		return eInfo, err
	case "get_node_details_by_epoch_and_index":

		var args0, args1 int
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)

		chainService.Lock()
		defer chainService.Unlock()
		for _, nodeDetails := range chainService.nodeRegisterMap[args0].NodeList {
			if int(nodeDetails.Index.Int64()) == args1 {
				return nodeDetails.Serialize(), nil
			}
		}
		return nil, fmt.Errorf("node could not be found for %v %v", args0, args1)
	case "get_clientid_by_verifier":
		var args0, args1 string
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)

		log.WithFields(log.Fields{
			"args0": args0,
			"args1": args1,
		}).Debug("get_clientid_by_verifier")
		return ClientID(args0, args1)
	case "get_app_partition":
		var args0 string
		_ = common.CastOrUnmarshal(args[0], &args0)

		return chainService.getKeyPartition(args0)
	}
	return "", nil
}

func (e *ChainService) GetEpochInfo(epoch int, skipCache bool) (common.EpochInfo, error) {
	if !skipCache {
		eInfo, found := e.cachedEpochInfo.Get(epoch)
		if found {
			return eInfo, nil
		}
	}

	opts := e.CallOpts()
	if epoch == 0 {
		return common.EpochInfo{}, fmt.Errorf("epoch %v is invalid", epoch)
	}
	result, err := e.nodeList.GetEpochInfo(opts, big.NewInt(int64(epoch)))
	log.WithField("epoch data", result).Info("GetEpochInfo()")
	if err != nil {
		return common.EpochInfo{}, err
	}
	if result.Id.Cmp(big.NewInt(0)) == 0 {
		return common.EpochInfo{}, fmt.Errorf("epoch %v has not been initialized", epoch)
	}
	eInfo := common.EpochInfo{
		Id:        *result.Id,
		N:         *result.N,
		K:         *result.K,
		T:         *result.T,
		PrevEpoch: *result.PrevEpoch,
		NextEpoch: *result.NextEpoch,
	}
	e.cachedEpochInfo.Set(epoch, eInfo)
	return eInfo, nil
}

func (chainService *ChainService) verifyDataWithEpoch(pk common.Point, sig []byte, data []byte, epoch int) (senderDetails common.KeygenNodeDetails, err error) {
	// Check if PubKey Exists in Nodelist
	nodeExists := false
	var foundNode *common.NodeReference
	chainService.Lock()
	nodeRegister, ok := chainService.nodeRegisterMap[epoch]
	if !ok {
		err = fmt.Errorf("epoch doesnt exist in node register map, verifyDataWithEpoch")
		return
	}
	for _, nodeRef := range nodeRegister.NodeList {
		if nodeRef.PublicKey.X.Cmp(&pk.X) == 0 && nodeRef.PublicKey.Y.Cmp(&pk.Y) == 0 {
			foundNode = nodeRef
			nodeExists = true
		}
	}
	chainService.Unlock()
	if !nodeExists {
		err = fmt.Errorf("node doesnt exist in node register map")
		return
	}

	// Check validity of signature
	valid := crypto.VerifyPtFromRaw(data, pk, sig)
	if !valid {
		err = fmt.Errorf("invalid ecdsa sig for data %v", data)
		return
	}
	return common.KeygenNodeDetails{
		Index: int(foundNode.Index.Int64()),
		PubKey: common.Point{
			X: *foundNode.PublicKey.X,
			Y: *foundNode.PublicKey.Y,
		},
	}, err
}

func (e *ChainService) getNodeRefsByEpoch(epoch int) ([]*common.NodeReference, error) {
	log.WithField("epoch", epoch).Debug("getNodeRefsByEpoch called")
	ethList, err := e.nodeList.GetNodes(nil, big.NewInt(int64(epoch)))
	if err != nil {
		return nil, fmt.Errorf("Could not get node list %v", err.Error())
	}
	var currNodeList []*common.NodeReference

	for i := 0; i < len(ethList); i++ {
		detailsWithPubK, err := e.nodeList.NodeDetails(nil, ethList[i])
		if err != nil {
			return nil, fmt.Errorf("could not get node details with pub key %v", err.Error())
		}
		err = common.NewServiceBroker(e.bus, "chain").DBMethods().StoreNodePubKey(ethList[i], common.Point{X: *detailsWithPubK.PubKx, Y: *detailsWithPubK.PubKy})
		if err != nil {
			return nil, fmt.Errorf("could not store node details with pub key %v", err.Error())
		}
	}

	for i := 0; i < len(ethList); i++ {
		nodeRef, err := e.GetNodeRef(ethList[i])
		if err != nil {
			return nil, fmt.Errorf("Could not get node refs %v", err.Error())
		}
		currNodeList = append(currNodeList, nodeRef)
	}
	return currNodeList, nil
}

func (e *ChainService) GetNodeRef(nodeAddress ethCommon.Address) (n *common.NodeReference, err error) {
	details, err := e.nodeList.NodeDetails(nil, nodeAddress)
	if err != nil {
		return nil, err
	}

	var connectionDetails common.ConnectionDetails
	if details.DeclaredIp != "" && details.P2pListenAddress == "" || details.TmP2PListenAddress == "" {
		err = retry.Do(func() error {
			var retryErr error
			connectionDetails, retryErr = e.broker.ServerMethods().RequestConnectionDetails(details.DeclaredIp)
			log.WithField("connectionDetails", connectionDetails).Debug("got back connection details from node")
			if retryErr != nil {
				return fmt.Errorf("could not get hidden connection details %v", retryErr)
			}
			retryErr = e.broker.DBMethods().StoreConnectionDetails(nodeAddress, connectionDetails)
			if retryErr != nil {
				return fmt.Errorf("could not store connection details %v", retryErr)
			}
			return nil
		})
		if err != nil {
			log.WithField("nodeAddress", nodeAddress).WithError(err).Error("could not get connection details from node, get from DB")
			connectionDetails, err = e.broker.DBMethods().RetrieveConnectionDetails(nodeAddress)
			if err != nil {
				log.WithField("nodeAddress", nodeAddress).Error("could not get connection details from DB either")
				return nil, fmt.Errorf("unable to get connection details for nodeAddress %v", nodeAddress)
			}
		}
	} else {
		connectionDetails = common.ConnectionDetails{
			TMP2PConnection: details.TmP2PListenAddress,
			P2PConnection:   details.P2pListenAddress,
		}
	}

	peerid, err := common.GetPeerIDFromP2pListenAddress(connectionDetails.P2PConnection)
	if err != nil {
		return nil, err
	}
	return &common.NodeReference{
		Address:         &nodeAddress,
		PeerID:          *peerid,
		Index:           details.Position,
		PublicKey:       &ecdsa.PublicKey{Curve: secp256k1.Curve, X: details.PubKx, Y: details.PubKy},
		TMP2PConnection: connectionDetails.TMP2PConnection,
		P2PConnection:   connectionDetails.P2PConnection,
	}, nil
}

func currentNodesMonitor(e *ChainService) {
	interval := time.NewTicker(10 * time.Second)

	for range interval.C {
		currEpoch := e.broker.ChainMethods().GetCurrentEpoch()
		currEpochInfo, err := e.GetEpochInfo(currEpoch, true)
		if err != nil {
			log.WithError(err).Error("could not get curr epoch")
			continue
		}
		e.Lock()
		if _, ok := e.nodeRegisterMap[currEpoch]; !ok {
			e.nodeRegisterMap[currEpoch] = &NodeRegister{}
		}
		e.Unlock()
		log.WithField("currEpoch", currEpoch).Info("currentNodesMonitor calling NodeRefs")
		currNodeList, err := e.getNodeRefsByEpoch(currEpoch)
		log.WithField("currNodeList", currNodeList).Info("currentNodesMonitor calling currNodeList result")
		if err != nil {
			log.WithError(err).Error("could not get currNodeList")
			continue
		}
		if currEpochInfo.N.Cmp(big.NewInt(int64(len(currNodeList)))) != 0 {
			log.WithFields(log.Fields{
				"currNodeList":  currNodeList,
				"currEpochInfo": currEpochInfo,
			}).Error("currentNodeList does not equal in length to expected currEpochInfo")
			continue
		}
		allNodesConnected := true
		for _, nodeRef := range currNodeList {
			err = e.broker.P2PMethods().ConnectToP2PNode(nodeRef.P2PConnection, nodeRef.PeerID)
			if err != nil {
				log.WithField("Address", *nodeRef.Address).Error("could not connect to p2p node ...continuing...")
				allNodesConnected = false
			}
			if nodeRef.PeerID == e.broker.P2PMethods().ID() {
				e.broker.ChainMethods().SetSelfIndex(int(nodeRef.Index.Int64()))
			}
		}
		if !allNodesConnected {
			continue
		}
		log.WithField("currNodeList", currNodeList).Debug("connected to all nodes in current epoch")
		e.Lock()
		e.nodeRegisterMap[currEpoch].NodeList = currNodeList
		e.Unlock()
		break
	}
}
