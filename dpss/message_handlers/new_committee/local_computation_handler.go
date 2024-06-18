package new_committee

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var LocalComputationMessageType string = "dpss_local_computation"

type LocalComputationMsg struct {
	DPSSBatchRecDetails common.DPSSBatchRecDetails
	Kind                string
	CurveName           common.CurveName
	Coefficients        []string
	UserIds             []string
	T                   []int
}

func NewLocalComputationMsg(
	dpssBatchRecDetails common.DPSSBatchRecDetails,
	curve common.CurveName,
	coefficients []string,
	T []int,
	userIds []string,

) (*common.PSSMessage, error) {
	msg := LocalComputationMsg{
		DPSSBatchRecDetails: dpssBatchRecDetails,
		Kind:                LocalComputationMessageType,
		CurveName:           curve,
		Coefficients:        coefficients,
		UserIds:             userIds,
		T:                   T,
	}

	msgBytes, err := bijson.Marshal(msg)
	if err != nil {
		return nil, err
	}

	pssMessage := common.CreatePSSMessage(
		msg.DPSSBatchRecDetails.PSSRoundDetails,
		msg.Kind,
		msgBytes,
	)

	return &pssMessage, nil
}

func getHash(input []curves.Scalar) string {
	var bytes []byte
	for _, b := range input {
		bytes = append(bytes, b.Bytes()...)
	}
	hash := hex.EncodeToString(common.Keccak256(bytes))
	return hash
}

func (msg *LocalComputationMsg) ProcessUserIDData(sender common.NodeDetails, self common.PSSParticipant) {

	state, _ := self.State().PSSStore.GetOrSetIfNotComplete(msg.DPSSBatchRecDetails.PSSRoundDetails.PssID)

	state.Lock()
	defer state.Unlock()

	n, _, t := self.OldParams()
	batchRecSize := n - 2*t

	// This whole thing is being done seperately
	// because all nodes might not have public key.
	// They might have been offline during assignment.
	for i, id := range msg.UserIds {
		// positioning inside the batch rec batches
		j := (batchRecSize * msg.DPSSBatchRecDetails.BatchRecCount) + i
		if id != "" {
			_, ok := state.UserIDs[id]
			if !ok {
				state.UserIDs[id] = &common.LocalComputationUserIDS{
					Count: 0,
					ID:    j,
				}
			}

			if state.UserIDs[id].Count == -1 {
				continue
			}

			state.UserIDs[id].Count = state.UserIDs[id].Count + 1

			if state.UserIDs[id].Count >= t+1 {
				// Assumption: All batch sizes are same except for the last batch
				// pssID = 1, i = 97 => index = 1 * 300 + 97 = 397
				batchSize := self.DefaultBatchSize()
				pssIndex := common.GetIndexFromPSSID(msg.DPSSBatchRecDetails.PSSRoundDetails.PssID)
				keyIndex := (pssIndex * batchSize) + state.UserIDs[id].ID
				// FIXME: this function needs to be created, where public key and appID?
				self.StoreIndexToUser(keyIndex, id, msg.CurveName)
				state.UserIDs[id].Count = -1 // -1 to denote already done

			}
		}
	}
}
func (msg *LocalComputationMsg) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.Info("LocalComputationMsg: Process")

	n, _, t := self.OldParams()

	batchRecSize := n - 2*t
	nrBatches := int(math.Ceil(float64(msg.DPSSBatchRecDetails.PSSRoundDetails.BatchSize) / float64(batchRecSize)))

	batchCount := msg.DPSSBatchRecDetails.BatchRecCount
	log.Debugf("LocalComputationProcess:: Sender=%d, BatchRecCount=%d, size=%d", sender.Index, msg.DPSSBatchRecDetails.BatchRecCount, len(msg.Coefficients))
	state, _ := self.State().PSSStore.GetOrSetIfNotComplete(msg.DPSSBatchRecDetails.PSSRoundDetails.PssID)

	state.Lock()
	defer state.Unlock()

	curve := common.CurveFromName(msg.CurveName)

	// id => sender+batchRecCount
	id := fmt.Sprintf("%d:%d", sender.Index, msg.DPSSBatchRecDetails.BatchRecCount)
	if state.LocalCompReceived[id] {
		// Duplicate Message
		return
	}

	state.LocalCompReceived[id] = true

	go msg.ProcessUserIDData(sender, self)

	numShares := msg.DPSSBatchRecDetails.PSSRoundDetails.BatchSize
	alpha := int(math.Ceil(float64(numShares) / float64(n-2*t)))

	coefficients := make([]curves.Scalar, 0)

	for _, m := range msg.Coefficients {
		v, _ := new(big.Int).SetString(m, 16)
		s, _ := curve.Scalar.SetBigInt(v)
		coefficients = append(coefficients, s)
	}

	hash := getHash(coefficients)

	_, ok := state.LocalComp[batchCount]
	if !ok {
		state.LocalComp[batchCount] = &common.LocalComputation{
			Hash:  hash,
			Count: 0,
		}
	}

	// If the first sender with this batchcount is wrong then upcoming
	// correct ones might get ignored. maybe need to keep multiple values
	// seems weird though. FIXME??
	if state.LocalComp[batchCount].Hash != hash {
		return
	}

	state.LocalComp[batchCount].Count = state.LocalComp[batchCount].Count + 1

	// Check if t + 1 have sent similar message for a particular batch count
	if state.LocalComp[batchCount].Count < t+1 {
		return
	}

	state.LocalComp[batchCount].Coefficients = msg.Coefficients

	// Check if all the batches have been completed
	for i := range nrBatches {
		val, ok := state.LocalComp[i]
		if !ok {
			return
		}
		if val.Count < t+1 {
			return
		}
	}

	matrixSize := int(math.Ceil(float64(numShares)/float64(n-2*t))) * (n - t)
	hiMatrix := sharing.CreateHIM(matrixSize, common.CurveFromName(msg.CurveName))

	log.Debugf("msg.T=%v, self=%d, matrixSize=%d", msg.T, self.Details().Index, matrixSize)
	shares, err := state.GetSharesFromT(msg.T, alpha, curve)
	if err != nil {
		// FIXME: Add waiting for shares.
		log.Errorf("Error: LocalComputation: GetShares: %s", err)
		return
	}
	log.Debugf("msg.T=%v, self=%d, matrixSize=%d, shareSize=%d", msg.T, self.Details().Index, matrixSize, len(shares))

	globalRandomR, err := sharing.HimMultiplication(hiMatrix, shares)
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "error in HIM Matrix Multiplication",
			},
		).Error("LocalCompMessageHandler: Process")
		return
	}

	rPrimeValues := globalRandomR[:numShares]

	combinedCoefficients := []curves.Scalar{}
	for i := range nrBatches {
		val := state.LocalComp[i]
		for _, v := range val.Coefficients {
			b, ok := new(big.Int).SetString(v, 16)
			if !ok {
				return
			}
			s, err := curve.Scalar.SetBigInt(b)
			if err != nil {
				return
			}
			combinedCoefficients = append(combinedCoefficients, s)
		}
	}

	refreshedShares := []curves.Scalar{}
	for i, sr := range combinedCoefficients {
		newShare := sr.Sub(rPrimeValues[i])
		refreshedShares = append(refreshedShares, newShare) // ((s + r) - r')
	}

	// Assumption: All batch sizes are same except for the last batch
	// pssID = 1, i = 97 => index = 1 * 299 + 97 = 397
	defaultBatchSize := self.DefaultBatchSize()
	log.Debugf("DefaultBatchSize=%d, refreshedShareSize=%d", defaultBatchSize, len(refreshedShares))
	pssIndex := common.GetIndexFromPSSID(msg.DPSSBatchRecDetails.PSSRoundDetails.PssID)
	for i, share := range refreshedShares {
		keyIndex := (pssIndex * defaultBatchSize) + i
		log.Debugf("self=%d, keyIndex=%d, share=%v", self.Details().Index, keyIndex, share)
		self.StoreShare(keyIndex, share, msg.CurveName)
	}
}
