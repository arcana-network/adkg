package tendermint

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"
	"github.com/avast/retry-go"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/arcana-network/dkgnode/eventbus"

	log "github.com/sirupsen/logrus"
	btcec "github.com/tendermint/btcd/btcec"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmsecp "github.com/tendermint/tendermint/crypto/secp256k1"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmlog "github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	tmp2p "github.com/tendermint/tendermint/p2p"
	privval "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/rpc/client/http"
	tmtypes "github.com/tendermint/tendermint/types"
)

type TendermintService struct {
	bus             eventbus.Bus
	tmNodeKey       *tmp2p.NodeKey
	node            *nm.Node
	bftSemaphore    *Semaphore
	bftrpc          *BFTRPC
	websocketStatus Status
}

type Status int

const (
	WebSocketDown Status = iota
	WebSocketUp
)

func NewCore(bus eventbus.Bus) *TendermintService {
	service := TendermintService{
		bus: bus,
	}
	return &service
}

func (*TendermintService) ID() string {
	return common.TENDERMINT_SERVICE_NAME
}

func (t *TendermintService) IsRunning() bool {
	return true
}

func (t *TendermintService) Stop() error {
	return t.node.Stop()
}

func (t *TendermintService) Start() error {
	err := createTendermintFolderStructure(config.GlobalConfig.BasePath)
	if err != nil {
		log.WithError(err).Fatalln("Error during creation of folder structure")
	}

	tmRootPath := config.GlobalConfig.BasePath + "/tendermint"

	nodeKey, err := getTendermintNodeKey(tmRootPath)
	if err != nil {
		log.WithError(err).Fatal("NodeKey generation issue")
	}
	t.tmNodeKey = nodeKey
	t.websocketStatus = WebSocketDown
	go abciMonitor(t)
	go t.startTendermintCore(tmRootPath, nodeKey)
	return nil
}

func createTendermintFolderStructure(basePath string) error {
	err := os.MkdirAll(basePath+"/tendermint", os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not makedir for tendermint")
	}
	err = os.MkdirAll(basePath+"/tendermint/config", os.ModePerm)
	if err != nil {
		log.WithError(err).Error("could not makedir for tendermint config")
		return fmt.Errorf("could not makedir for tendermint config")
	}
	err = os.MkdirAll(basePath+"/tendermint/data", os.ModePerm)
	if err != nil {
		log.WithError(err).Error("could not makedir for tendermint data")
		return fmt.Errorf("could not makedir for tendermint data")
	}
	err = os.Remove(basePath + "/tendermint/data/cs.wal/wal")
	if err != nil {
		log.WithError(err).Error("could not remove write ahead log")
	} else {
		log.Debug("Removed write ahead log")
	}
	return nil
}

func getTendermintNodeKey(tendermintRootPath string) (*tmp2p.NodeKey, error) {
	dftConfig := cfg.DefaultConfig()
	dftConfig.SetRoot(tendermintRootPath)
	var tmNodeKey *tmp2p.NodeKey
	if len(config.GlobalConfig.TMPrivateKey) != 0 {
		tmNodeKey = &tmp2p.NodeKey{
			PrivKey: ed25519.PrivKey(config.GlobalConfig.TMPrivateKey),
		}
	} else {
		k, err := tmp2p.LoadOrGenNodeKey(dftConfig.NodeKeyFile())
		if err != nil {
			return nil, err
		}
		tmNodeKey = k
	}
	err := tmNodeKey.SaveAs(dftConfig.NodeKeyFile())
	return tmNodeKey, err
}

func (t *TendermintService) startTendermintCore(buildPath string, nodeKey *tmp2p.NodeKey) {
	chainMethods := common.NewServiceBroker(t.bus, "tendermint").ChainMethods()

	nodeList := chainMethods.AwaitCompleteNodeList(chainMethods.GetCurrentEpoch())

	peerList, validators := getValidatorsAndPeerFromNodeList(nodeList)
	log.WithFields(log.Fields{
		"peerList":   peerList,
		"validators": validators,
	}).Info("adding these peers")

	defaultConfig := getTendermintConfig(buildPath, strings.Join(peerList, ","))
	saveValidatorKey(chainMethods.GetSelfPrivateKey(), defaultConfig.PrivValidatorKeyFile(), defaultConfig.PrivValidatorStateFile())

	genesisDoc := createGenesisDoc(validators)
	saveGenesisDoc(genesisDoc, defaultConfig.GenesisFile())

	err := verifyAndSaveConfig(defaultConfig)
	if err != nil {
		log.WithError(err).Fatal("config doesnt pass validation checks")
	}

	node, err := createTendermintNode(defaultConfig)
	if err != nil {
		log.WithError(err).Fatal("failed to create tendermint node")
	}

	t.node = node

	if err := node.Start(); err != nil {
		log.WithError(err).Fatal("failed to start tendermint node")
	}
	log.WithField("NodeInfo", node.Switch().NodeInfo()).Info("started tendermint")
}

func saveGenesisDoc(genesisDoc tmtypes.GenesisDoc, savePath string) {
	if err := genesisDoc.SaveAs(savePath); err != nil {
		log.WithError(err).Error("could not save gendoc")
	}
}

func getValidatorsAndPeerFromNodeList(nodeList []common.NodeReference) ([]string, []tmtypes.GenesisValidator) {
	var validators []tmtypes.GenesisValidator
	var persistantPeersList []string
	for i := range nodeList {
		pubkeyBytes := rawPointToTMPubKey(nodeList[i].PublicKey.X, nodeList[i].PublicKey.Y)
		validators = append(validators, tmtypes.GenesisValidator{
			Address: pubkeyBytes.Address(),
			PubKey:  pubkeyBytes,
			Power:   1,
		})
		persistantPeersList = append(persistantPeersList, nodeList[i].TMP2PConnection)
	}
	return persistantPeersList, validators
}
func createGenesisDoc(validators []tmtypes.GenesisValidator) tmtypes.GenesisDoc {
	genesisDoc := tmtypes.GenesisDoc{
		ChainID:     "test-net-1",
		GenesisTime: time.Unix(1578036594, 0),
		Validators:  validators,
	}
	return genesisDoc
}

func saveValidatorKey(privKey big.Int, keyFilePath string, stateFilePath string) {
	pv := tmPrivateKeyFromBigInt(privKey)
	pvF := privval.NewFilePV(pv, keyFilePath, stateFilePath)
	pvF.Save()
}

func getTendermintConfig(buildPath string, peers string) *cfg.Config {
	defaultConfig := cfg.DefaultConfig()
	defaultConfig.SetRoot(buildPath)

	defaultConfig.ProxyApp = common.GetSocketAddress()

	defaultConfig.Consensus.CreateEmptyBlocks = false
	defaultConfig.LogLevel = "main:info,state:info,statesync:info,*:error"

	defaultConfig.Mempool.Size = 20000

	defaultConfig.LogLevel = "main:info,state:info,statesync:info,*:error"

	defaultConfig.BaseConfig.DBBackend = "goleveldb"
	defaultConfig.FastSyncMode = false
	// defaultConfig.RPC.ListenAddress = fmt.Sprintf("tcp://%s:26657", config.GlobalConfig.IPAddress)
	defaultConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
	defaultConfig.RPC.MaxSubscriptionClients = 5
	defaultConfig.RPC.MaxSubscriptionsPerClient = 200

	// defaultConfig.P2P.ListenAddress = fmt.Sprintf("tcp://%s:26656", config.GlobalConfig.IPAddress)
	defaultConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
	defaultConfig.P2P.MaxNumInboundPeers = 300
	defaultConfig.P2P.PersistentPeers = peers
	defaultConfig.P2P.MaxNumOutboundPeers = 300
	// recommended to run in production
	defaultConfig.P2P.SendRate = 20000000
	defaultConfig.P2P.RecvRate = 20000000
	defaultConfig.P2P.FlushThrottleTimeout = 10
	defaultConfig.P2P.MaxPacketMsgPayloadSize = 10240 // 10KB

	return defaultConfig
}

func verifyAndSaveConfig(defaultConfig *cfg.Config) error {
	err := defaultConfig.ValidateBasic()
	if err != nil {
		return fmt.Errorf("config doesnt pass validation checks")
	}
	cfg.WriteConfigFile(defaultConfig.RootDir+"/config/config.toml", defaultConfig)
	return nil
}

func createTendermintNode(defaultConfig *cfg.Config) (*nm.Node, error) {
	logr := log.New()
	logr.SetOutput(&lumberjack.Logger{
		Filename:   "/var/log/arcana/tm.log",
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	})
	logger := tmlog.NewTMLogger(logr.Writer())
	n, err := nm.DefaultNewNode(defaultConfig, logger)
	return n, err
}

func abciMonitor(t *TendermintService) {
	interval := time.NewTicker(5 * time.Second)
	defer interval.Stop()
	for range interval.C {
		bftClient, err := http.New(fmt.Sprintf("tcp://%s:26657", "127.0.0.1"), "/websocket")
		if err != nil {
			log.WithField("error during starting rpc abci", err).Error("ABCI")
		} else {
			err = bftClient.Start()
			if err != nil {
				log.WithField("error during starting ws", err).Error("ABCI")
			} else {
				t.bftrpc = NewBFTRPC(bftClient, t.bus)
				t.bftSemaphore = NewSemaphore(30)
				t.websocketStatus = WebSocketUp
				break
			}
		}
	}
}

func (t *TendermintService) Call(method string, args ...interface{}) (interface{}, error) {
	switch method {
	case "get_node_key":
		return tmjson.Marshal(*t.tmNodeKey)
	case "tx_status":
		var args0 []byte
		_ = common.CastOrUnmarshal(args[0], &args0)
		return t.isHashValid(args0), nil
	case "broadcast":
		err := retry.Do(func() error {
			if t.bftrpc == nil {
				log.Error("broadcast: bft rpc is not initialized yet")
				return fmt.Errorf("bft rpc is not initialized")
			}
			if t.websocketStatus == WebSocketDown {
				log.Error("broadcast: bft ws is not up")
				return fmt.Errorf("bft ws is not up yet")
			}
			return nil
		},
			retry.Attempts(6),
			retry.LastErrorOnly(true),
		)
		if err != nil {
			return nil, err
		}

		var args0 interface{}
		_ = common.CastOrUnmarshal(args[0], &args0)

		release := t.bftSemaphore.Acquire()
		txHash, err := t.bftrpc.Broadcast(args0)
		release()
		if err != nil {
			return nil, err
		}
		return *txHash, nil
	case "register_query":

		var args0 string
		_ = common.CastOrUnmarshal(args[0], &args0)

		query := args0
		release := t.bftSemaphore.Acquire()
		responseCh, cancel, err := t.RegisterQuery(query)
		release()
		if err != nil {
			return nil, err
		}
		go func() {
			for response := range responseCh {
				t.bus.Publish("tendermint:forward:"+query, common.MethodResponse{
					Error: nil,
					Data:  response,
				})
			}
			cancel()
		}()
		return nil, nil
	default:
		return nil, fmt.Errorf("tendermint service method %v not found", method)
	}
}

func (t *TendermintService) RegisterQuery(query string) (chan []byte, context.CancelFunc, error) {
	log.WithField("RegisterQuery", query).Debug("TendermintService")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*7)
	responseCh := make(chan []byte, 1)
	ch, err := t.bftrpc.Subscribe(ctx, "self", query)
	if err != nil {
		log.WithError(err).Error("RegisterQuery:Subscribe()")
		cancel()
		return nil, nil, err
	}

	go func() {
		log.Info("RegisterQuery:Subscribe() - success")
		result := <-ch
		eventDataTx := result.Data.(tmtypes.EventDataTx)
		d, err := eventDataTx.Marshal()
		if err != nil {
			log.WithError(err).Error("RegisterQuery:EventDataTx.Marshal()")
		} else {
			responseCh <- d
		}
		close(responseCh)
		err = t.bftrpc.Unsubscribe(context.Background(), "self", query)
		if err != nil {
			log.WithError(err).Error("Query:Unsubscribe()")
			return
		}
	}()
	return responseCh, cancel, nil
}

func (t *TendermintService) isHashValid(hash []byte) bool {
	res, err := t.bftrpc.Tx(context.Background(), hash, false)
	if err != nil {
		return false
	}
	if res.TxResult.Code == 0 {
		return true
	}
	return false
}

func rawPointToTMPubKey(X, Y *big.Int) tmsecp.PubKey {
	var pubkeyBytes tmsecp.PubKey
	pubkeyObject := btcec.PublicKey{
		X: X,
		Y: Y,
	}
	pubkeyBytes = pubkeyObject.SerializeCompressed()
	return pubkeyBytes
}

func tmPrivateKeyFromBigInt(key big.Int) tmsecp.PrivKey {
	var pv tmsecp.PrivKey
	keyBytes := common.PadPrivKeyBytes(key.Bytes())
	for i := 0; i < 32; i++ {
		pv = append(pv, keyBytes[i])
	}
	return pv
}
