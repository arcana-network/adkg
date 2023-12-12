package rpc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/crypto"
	"github.com/arcana-network/dkgnode/keygen"
	"github.com/arcana-network/dkgnode/secp256k1"
	"github.com/arcana-network/dkgnode/telemetry"

	tronCrypto "github.com/TRON-US/go-eccrypto"
	"github.com/arcana-network/dkgnode/eventbus"
	fastjson "github.com/goccy/go-json"
	"github.com/osamingo/jsonrpc/v2"
	tmtypes "github.com/tendermint/tendermint/types"

	tmquery "github.com/tendermint/tendermint/libs/pubsub/query"

	logger "github.com/arcana-network/groot/logger"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var statLogger = logger.NewZapGlobal("dkg_statistics")

type (
	KeyLookupParams struct {
		PubKeyX big.Int `json:"pub_key_X"`
		PubKeyY big.Int `json:"pub_key_Y"`
	}
	VerifierLookupResult struct {
		Keys []VerifierLookupItem `json:"keys"`
	}
	VerifierLookupItem struct {
		KeyIndex string `json:"key_index"`
		PubKeyX  string `json:"pub_key_X"`
		PubKeyY  string `json:"pub_key_Y"`
		Address  string `json:"address"`
	}
	PublicKeyLookupHandler struct {
		eventBus eventbus.Bus
	}
	CommitmentRequestParams struct {
		MessagePrefix      string `json:"messageprefix"`
		TokenCommitment    string `json:"tokencommitment"`
		TempPubX           string `json:"temppubx"`
		TempPubY           string `json:"temppuby"`
		VerifierIdentifier string `json:"verifieridentifier"`
	}
	VerifierLookupParams struct {
		Provider string `json:"provider"`
		UserID   string `json:"user_id"`
		AppID    string `json:"app_id"`
	}
	KeyLookupResult struct {
		common.KeyAssignmentPublic
	}
	KeyAssignHandler struct {
		bus eventbus.Bus
	}
	HealthHandler struct {
	}
	ConnectionDetailsHandler struct {
		eventBus eventbus.Bus
	}
	ShareRequestItem struct {
		IDToken        string          `json:"id_token"`
		NodeSignatures []NodeSignature `json:"node_signatures"`
		UserID         string          `json:"user_id"`
		AppID          string          `json:"app_id"`
	}
	NodeSignature struct {
		Signature   string `json:"signature"`
		Data        string `json:"data"`
		NodePubKeyX string `json:"nodepubx"`
		NodePubKeyY string `json:"nodepuby"`
	}
	ShareRequestParams struct {
		Item []fastjson.RawMessage `json:"item"`
	}
	StoreKeyRequestParams struct {
		TxHash         string `json:"txHash"`
		EncryptedShare string `json:"encryptedShare"`
		PublicKey      string `json:"publicKey"`
		DID            string `json:"did"`
	}
	RetrieveKeyRequestParams struct {
		DID       string `json:"did"`
		TxHash    string `json:"txHash"`
		PublicKey string `json:"publicKey"`
	}
	ShareRequestResult struct {
		Keys []ShareRequestResultItem `json:"keys"`
	}
	PublicKeyHex struct {
		X string `json:"pub_x"`
		Y string `json:"pub_y"`
	}
	ShareRequestResultItem struct {
		Index     string                   `json:"index"`
		PublicKey PublicKeyHex             `json:"pub_key"`
		Verifiers map[string][]string      `json:"verifiers"`
		Share     []byte                   `json:"share"`
		Metadata  tronCrypto.EciesMetadata `json:"metadata"`
	}
	KeyAssignParams struct {
		Provider string `json:"provider"`
		UserID   string `json:"user_id"`
		AppID    string `json:"app_id"`
	}
	HealthParams struct {
	}
	KeyAssignItem struct {
		KeyIndex string  `json:"key_index"`
		PubKeyX  big.Int `json:"pub_key_X"`
		PubKeyY  big.Int `json:"pub_key_Y"`
		Address  string  `json:"address"`
	}
	KeyAssignResult struct {
		Keys []KeyAssignItem `json:"keys"`
	}
	HealthResult struct {
		Status string `json:"status"`
	}
	ValidatedNodeSignature struct {
		NodeSignature
		NodeIndex big.Int
	}
	KeyAssignment struct {
		common.KeyAssignmentPublic
		Share    []byte
		Metadata tronCrypto.EciesMetadata
	}
	CommitmentRequestResultData struct {
		MessagePrefix      string `json:"messageprefix"`
		TokenCommitment    string `json:"tokencommitment"`
		TempPubX           string `json:"temppubx"`
		TempPubY           string `json:"temppuby"`
		VerifierIdentifier string `json:"verifieridentifier"`
		TimeSigned         string `json:"timesigned"`
	}
	CommitmentRequestResult struct {
		Signature string `json:"signature"`
		Data      string `json:"data"`
		NodePubX  string `json:"nodepubx"`
		NodePubY  string `json:"nodepuby"`
	}
)

type AssignmentTx struct {
	Provider string
	UserID   string
	AppID    string
}

func (c *CommitmentRequestResultData) ToString() string {
	return strings.Join([]string{
		c.MessagePrefix,
		c.TokenCommitment,
		c.TempPubX,
		c.TempPubY,
		c.VerifierIdentifier,
		c.TimeSigned,
	}, common.Delimiter1)
}

func (nodeSig *NodeSignature) NodeValidation(nodeList []common.NodeReference) (*common.NodeReference, error) {
	var node *common.NodeReference
	for i, currNode := range nodeList {
		log.WithFields(log.Fields{
			"currNode": (currNode),
			"x":        currNode.PublicKey.X.Text(16),
			"y":        currNode.PublicKey.Y.Text(16),
			"nodeSig":  (nodeSig),
		}).Debug()
		if currNode.PublicKey.X.Text(16) == nodeSig.NodePubKeyX &&
			currNode.PublicKey.Y.Text(16) == nodeSig.NodePubKeyY {
			node = &nodeList[i]
		}
	}
	if node == nil {
		return nil, fmt.Errorf("Node not found for nodeSig: %v", *nodeSig)
	}
	recSig := crypto.HexToSig(nodeSig.Signature)
	var sig32 [32]byte
	copy(sig32[:], secp256k1.Keccak256([]byte(nodeSig.Data))[:32])
	recoveredSig := crypto.Signature{
		Raw:  recSig.Raw,
		Hash: sig32,
		R:    recSig.R,
		S:    recSig.S,
		V:    recSig.V - 27,
	}
	valid := crypto.IsValidSignature(*node.PublicKey, recoveredSig)
	if !valid {
		return nil, fmt.Errorf("Could not validate ecdsa signature %v", (recoveredSig))
	}

	return node, nil
}

func (c *CommitmentRequestResultData) FromString(data string) (bool, error) {
	dataArray := strings.Split(data, common.Delimiter1)

	if len(dataArray) < 6 {
		return false, errors.New("Could not parse commitmentrequestresultdata")
	}
	c.MessagePrefix = dataArray[0]
	c.TokenCommitment = dataArray[1]
	c.TempPubX = dataArray[2]
	c.TempPubY = dataArray[3]
	c.VerifierIdentifier = dataArray[4]
	c.TimeSigned = dataArray[5]
	return true, nil
}

func getVerifierClientID(broker *common.MessageBroker, appID, verifier string) (string, error) {
	cachedParams := broker.CacheMethods().RetrieveClientIDFromVerifier(appID, verifier)
	log.WithField("cachedClientID", cachedParams).Info("getVerifierClientID")
	if cachedParams == nil {
		params, err := broker.ChainMethods().GetClientIDViaVerifier(appID, verifier)
		log.WithField("clientID", params).Info("GetClientIDViaVerifier")

		if err != nil {
			return "", err
		}
		if params == nil {
			return "", errors.New("could not get clientID from specified appID")
		}
		broker.CacheMethods().StoreVerifierToClientID(appID, verifier, params)
		return params.ClientID, nil
	}
	return cachedParams.ClientID, nil
}

func (h KeyAssignHandler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {

	var p KeyAssignParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	err := assignKey(c, h.bus, p.Provider, p.UserID, p.AppID)
	if err != nil {
		return nil, err
	}

	telemetry.IncrementKeyAssigned()
	statLogger.Info("key_assign", logger.Field{"appId": p.AppID, "verifier": p.Provider})
	return KeyAssignResult{Keys: make([]KeyAssignItem, 0)}, nil
}

func (h HealthHandler) ServeJSONRPC(_ context.Context, _ *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	return HealthResult{Status: "Ok"}, nil
}

var requestTimer = 45

func assignKey(c context.Context, eventBus eventbus.Bus, provider string, userID string, appID string) *jsonrpc.Error {
	broker := common.NewServiceBroker(eventBus, "key_assign_handler")
	requestContext, requestContextCancel := context.WithTimeout(c, time.Duration(requestTimer)*time.Second)
	defer requestContextCancel()
	if userID == "" {
		return &jsonrpc.Error{Code: -32602, Message: "Input error", Data: "VerifierID is empty"}
	}

	if appID == "" {
		return &jsonrpc.Error{Code: -32602, Message: "Input error", Data: "AppID is empty"}
	}

	clientID, err := getVerifierClientID(broker, appID, provider)
	if err != nil || clientID == "" {
		return &jsonrpc.Error{Code: -32602, Message: "Input error", Data: "Invalid AppID"}
	}

	log.WithFields(log.Fields{
		"provider": provider,
		"userID":   userID,
		"appID":    appID,
	}).Info("BroadcastingAssignmentTx")

	keyIndexes, err := broker.ABCIMethods().GetIndexesFromVerifierID(provider, userID, appID)
	if err == nil {
		if len(keyIndexes) > 0 {
			return &jsonrpc.Error{Code: -32602, Message: "Input Error", Data: "Key is already assigned"}
		}
	}

	msg := AssignmentTx{UserID: userID, Provider: provider, AppID: appID}
	hash, err := broker.TendermintMethods().Broadcast(msg)
	if err != nil {
		return &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "Unable to broadcast: " + err.Error()}
	}
	log.WithField("hash", hash.String()).Info("BFT")
	rpcErr := waitForTransaction(hash, requestContext, broker)
	if rpcErr != nil {
		return rpcErr
	}
	return nil
}

func getTxStatus(broker *common.MessageBroker, hash []byte) *jsonrpc.Error {
	valid, err := broker.TendermintMethods().TxStatus(hash)
	if err != nil {
		return &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "unable to get tx status"}
	}
	if !valid {
		return &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "unable to get tx status"}
	}
	return nil
}

func waitForTransaction(hash common.Hash, requestContext context.Context,
	broker *common.MessageBroker) *jsonrpc.Error {

	query := tmquery.MustParse("tx.hash='" + hash.String() + "'")
	responseCh, err := broker.TendermintMethods().RegisterQuery(query.String())
	if err != nil {
		err := getTxStatus(broker, hash.Bytes())
		return err
	}
	var tmpJrpcErr jsonrpc.Error
	for {
		finished := false
		select {
		case e := <-responseCh:
			log.Info("GotResponseCh: WaitForTransaction()")
			err := checkTransactionResult(e, query)
			if err != nil {
				tmpJrpcErr = jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "Tx failed"}
				finished = true
				break
			}
			finished = true
		case <-requestContext.Done():
			tmpJrpcErr = jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "Request timeout"}
			finished = true
		}

		if finished {
			break
		}
	}
	if tmpJrpcErr != (jsonrpc.Error{}) {
		return &tmpJrpcErr
	}

	return nil
}

func checkTransactionResult(e []byte, query *tmquery.Query) error {
	var txResult = tmtypes.EventDataTx{}
	err := txResult.Unmarshal(e)
	if err != nil {
		log.WithFields(log.Fields{
			"txQuery": query.String(),
			"code":    txResult.Result.GetCode(),
		}).Error("CheckTransactionResult:Unmarshal()")
		return err
	}
	log.WithFields(log.Fields{
		"txQuery": query.String(),
		"code":    txResult.Result.GetCode(),
	}).Info("CheckTransactionResult")

	if txResult.Result.IsOK() {
		return nil
	}
	return errors.New("tx failed?")
}

func (h PublicKeyLookupHandler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {

	broker := common.NewServiceBroker(h.eventBus, "public_lookup_handler")
	var p VerifierLookupParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	if p.UserID == "" {
		return nil, &jsonrpc.Error{Code: -32602, Message: "Input error", Data: "VerifierID is empty"}
	}

	clientID, err := getVerifierClientID(broker, p.AppID, p.Provider)
	log.WithFields(log.Fields{
		"clientID": clientID,
		"err":      err,
	}).Info("In VerifierLookupHandler")
	if err != nil || clientID == "" {
		return nil, &jsonrpc.Error{Code: -32602, Message: "Input error", Data: "Invalid AppID"}
	}

	keyIndexes, err := broker.ABCIMethods().GetIndexesFromVerifierID(p.Provider, p.UserID, p.AppID)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32602, Message: "Input Error", Data: "Verifier + VerifierID has not yet been assigned"}
	}

	// prepare and send response
	result := VerifierLookupResult{}
	for _, index := range keyIndexes {
		publicKeyAss, err := broker.ABCIMethods().RetrieveKeyMapping(index)
		if err != nil {
			return nil, &jsonrpc.Error{Code: -32603, Message: fmt.Sprintf("Could not find address to key index error: %v", err)}
		}
		pk := publicKeyAss.PublicKey
		//form address eth
		addr := crypto.PointToEthAddress(pk)
		log.WithField("X", pk.X.Text(16)).Info("Lookup")
		log.WithField("Y", pk.Y.Text(16)).Info("Lookup")
		result.Keys = append(result.Keys, VerifierLookupItem{
			KeyIndex: index.Text(16),
			PubKeyX:  pk.X.Text(16),
			PubKeyY:  pk.Y.Text(16),
			Address:  addr.String(),
		})
	}

	statLogger.Info("key_lookup", logger.Field{"appId": p.AppID, "verifier": p.Provider})

	return result, nil
}

func (h KeyCommitmentRequestHandler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {

	broker := common.NewServiceBroker(h.bus, "commitment_request_handler")
	log.WithField("params", params).Debug("CommitmentRequestHandler")
	var p CommitmentRequestParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	tokenCommitment := p.TokenCommitment
	verifierIdentifier := p.VerifierIdentifier

	// check if message prefix is correct
	if p.MessagePrefix != "arc00" {
		log.WithField("params", (params)).Debug("incorrect message prefix")
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "Incorrect message prefix"}
	}

	found := broker.CacheMethods().TokenCommitExists(verifierIdentifier, tokenCommitment)
	if found {
		return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "Duplicate token found"}
	}

	tempPubKey := common.Point{
		X: *secp256k1.HexToBigInt(p.TempPubX),
		Y: *secp256k1.HexToBigInt(p.TempPubY),
	}

	broker.CacheMethods().RecordTokenCommit(verifierIdentifier, tokenCommitment, tempPubKey)

	// sign data
	commitmentRequestResultData := CommitmentRequestResultData{
		p.MessagePrefix,
		p.TokenCommitment,
		p.TempPubX,
		p.TempPubY,
		p.VerifierIdentifier,
		strconv.FormatInt(time.Now().Unix(), 10),
	}

	k := broker.ChainMethods().GetSelfPrivateKey()
	pk := broker.ChainMethods().GetSelfPublicKey()
	sig := crypto.SignData([]byte(commitmentRequestResultData.ToString()), crypto.BigIntToECDSAPrivateKey(k))
	res := CommitmentRequestResult{
		Signature: crypto.SigToHex(sig),
		Data:      commitmentRequestResultData.ToString(),
		NodePubX:  pk.X.Text(16),
		NodePubY:  pk.Y.Text(16),
	}
	log.WithField("CommitmentRequestResult", (res)).Debug()
	return res, nil
}

func (h KeyShareRequestHandler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {

	broker := common.NewServiceBroker(h.bus, "share_request_handler")
	epoch := broker.ChainMethods().GetCurrentEpoch()
	epochInfo, err := broker.ChainMethods().GetEpochInfo(epoch, false)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "Error occurred while current epoch"}
	}

	var p ShareRequestParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	threshold := int(epochInfo.K.Int64())
	allKeyIndexes := make(map[string]big.Int)    // String keyindex => keyindex
	allValidVerifierIDs := make(map[string]bool) // verifier + pcmn.Delimiter1 + verifierIDs => bool
	var pubKey common.Point

	nodeList := broker.ChainMethods().AwaitCompleteNodeList(epoch)

	// For Each VerifierItem we check its validity
	for _, rawItem := range p.Item {
		var parsedVerifierParams ShareRequestItem
		err = fastjson.Unmarshal(rawItem, &parsedVerifierParams)
		if err != nil {
			return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "Error occurred while parsing sharerequestitem"}
		}
		log.WithField("parsedVerifierParams", (parsedVerifierParams)).Debug("ShareRequestHandler:Unmarshal()")
		jsonMap := make(map[string]interface{})
		err = fastjson.Unmarshal(rawItem, &jsonMap)
		if err != nil {
			return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "Error occurred while parsing jsonmap"}
		}
		delete(jsonMap, "nodesignatures")
		redactedRawItem, err := fastjson.Marshal(jsonMap)
		if err != nil {
			return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "Error occurred while marshalling" + err.Error()}
		}
		verified, userID, err := broker.VerifierMethods().Verify((*bijson.RawMessage)(&redactedRawItem))
		if err != nil {
			return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "Error occurred while verifying params" + err.Error()}
		}

		if !verified {
			return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "Could not verify params"}
		}
		// Validate signatures
		var validSignatures []ValidatedNodeSignature
		for i := 0; i < len(parsedVerifierParams.NodeSignatures); i++ {
			nodeRef, err := parsedVerifierParams.NodeSignatures[i].NodeValidation(nodeList)
			if err == nil {
				validSignatures = append(validSignatures, ValidatedNodeSignature{
					parsedVerifierParams.NodeSignatures[i],
					*nodeRef.Index,
				})
			} else {
				log.WithError(err).Error("could not validate signatures")
			}
		}
		// Check if we have threshold number of signatures
		if len(validSignatures) < threshold {
			return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "Not enough valid signatures. Only " + strconv.Itoa(len(validSignatures)) + "valid signatures found."}
		}
		// Find common data string, and filter valid signatures on the wrong data
		// this is to prevent nodes from submitting valid signatures on wrong data
		commonDataMap := make(map[string]int)
		for i := 0; i < len(validSignatures); i++ {
			var commitmentRequestResultData CommitmentRequestResultData
			ok, err := commitmentRequestResultData.FromString(validSignatures[i].Data)
			if !ok || err != nil {
				log.WithField("ok", ok).WithError(err).Error("could not get commitmentRequestResultData from string")
			}
			stringData := strings.Join([]string{
				commitmentRequestResultData.MessagePrefix,
				commitmentRequestResultData.TokenCommitment,
				commitmentRequestResultData.VerifierIdentifier,
			}, common.Delimiter1)
			commonDataMap[stringData]++
		}
		var commonDataString string
		var commonDataCount int
		for k, v := range commonDataMap {
			if v > commonDataCount {
				commonDataString = k
			}
		}
		var validCommonSignatures []ValidatedNodeSignature
		for i := 0; i < len(validSignatures); i++ {
			var commitmentRequestResultData CommitmentRequestResultData
			ok, err := commitmentRequestResultData.FromString(validSignatures[i].Data)
			if !ok || err != nil {
				log.WithField("ok", ok).WithError(err).Error("could not get commitmentRequestResultData from string")
			}
			stringData := strings.Join([]string{
				commitmentRequestResultData.MessagePrefix,
				commitmentRequestResultData.TokenCommitment,
				commitmentRequestResultData.VerifierIdentifier,
			}, common.Delimiter1)
			if stringData == commonDataString {
				validCommonSignatures = append(validCommonSignatures, validSignatures[i])
			}
		}
		if len(validCommonSignatures) < threshold {
			return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "Not enough valid signatures on the same data, " + strconv.Itoa(len(validCommonSignatures)) + " valid signatures."}
		}

		commonData := strings.Split(commonDataString, common.Delimiter1)

		if len(commonData) != 3 {
			return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "Could not parse common data"}
		}

		commonTokenCommitment := commonData[1]
		commonVerifierIdentifier := commonData[2]

		// Lookup verifier and
		// verify that hash of token = tokenCommitment
		cleanedToken, err := broker.VerifierMethods().CleanToken(commonVerifierIdentifier, parsedVerifierParams.IDToken)
		if err != nil {
			return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "Error when cleaning token " + err.Error()}
		}
		if hex.EncodeToString(secp256k1.Keccak256([]byte(cleanedToken))) != commonTokenCommitment {
			return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "Token commitment and token are not compatible"}
		}

		keyIndexes, err := broker.ABCIMethods().GetIndexesFromVerifierID(commonVerifierIdentifier, userID, parsedVerifierParams.AppID)
		if err != nil {
			return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: fmt.Sprintf("share request could not retrieve keyIndexes: %v", err)}
		}

		// Add to overall list and valid verifierIDs
		for _, index := range keyIndexes {
			allKeyIndexes[index.Text(16)] = index
		}

		statLogger.Info("key_share_fetch", logger.Field{
			"appId": parsedVerifierParams.AppID,
			"Id":    userID,
		})

		pubKey = broker.CacheMethods().GetTokenCommitKey(commonVerifierIdentifier, commonTokenCommitment)

		allValidVerifierIDs[strings.Join([]string{parsedVerifierParams.UserID, userID}, common.Delimiter1)] = true
	}

	response := ShareRequestResult{}
	var allKeyIndexesSorted []big.Int
	for _, index := range allKeyIndexes {
		allKeyIndexesSorted = append(allKeyIndexesSorted, index)
	}
	sort.Slice(allKeyIndexesSorted, func(a, b int) bool {
		return allKeyIndexesSorted[a].Cmp(&allKeyIndexesSorted[b]) == -1
	})
	if len(allKeyIndexesSorted) > 0 {
		index := allKeyIndexesSorted[0]
		pubKeyAccessStructure, err := broker.ABCIMethods().RetrieveKeyMapping(index)
		log.WithFields(log.Fields{
			"publicX": pubKeyAccessStructure.PublicKey.X,
			"publicY": pubKeyAccessStructure.PublicKey.Y,
		}).Debug("public_key")
		if err != nil {
			telemetry.IncrementShareReqFail()
			return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: fmt.Sprintf("could not retrieve access structure: %v", err)}
		}

		si, _, err := broker.DBMethods().RetrieveCompletedShare(index)
		if err != nil {
			telemetry.IncrementShareReqFail()
			return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "could not retrieve completed share"}
		}

		log.WithFields(log.Fields{
			"share": si.Text(16),
			"index": index.Int64(),
		}).Debug("share_data")

		keyAssignment := KeyAssignment{
			Share: si.Bytes(),
		}
		log.WithField("AssignedKeyShare", keyAssignment).Debug("GetKeyShare")
		log.WithField("PublicKey", pubKeyAccessStructure).Debug("GetKeyShare")
		pubKeyHex := "04" + fmt.Sprintf("%064s", pubKey.X.Text(16)) + fmt.Sprintf("%064s", pubKey.Y.Text(16))
		encrypted, metadata, err := tronCrypto.Encrypt(pubKeyHex, keyAssignment.Share)
		if err != nil {
			telemetry.IncrementShareReqFail()
			return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: fmt.Sprintf("could not encrypt shares with err: %v", err)}
		}

		if metadata == nil {
			telemetry.IncrementShareReqFail()
			return nil, &jsonrpc.Error{Code: -32603, Message: "Internal error", Data: "could not encrypt shares, metadata nil"}
		}

		keyAssignment.Share = []byte(encrypted)
		keyAssignment.Metadata = *metadata

		response.Keys = append(response.Keys, ShareRequestResultItem{
			Index: pubKeyAccessStructure.Index.String(),
			PublicKey: PublicKeyHex{
				X: pubKeyAccessStructure.PublicKey.X.String(),
				Y: pubKeyAccessStructure.PublicKey.Y.String(),
			},
			Verifiers: pubKeyAccessStructure.Verifiers,
			Share:     keyAssignment.Share,
			Metadata:  *metadata,
		})
	}

	telemetry.IncrementShareReqSuccess()
	return response, nil
}

func (h ConnectionDetailsHandler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	log.WithField("ConnectionDetailsParams", string(*params)).Debug("connection details handler handling request")
	broker := common.NewServiceBroker(h.eventBus, "connection_details_handler")
	var p ConnectionDetailsParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	x := secp256k1.HexToBigInt(p.PubKeyX)
	y := secp256k1.HexToBigInt(p.PubKeyY)
	if !broker.ChainMethods().ValidateEpochPubKey(p.ConnectionDetailsMessage.NodeAddress, common.Point{X: *x, Y: *y}) {
		return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "Invalid pub key for provided epoch"}
	}
	valid, err := p.ConnectionDetailsMessage.Validate(*x, *y, p.Signature)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: fmt.Sprintf("Could not validate connection details message, err: %v", err)}
	}
	if !valid {
		return nil, &jsonrpc.Error{Code: -32602, Message: "Internal error", Data: "invalid connection details message"}
	}
	tmp2pConnection := broker.ChainMethods().GetTMP2PConnection()
	p2pConnection := broker.ChainMethods().GetP2PConnection()
	log.WithFields(log.Fields{
		tmp2pConnection: tmp2pConnection,
		p2pConnection:   p2pConnection,
	}).Debug("ConnectionDetailsHandler")
	return ConnectionDetailsResult{
		TMP2PConnection: tmp2pConnection,
		P2PConnection:   p2pConnection,
	}, nil
}

func (c *ConnectionDetailsMessage) Validate(pubKeyX, pubKeyY big.Int, sig []byte) (bool, error) {
	message := c.Message
	if message != "ConnectionDetails" {
		log.WithField("cMessage", c.Message).Error("message not ConnectionDetails")
		return false, errors.New("message is not ConnectionDetails")
	}
	timeSigned := c.Timestamp
	unixTime, err := strconv.ParseInt(timeSigned, 10, 64)
	if err != nil {
		log.WithError(err).Error("could not parse time signed")
		return false, err
	}
	if time.Unix(unixTime, 0).Add(10 * time.Minute).Before(time.Now()) {
		log.WithError(err).Error("signature expired")
		return false, err
	}
	return keygen.ECDSAVerify(c.String(), &common.Point{X: pubKeyX, Y: pubKeyY}, sig), nil
}
func (c *ConnectionDetailsMessage) String() string {
	return strings.Join([]string{c.Timestamp, c.Message, c.NodeAddress.String()}, common.Delimiter1)
}
