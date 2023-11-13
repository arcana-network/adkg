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
			log.WithError(err).Error("CheckTx:Assignment")
			return false, err
		}

		if !state.KeyAvailable(parsedTx.Curve) {
			log.WithError(ErrKeyNotAvailable).Error("CheckTx:Assignment")
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

			key := string(adkgid)

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

func (app *ABCI) assignKey(pk common.KeyAssignmentPublic, curve common.CurveName) {
	if curve == common.SECP256K1 {
		app.state.LastUnassignedIndex = uint(pk.Index.Int64()) + 1
	} else {
		app.state.C25519State.LastUnassignedIndex = uint(pk.Index.Int64()) + 1
	}
	app.state.NewKeyAssignments = append(app.state.NewKeyAssignments, pk)
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
		return getPartitionedKeyspace(tx.AppID, tx.UserID, tx.Curve)
	}
	return getUnpartitionedKeyspace(tx.UserID, tx.Curve)
}

func (abci *ABCI) ValidateAndUpdateAndTagBFTTx(bftTx []byte, msgType byte, senderDetails common.KeygenNodeDetails) (bool, *[]abcitypes.EventAttribute, error) {
	var tags []abcitypes.EventAttribute

	currEpoch := abci.broker.ChainMethods().GetCurrentEpoch()
	currEpochInfo, err := abci.broker.ChainMethods().GetEpochInfo(currEpoch, false)
	if err != nil {
		return false, &tags, fmt.Errorf("could not get current epoch with err: %v", err)
	}
	threshold := int(currEpochInfo.K.Int64())

	switch msgType {
	case byte(1): // Assignment tx

		log.Debug("Received assignment tx on tendermint")
		var tx AssignmentTx
		if err := bijson.Unmarshal(bftTx, &tx); err != nil {
			log.WithError(err).Error("AssignmentBFTTx failed")
			return false, &tags, err
		}

		if !abci.state.KeyAvailable(tx.Curve) {
			abci.state.ConsecutiveFailedPubKeyAssigns++
			return false, &tags, errors.New("key not available!")
		}

		dkgID, assignIndex, err := findUnassignedKey(abci, tx.Curve)
		if err != nil {
			return false, &tags, fmt.Errorf("could not find key: %v ", err)
		}

		abci.state.ConsecutiveFailedPubKeyAssigns = 0

		keyIndexes := abci.getKeyAssignment(*assignIndex, tx)
		pk := abci.state.KeygenPubKeys[dkgID].Point

		keyAssignment := createKeyAssignment(tx, *assignIndex, pk)

		err = abci.storeKeyMapping(*assignIndex, tx.Curve, keyAssignment)
		if err != nil {
			return false, &tags, fmt.Errorf("could not storeKeyMapping: %v ", err)
		}

		// Get aggregate login options here
		partitioned, err := GetAppKeyPartition(abci.broker, tx.AppID)
		if err != nil {
			return false, &tags, fmt.Errorf("AppID %v not found", tx.AppID)
		}
		verifierKey := getVerifierKey(tx, partitioned)

		err = abci.storeVerifierToKeyIndex(verifierKey, keyIndexes)

		if err != nil {
			return false, &tags, fmt.Errorf("could not storeVerifierToKeyIndex: %v ", err)
		}

		// increment counters
		abci.assignKey(keyAssignment, tx.Curve)
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

			key := string(adkgid)

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

			log.Infof("abci.decisions: current=%d, threshold=%d", len(abci.state.KeygenDecisions[key].Nodes), threshold)
			if len(abci.state.KeygenDecisions[key].Nodes) == threshold {
				curve, _ := adkgid.GetCurve()
				log.Infof("Generated PK: index=%d, publickey=%s%s", keyIndex.Int64(), m.PublicKey.X.Text(16), m.PublicKey.Y.Text(16))
				err = abci.broker.DBMethods().StorePublicKeyToIndex(m.PublicKey, keyIndex, curve)
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

				if curve == common.SECP256K1 {
					if uint(index) > abci.state.LastCreatedIndex {
						abci.state.LastCreatedIndex = uint(index)
					}
				} else if curve == common.ED25519 {
					if uint(index) > abci.state.C25519State.LastCreatedIndex {
						abci.state.C25519State.LastCreatedIndex = uint(index)
					}
				}

				_ = abci.broker.KeygenMethods().Cleanup(adkgid)

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

func findUnassignedKey(abci *ABCI, curve common.CurveName) (string, *big.Int, error) {
	var dkgID string
	var pk common.Point

	lastUnassignedIndex := abci.state.LastUnassignedIndex
	lastCreatedIndex := abci.state.LastCreatedIndex

	if curve == common.ED25519 {
		lastUnassignedIndex = abci.state.C25519State.LastUnassignedIndex
		lastCreatedIndex = abci.state.C25519State.LastCreatedIndex
	}

	assignedKeyIndex := *big.NewInt(int64(lastUnassignedIndex))
	for {
		dkgID = string(common.NewADKGID(assignedKeyIndex, curve))
		pk = abci.state.KeygenPubKeys[dkgID].Point
		if assignedKeyIndex.Cmp(new(big.Int).SetInt64(int64(lastCreatedIndex))) > -1 {
			return dkgID, nil, errors.New("could not assign key, key not found")
		}
		if pk.X.Cmp(big.NewInt(0)) == 0 || pk.Y.Cmp(big.NewInt(0)) == 0 {
			assignedKeyIndex = *new(big.Int).Add(&assignedKeyIndex, new(big.Int).SetInt64(1))
		} else {
			break
		}
	}

	return dkgID, &assignedKeyIndex, nil
}
func (abci *ABCI) getKeyAssignment(assignedKeyIndex big.Int, parsedTx AssignmentTx) []big.Int {
	partitioned, _ := GetAppKeyPartition(abci.broker, parsedTx.AppID)
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
	return keyIndexes
}
