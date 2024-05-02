package tendermint

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/acss"
	"github.com/arcana-network/dkgnode/secp256k1"

	log "github.com/sirupsen/logrus"
	code "github.com/tendermint/tendermint/abci/example/code"
	"github.com/tendermint/tendermint/abci/types"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/version"
	tmdb "github.com/tendermint/tm-db"
	"github.com/torusresearch/bijson"
)

var (
	stateKey                    = []byte("sk")
	keyMappingPrefixKey         = []byte("km")
	EDKeyMappingPrefixKey       = []byte("ekm")
	verifierToKeyIndexPrefixKey = []byte("vt")
	appInfoKey                  = []byte("ai")
)

type ABCI struct {
	db          *tmdb.GoLevelDB
	dbIterators *DBIteratorsSyncMap
	broker      *common.MessageBroker
	state       *State
	prevState   *State
	info        *AppInfo
}

type KeygenPubKey struct {
	ID    string       `json:"ID"`
	Point common.Point `json:"point"`
}

type KeygenDecision struct {
	Nodes []int `json:"nodes"`
}

type FinDecision struct {
	Nodes []int `json:"nodes"`
}

type getIndexesQuery struct {
	Provider string           `json:"provider"`
	UserID   string           `json:"user_id"`
	AppID    string           `json:"app_id"`
	Curve    common.CurveName `json:"curve"`
}

func getPartitionedKeyspace(appID, userID string, curve common.CurveName) []byte {
	key := []byte(strings.Join([]string{appID, userID}, common.Delimiter1))
	if curve == common.ED25519 {
		key = []byte(strings.Join([]string{appID, userID, string(common.ED25519)}, common.Delimiter1))
	}
	return append(verifierToKeyIndexPrefixKey, key...)
}
func getUnpartitionedKeyspace(userID string, curve common.CurveName) []byte {
	key := []byte(strings.Join([]string{"global", userID}, common.Delimiter1))
	if curve == common.ED25519 {
		key = []byte(strings.Join([]string{"global", userID, string(common.ED25519)}, common.Delimiter1))
	}
	return append(verifierToKeyIndexPrefixKey, key...)
}

type TransferSummaryID string
type TransferSummary struct {
	LastUnassignedIndex uint `json:"last_unassigned_index"`
}

func (t *TransferSummary) ID() TransferSummaryID {
	return TransferSummaryID(strconv.Itoa(int(t.LastUnassignedIndex)))
}

type MappingCounter struct {
	RequiredCount int
	KeyCount      int
}

type C25519State struct {
	LastCreatedIndex    uint `json:"last_created_index"`
	LastUnassignedIndex uint `json:"last_unassigned_index"`
}

type State struct {
	LastUnassignedIndex            uint                         `json:"last_unassigned_index"`
	LastCreatedIndex               uint                         `json:"last_created_index"`
	BlockTime                      time.Time                    `json:"-"`
	NewKeyAssignments              []common.KeyAssignmentPublic `json:"new_key_assignments"`
	KeygenDecisions                map[string]KeygenDecision    `json:"keygen_decisions"`
	KeygenPubKeys                  map[string]KeygenPubKey      `json:"keygen_pubkeys"`
	ConsecutiveFailedPubKeyAssigns uint                         `json:"consecutive_failed_pubkey_assigns"`
	C25519State                    C25519State                  `json:"c25519_state"`
	//TODO - BatchFinDecisions and DpssFinDecisions are very similar, consider merging both
	// to vote on wheter a batch has finished
	BatchFinDecisions map[string]FinDecision `json:"batch_fin_decisions"`
	// to vote on whether every batch has finished
	DpssFinDecisions map[string]FinDecision `json:"dpss_fin_decisions"`
}

func (state *State) KeyAvailable(curve common.CurveName) bool {
	if curve == common.SECP256K1 {
		return state.LastUnassignedIndex < state.LastCreatedIndex
	}
	if curve == common.ED25519 {
		return state.C25519State.LastUnassignedIndex < state.C25519State.LastCreatedIndex
	}
	return false
}

type AppInfo struct {
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

type DBIteratorsSyncMap struct {
	sync.Map
}

func (a *ABCI) NewABCI(broker *common.MessageBroker) *ABCI {
	db, err := tmdb.NewGoLevelDB("tmstate", config.GlobalConfig.BasePath+"/tmstate")
	if err != nil {
		log.WithError(err).Fatal("could not start GoLevelDB for tendermint state")
	}
	abci := ABCI{db: db, dbIterators: &DBIteratorsSyncMap{}, broker: broker}
	_, stateExists := abci.LoadState()

	if !stateExists {
		abci.state = &State{
			LastUnassignedIndex: 0,
			LastCreatedIndex:    0,
			KeygenDecisions:     make(map[string]KeygenDecision),
			KeygenPubKeys:       make(map[string]KeygenPubKey),
			C25519State: C25519State{
				LastCreatedIndex:    0,
				LastUnassignedIndex: 0,
			},
			BatchFinDecisions: make(map[string]FinDecision),
			DpssFinDecisions:  make(map[string]FinDecision),
		}
		abci.info = &AppInfo{
			Height: 0,
		}
	}

	return &abci
}

func (abci *ABCI) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	tx := req.GetTx()
	parsedTx, senderDetails, err := authenticateBftTx(tx, abci.broker)
	if err != nil {
		return abcitypes.ResponseDeliverTx{Code: code.CodeTypeUnauthorized}
	}

	correct, tags, err := abci.ValidateAndUpdateAndTagBFTTx(parsedTx.BFTTx, parsedTx.MsgType, senderDetails)
	if err != nil {
		log.WithError(err).Error("could not validate BFTTx")
		return abcitypes.ResponseDeliverTx{Code: code.CodeTypeUnauthorized}
	}

	if !correct {
		log.Error("tx not correct, could not be validated: err=%w", err)
		return abcitypes.ResponseDeliverTx{Code: code.CodeTypeUnknownError}
	}

	if tags == nil {
		tags = new([]abcitypes.EventAttribute)
	}

	return abcitypes.ResponseDeliverTx{Code: code.CodeTypeOK, Events: []abcitypes.Event{{Type: "transfer", Attributes: *tags}}}
}
func (abci *ABCI) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	tx := req.GetTx()
	parsedTx, senderDetails, err := authenticateBftTx(tx, abci.broker)
	if err != nil {
		return abcitypes.ResponseCheckTx{Code: code.CodeTypeUnauthorized}
	}
	validated, err := abci.validateTx(parsedTx.BFTTx, parsedTx.MsgType, senderDetails, abci.prevState)
	if err != nil {
		log.WithError(err).Error("could not validate BFTTx in checkTx")
	}

	if !validated {
		return abcitypes.ResponseCheckTx{Code: code.CodeTypeUnauthorized}
	}

	return abcitypes.ResponseCheckTx{Code: code.CodeTypeOK}
}

func (abci *ABCI) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	abci.state.BlockTime = req.Header.GetTime()
	abci.state.NewKeyAssignments = []common.KeyAssignmentPublic{}
	return abcitypes.ResponseBeginBlock{}
}
func (abci *ABCI) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	return abcitypes.ResponseInitChain{}
}
func (abci *ABCI) ListSnapshots(abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	resp := abcitypes.ResponseListSnapshots{Snapshots: []*abcitypes.Snapshot{}}
	return resp
}

func (abci *ABCI) LoadSnapshotChunk(req abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

func (abci *ABCI) SetOption(req abcitypes.RequestSetOption) abcitypes.ResponseSetOption {
	return abcitypes.ResponseSetOption{}
}

func (abci *ABCI) OfferSnapshot(abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

func (abci *ABCI) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	log.WithFields(log.Fields{
		"EndBlockHeight":      req.Height,
		"LastCreatedIndex":    int(abci.state.LastCreatedIndex),
		"LastUnassignedIndex": int(abci.state.LastUnassignedIndex),
	}).Info("EndBlock")

	// if PSS is running, skip adkg
	pssRunning, err := abci.broker.ChainMethods().GetCurrentPssStatus()
	if err != nil {
		log.WithError(err).Error("Could not get current PSS status")
	}
	if pssRunning {
		log.Info("PSS running, ADKG on halt")
		return abcitypes.ResponseEndBlock{}
	}

	buffer := abci.broker.ChainMethods().KeyBuffer()
	var maxKeyInit int
	if buffer > 500 {
		maxKeyInit = buffer / 500
	} else {
		maxKeyInit = 100
	}

	if int(abci.state.LastCreatedIndex)-int(abci.state.LastUnassignedIndex) < buffer {
		log.WithFields(log.Fields{
			"EndBlockHeight":      req.Height,
			"LastCreatedIndex":    int(abci.state.LastCreatedIndex),
			"LastUnassignedIndex": int(abci.state.LastUnassignedIndex),
		}).Info("EndBlock")
		end := MinOf(int(abci.state.LastCreatedIndex)+maxKeyInit, int(abci.state.LastUnassignedIndex)+buffer)
		log.WithFields(log.Fields{
			"Start":  int(abci.state.LastCreatedIndex),
			"End":    end,
			"Buffer": buffer,
		}).Info("EndBlock: Starting Keygens")
		for i := int(abci.state.LastCreatedIndex); i < end; i++ {
			id := common.GenerateADKGID(*big.NewInt(int64(i)))
			round := common.RoundDetails{
				ADKGID: id,
				Dealer: abci.broker.ChainMethods().GetSelfIndex(),
				Kind:   "acss",
			}
			msg, err := acss.NewShareMessage(
				round.ID(),
				common.SECP256K1,
			)
			if err != nil {
				log.WithError(err).Error("EndBlock:Acss.NewShareMessage")
				continue
			}
			err = abci.broker.KeygenMethods().ReceiveMessage(*msg)
			if err != nil {
				log.WithError(err).Error("Could not receive keygenmessage share")
			}
		}
	}

	edBuffer := buffer / 10
	maxEDKeyInit := maxKeyInit / 10
	endED := MinOf(int(abci.state.C25519State.LastCreatedIndex)+maxEDKeyInit, int(abci.state.C25519State.LastUnassignedIndex)+edBuffer)
	log.WithFields(log.Fields{
		"LastCreatedIndex":    int(abci.state.C25519State.LastCreatedIndex),
		"LastUnassignedIndex": int(abci.state.C25519State.LastUnassignedIndex),
	}).Info("EndBlock:ED25519")
	for i := int(abci.state.C25519State.LastCreatedIndex); i < endED; i++ {
		id := common.NewADKGID(*big.NewInt(int64(i)), common.ED25519)
		round := common.RoundDetails{
			ADKGID: id,
			Dealer: abci.broker.ChainMethods().GetSelfIndex(),
			Kind:   "acss",
		}
		msg, err := acss.NewShareMessage(
			round.ID(),
			common.ED25519,
		)
		if err != nil {
			log.WithError(err).Error("EndBlock:Acss.NewShareMessage")
			continue
		}
		err = abci.broker.KeygenMethods().ReceiveMessage(*msg)
		if err != nil {
			log.WithError(err).Error("Could not receive keygenmessage share")
		}
	}
	return abcitypes.ResponseEndBlock{}
}

func (app *ABCI) Info(req abcitypes.RequestInfo) (resInfo abcitypes.ResponseInfo) {
	return abcitypes.ResponseInfo{
		Version:          version.ABCIVersion,
		AppVersion:       version.BlockProtocol,
		LastBlockAppHash: app.info.AppHash,
		LastBlockHeight:  app.info.Height,
	}
}

func (abci *ABCI) ApplySnapshotChunk(abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}

func (abci *ABCI) Commit() abcitypes.ResponseCommit {
	// get the hash of the current state (including the previous app hash)
	byt, err := bijson.Marshal(abci.state)
	if err != nil {
		log.WithError(err).Fatal("could not marshal app state")
	}
	currAppHash := secp256k1.Keccak256(byt)

	// update prepare state for next block,
	abci.info.AppHash = currAppHash
	abci.info.Height += 1
	abci.SaveState()
	abci.prevState = nil
	err = bijson.Unmarshal(byt, &abci.prevState)
	if err != nil {
		log.WithError(err).Fatal("could not copy lagging state")
	}

	return abcitypes.ResponseCommit{Data: currAppHash}
}

func GetAppKeyPartition(broker *common.MessageBroker, appID string) (bool, error) {
	partitioned, err := broker.ChainMethods().GetPartitionForApp(appID)
	return partitioned, err
}

func (abci *ABCI) Query(reqQuery abcitypes.RequestQuery) (resQuery abcitypes.ResponseQuery) {
	log.WithFields(log.Fields{
		"Data":       reqQuery.Data,
		"stringData": string(reqQuery.Data),
	}).Info("query to ABCIApp")

	switch reqQuery.Path {
	case "GetIndexesFromVerifierID":
		var queryArgs getIndexesQuery
		err := bijson.Unmarshal(reqQuery.Data, &queryArgs)
		if err != nil {
			return abcitypes.ResponseQuery{Code: 10, Info: fmt.Sprintf("could not parse query into arguments: %v string ver: %s ", reqQuery.Data, string(reqQuery.Data))}
		}

		partitioned, err := GetAppKeyPartition(abci.broker, queryArgs.AppID)
		if err != nil {
			return abcitypes.ResponseQuery{Code: 10, Info: fmt.Sprintf("AppID %v not found", queryArgs.AppID)}
		}

		//TODO: Fix logs here
		tx := AssignmentTx(queryArgs)
		log.Debugf("Partitioned value in ABCI=%v", partitioned)
		verifierKey := getVerifierKey(tx, partitioned)
		log.WithFields(log.Fields{
			"verifierKey": string(verifierKey),
		}).Debug("GetIndexesFromVerifierID")
		keyIndexes, err := abci.retrieveVerifierToKeyIndex(verifierKey)
		if err != nil {
			return abcitypes.ResponseQuery{Code: 10, Info: fmt.Sprintf("val not found for query %v or data: %s, err: %v", reqQuery, string(reqQuery.Data), err)}
		}
		b, err := bijson.Marshal(keyIndexes)
		if err != nil {
			log.WithError(err).Error("error serialising KeyIndexes")
		}

		// uint -> string -> bytes, when receiving do bytes -> string -> uint
		return abcitypes.ResponseQuery{Code: 0, Value: []byte(b)}

	default:
		return abcitypes.ResponseQuery{Log: fmt.Sprintf("Invalid query path. Expected hash or tx, got %v", reqQuery.Path)}
	}
}

func (app *ABCI) retrieveVerifierToKeyIndex(verifierKey []byte) ([]big.Int, error) {
	b, err := app.db.Get(verifierKey)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, fmt.Errorf("retrieveVerifierToKeyIndex keyIndexes do not exist for verifier, and verifierID")
	}
	var res []big.Int
	err = bijson.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (app *ABCI) LoadState() (State, bool) {
	stateBytes, err := app.db.Get(stateKey)
	if err != nil {
		log.Error(err)
	}
	infoBytes, err := app.db.Get(appInfoKey)
	if err != nil {
		log.Error(err)
	}
	var state, prevState State
	var info AppInfo
	stateExists := false
	if len(stateBytes) != 0 {
		stateExists = true
		err := bijson.Unmarshal(stateBytes, &state)
		if err != nil {
			panic(err)
		}
		err = bijson.Unmarshal(stateBytes, &prevState)
		if err != nil {
			panic(err)
		}
		err = bijson.Unmarshal(infoBytes, &info)
		if err != nil {
			panic(err)
		}
	}
	app.state = &state
	app.prevState = &prevState
	app.info = &info
	return state, stateExists
}

func (abci *ABCI) SaveState() State {
	stateBytes, err := bijson.Marshal(abci.state)
	if err != nil {
		panic(err)
	}
	if err = abci.db.Set(stateKey, stateBytes); err != nil {
		log.Errorf("error during setting state, err=%s", err)
	}
	infoBytes, err := bijson.Marshal(abci.info)
	if err != nil {
		panic(err)
	}
	if err = abci.db.Set(appInfoKey, infoBytes); err != nil {
		log.Errorf("error during setting state, err=%s", err)
	}
	return *abci.state
}

func authenticateBftTx(tx []byte, broker *common.MessageBroker) (parsedTx DefaultBFTTxWrapper, senderDetails common.NodeDetails, err error) {
	err = bijson.Unmarshal(tx, &parsedTx)
	if err != nil {
		log.Errorf("could not unmarshal headers from tx: %v", err)
		return parsedTx, senderDetails, err
	}

	curEpoch := broker.ChainMethods().GetSelfEpoch()
	senderDetails, err = broker.ChainMethods().VerifyDataWithEpoch(parsedTx.PubKey, parsedTx.Signature, parsedTx.GetSerializedBody(), curEpoch)
	if err != nil {
		log.Errorf("bfttx not valid: error %v, tx %v", err, parsedTx)
		return parsedTx, senderDetails, err
	}
	return
}

func (app *ABCI) retrieveKeyMapping(keyIndex big.Int, curve common.CurveName) (*common.KeyAssignmentPublic, error) {
	b, err := app.db.Get(prefixKeyMapping([]byte(keyIndex.Text(16)), curve))
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("retrieveKeyMapping, KeyMapping do not exist for index")
	}
	var res common.KeyAssignmentPublic
	err = bijson.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (app *ABCI) getIndexesFromVerifierID(provider, userID, appID string, curve common.CurveName) (keyIndexes []big.Int, err error) {
	// struct for query args
	args := getIndexesQuery{AppID: appID, Provider: provider, UserID: userID, Curve: curve}
	argBytes, err := bijson.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("could not marshal query args error: %v", err)
	}
	reqQuery := types.RequestQuery{
		Data: argBytes,
		Path: "GetIndexesFromVerifierID",
	}

	res := app.Query(reqQuery)
	if res.Code == 10 {
		return nil, fmt.Errorf("failed to find keyindexes with response code: %v", res.Info)
	}
	err = bijson.Unmarshal(res.Value, &keyIndexes)
	if err != nil {
		return nil, fmt.Errorf("could not parse retrieved keyindex list for %s error: %v", string(res.Value), err)
	}
	return keyIndexes, nil
}

func (app *ABCI) storeKeyMapping(keyIndex big.Int, curve common.CurveName, assignment common.KeyAssignmentPublic) error {
	b, err := bijson.Marshal(assignment)
	if err != nil {
		return err
	}
	err = app.db.Set(prefixKeyMapping([]byte(keyIndex.Text(16)), curve), b)
	return err
}

func (app *ABCI) storeVerifierToKeyIndex(verifierKey []byte, keyIndexes []big.Int) error {
	b, err := bijson.Marshal(keyIndexes)
	if err != nil {
		return err
	}
	err = app.db.Set(verifierKey, b)
	return err
}

func prefixKeyMapping(key []byte, curve common.CurveName) []byte {
	if curve == common.ED25519 {
		return append(EDKeyMappingPrefixKey, key...)
	}
	return append(keyMappingPrefixKey, key...)
}

func MinOf(vars ...int) int {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}
