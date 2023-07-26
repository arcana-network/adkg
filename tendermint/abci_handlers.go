package tendermint

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/keyderivation"

	log "github.com/sirupsen/logrus"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/torusresearch/bijson"
)

var ErrKeyNotAvailable = errors.New("Key not available for assignment!")

func (abci *ABCI) validateTx(tx []byte, msgType byte, senderDetails common.KeygenNodeDetails, state *State) (bool, error) {
	log.WithFields(log.Fields{
		"msgType": msgType,
	}).Info("Got message in validateTx")
	switch msgType {
	case byte(1):
		var parsedTx AssignmentTx
		if err := bijson.Unmarshal(tx, &parsedTx); err != nil {
			log.WithError(err).Error("ValidateTx:Assignment")
			return false, err
		}

		if !state.KeyAvailable() {
			log.WithError(ErrKeyNotAvailable).Error("ValidateTx:Assignment")
			return false, ErrKeyNotAvailable
		}
		return true, nil

	case byte(2):
		log.Debug("Received keygen message in tendermint")
		var msg = common.DKGMessage{}
		err := bijson.Unmarshal(tx, &msg)
		if err != nil {
			log.WithError(err).Error("CheckTx:DKGMessage")
			return false, err
		}
		log.WithFields(log.Fields{
			"type": msg.Method,
		}).Info("CheckTx:DKGMessage")
		if msg.Method == keyderivation.PubKeygenType {
			var m keyderivation.PubKeygenMessage
			if err = bijson.Unmarshal(msg.Data, &m); err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("CheckTx:PubKeygenMessage.Unmarshal()")
				return false, err
			}

			adkgid, err := common.ADKGIDFromRoundID(m.RoundID)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("CheckTx:ADKGIDFromRoundID()")
				return false, err
			}

			// Check if key is already added
			if k, ok := abci.state.KeygenPubKeys[string(adkgid)]; ok {
				log.WithFields(log.Fields{"pubKey": k.Point.ToHex()}).Error("CheckTx:key already processed")
				return false, errors.New("Key already processed")
			}

			key := string(adkgid) + ":" + common.PointToEthAddress(m.PublicKey).Hex()

			// Otherwise try to get decision
			decision, ok := abci.state.KeygenDecisions[key]

			// If decision exists
			if ok {
				alreadyAdded := false
				for _, v := range decision.Nodes {
					if v == senderDetails.Index {
						alreadyAdded = true
					}
				}

				if alreadyAdded {
					log.Error("CheckTx: node already added")
					return false, errors.New("node already added to verfied list")
				}
			}
			return true, nil
		}
		return false, errors.New("tendermint received dkg message with unimplemented method:" + msg.Method)
	}
	return false, errors.New("tx type not recognized")
}

func (app *ABCI) resetKeyAssigns() {
	app.refreshFailedKeyAssigns()
	app.state.LastUnassignedIndex = app.state.LastCreatedIndex
}

func (app *ABCI) assignKey(pk common.KeyAssignmentPublic) {
	app.state.LastUnassignedIndex = uint(pk.Index.Int64()) + 1
	app.state.NewKeyAssignments = append(app.state.NewKeyAssignments, pk)
}

func (app *ABCI) refreshFailedKeyAssigns() {
	app.state.ConsecutiveFailedPubKeyAssigns = 0
}

func createKeyAssignment(tx AssignmentTx, i big.Int, pk common.Point) common.KeyAssignmentPublic {
	verifierMap := make(map[string][]string)
	verifierMap[tx.Provider] = []string{tx.UserID}
	newKeyMapping := common.KeyAssignmentPublic{
		Index:     i,
		PublicKey: pk,
		Verifiers: verifierMap,
	}
	return newKeyMapping
}

func getVerifierKey(tx AssignmentTx, partitioned bool) []byte {
	if partitioned {
		return getPartitionedKeyspace(tx.AppID, tx.UserID)
	}
	return getUnpartitionedKeyspace(tx.UserID)
}

func (abci *ABCI) ValidateAndUpdateAndTagBFTTx(bftTx []byte, msgType byte, senderDetails common.KeygenNodeDetails) (bool, *[]abcitypes.EventAttribute, error) {
	log.WithFields(log.Fields{
		"msgType":       msgType,
		"senderDetails": senderDetails,
	}).Debug("Got message in ValidateAndUpdateAndTagBFTTx")
	var tags []abcitypes.EventAttribute

	currEpoch := abci.broker.ChainMethods().GetCurrentEpoch()
	currEpochInfo, err := abci.broker.ChainMethods().GetEpochInfo(currEpoch, false)
	if err != nil {
		return false, &tags, fmt.Errorf("could not get current epoch with err: %v", err)
	}
	t := int(currEpochInfo.T.Int64())
	n := int(currEpochInfo.N.Int64())
	switch msgType {
	case byte(1): // Assignment tx

		log.Debug("Received assignment tx on tendermint")
		var parsedTx AssignmentTx
		if err := bijson.Unmarshal(bftTx, &parsedTx); err != nil {
			log.WithError(err).Error("AssignmentBFTTx failed")
			return false, &tags, err
		}

		if !abci.state.KeyAvailable() {
			return false, &tags, nil
		}

		var dkgID string
		var keyIndexes []big.Int
		var pk common.Point
		assignedKeyIndex := *big.NewInt(int64(abci.state.LastUnassignedIndex))
		for {
			pk, keyIndexes = abci.getKeyAssignment(assignedKeyIndex, parsedTx)
			dkgID = string(common.GenerateADKGID(assignedKeyIndex))
			if assignedKeyIndex.Cmp(new(big.Int).SetInt64(int64(abci.state.LastCreatedIndex))) > -1 {
				return false, &tags, errors.New("could not assign key, key not found")
			}
			if pk.X.Cmp(big.NewInt(0)) == 0 || pk.Y.Cmp(big.NewInt(0)) == 0 {
				assignedKeyIndex = *new(big.Int).Add(&assignedKeyIndex, new(big.Int).SetInt64(1))
			} else {
				break
			}
		}
		log.WithFields(log.Fields{
			"selected-key":     abci.state.KeygenPubKeys[dkgID],
			"total-key-in-map": len(abci.state.KeygenPubKeys),
		}).Debug("Assign Tx")

		abci.refreshFailedKeyAssigns()

		keyAssignment := createKeyAssignment(parsedTx, assignedKeyIndex, pk)

		err := abci.storeKeyMapping(assignedKeyIndex, keyAssignment)
		if err != nil {
			return false, &tags, fmt.Errorf("could not storeKeyMapping: %v ", err)
		}

		// Get aggregate login options here
		partitioned, err := getAppKeyPartition(abci.broker, parsedTx.AppID)
		if err != nil {
			return false, &tags, fmt.Errorf("AppID %v not found", parsedTx.AppID)
		}
		verifierKey := getVerifierKey(parsedTx, partitioned)

		log.WithFields(log.Fields{
			"verifierKey": string(verifierKey),
		}).Debug("KeyAssignment")
		err = abci.storeVerifierToKeyIndex(verifierKey, keyIndexes)

		if err != nil {
			return false, &tags, fmt.Errorf("could not storeVerifierToKeyIndex: %v ", err)
		}

		// increment counters
		abci.assignKey(keyAssignment)
		// clean up pubkeys generated and stored on-chain from keygen
		delete(abci.state.KeygenPubKeys, dkgID)
		// add final tags
		tags = []abcitypes.EventAttribute{
			{Key: []byte("assignment"), Value: []byte("1")},
		}
		return true, &tags, nil

	case byte(2): // keygen message
		var msg = common.DKGMessage{}
		err := bijson.Unmarshal(bftTx, &msg)
		if err != nil {
			log.Errorf("keygenMessage unmarshalling failed with error %s", err)
			return false, &tags, err
		}
		log.WithFields(log.Fields{
			"type": msg.Method,
		}).Debug("ValidateAndUpdateTx()")
		if msg.Method == keyderivation.PubKeygenType {
			var m keyderivation.PubKeygenMessage
			if err = bijson.Unmarshal(msg.Data, &m); err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("DeliverTx:PubKeygenMessage.Unmarshal()")
				return false, &tags, err
			}

			adkgid, err := common.ADKGIDFromRoundID(m.RoundID)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("DeliverTx:ADKGIDFromRoundID()")
				return false, &tags, err
			}

			keyIndex, err := adkgid.GetIndex()
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("DeliverTx:ADKGID.GetIndex()")
				return false, &tags, err
			}

			// Check if key is already added
			if k, ok := abci.state.KeygenPubKeys[string(adkgid)]; ok {
				log.WithFields(log.Fields{"pubKey": k.Point.ToHex()}).Error("key already processed")
				return false, &tags, errors.New("Key already processed")
			}

			key := string(adkgid) + ":" + common.PointToEthAddress(m.PublicKey).Hex()

			// Otherwise try to get decision
			decision, ok := abci.state.KeygenDecisions[key]

			// If decision exists
			if ok {
				alreadyAdded := false
				for _, v := range decision.Nodes {
					if v == senderDetails.Index {
						alreadyAdded = true
					}
				}

				if alreadyAdded {
					log.WithFields(log.Fields{
						"decision": decision,
						"adkgid":   adkgid,
					}).Error("node already added to verfied list")

					return false, &tags, errors.New("node already added to verfied list")
				}

				decision.Nodes = append(decision.Nodes, senderDetails.Index)
				abci.state.KeygenDecisions[key] = decision

			} else {
				// if decision does not exist, add to decision
				abci.state.KeygenDecisions[key] = KeygenDecision{
					Nodes: []int{senderDetails.Index},
				}
			}

			if len(abci.state.KeygenDecisions[key].Nodes) == n-t {
				log.Infof("abci.decisions_threshold=%d", n-t)

				log.Infof("Generated PK: index=%d, publickey=%s%s", keyIndex.Int64(), m.PublicKey.X.Text(16), m.PublicKey.Y.Text(16))
				err = abci.broker.DBMethods().StorePublicKeyToIndex(m.PublicKey, keyIndex)
				if err != nil {
					log.Error("Could not store completed keygen pubkey")
					return false, &tags, err
				}

				// Add to generated public key
				abci.state.KeygenPubKeys[string(adkgid)] = KeygenPubKey{
					ID:    string(adkgid),
					Point: m.PublicKey,
				}

				delete(abci.state.KeygenDecisions, key)

				index := keyIndex.Int64()

				if uint(index) > abci.state.LastCreatedIndex {
					abci.state.LastCreatedIndex = uint(index)
				}

				log.WithFields(log.Fields{
					"key":    m.PublicKey,
					"index":  abci.state.LastCreatedIndex,
					"adkgid": adkgid,
				}).Info("Key generated")
			}

			return true, &tags, nil
		}
		return false, &tags, errors.New("tendermint: unimplemented method:" + msg.Method)
	}
	return false, &tags, errors.New("Invalid tx type")
}

func (abci *ABCI) getKeyAssignment(assignedKeyIndex big.Int, parsedTx AssignmentTx) (common.Point, []big.Int) {
	// Get aggregate login options here
	partitioned, _ := getAppKeyPartition(abci.broker, parsedTx.AppID)
	verifierKey := getVerifierKey(parsedTx, partitioned)

	keyIndexes, err := abci.retrieveVerifierToKeyIndex(verifierKey)
	if err != nil {
		// Store verifier into db
		keyIndexes = []big.Int{assignedKeyIndex}
	} else {
		keyIndexes = append(keyIndexes, assignedKeyIndex)
	}
	sort.Slice(keyIndexes, func(a, b int) bool {
		return keyIndexes[a].Cmp(&keyIndexes[b]) == -1
	})
	id := string(common.GenerateADKGID(assignedKeyIndex))
	pk := abci.state.KeygenPubKeys[id].Point
	return pk, keyIndexes
}
