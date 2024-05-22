package new_committee

import (
	"encoding/hex"
	"math"

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
	curveName           common.CurveName
	coefficients        [][]byte
	UserIds             []string
	T                   []int
}

func NewLocalComputationMsg(
	dpssBatchRecDetails common.DPSSBatchRecDetails,
	curve common.CurveName,
	coefficients [][]byte,
	T []int,
	userIds []string,

) (*common.PSSMessage, error) {
	msg := LocalComputationMsg{
		DPSSBatchRecDetails: dpssBatchRecDetails,
		Kind:                LocalComputationMessageType,
		curveName:           curve,
		coefficients:        coefficients,
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

func getHash(input [][]byte) string {
	var bytes []byte
	for _, b := range input {
		bytes = append(bytes, b...)
	}
	hash := hex.EncodeToString(common.Keccak256(bytes))
	return hash
}

func (msg *LocalComputationMsg) ProcessUserIDData(sender common.NodeDetails, self common.PSSParticipant) {
	_, _, t := self.Params()

	state, _ := self.State().PSSStore.Get(msg.DPSSBatchRecDetails.PSSRoundDetails.PssID)

	state.Lock()
	defer state.Unlock()

	// This whole thing is being done seperately
	// because all nodes might not have public key.
	// They might have been offline during assignment.
	for i, id := range msg.UserIds {
		if id != "" {
			val, ok := state.UserIDs[id]
			if !ok {
				state.UserIDs[id] = 0
			}

			if val == -1 {
				continue
			}

			state.UserIDs[id] = state.UserIDs[id] + 1

			if state.UserIDs[id] >= t+1 {
				// Assumption: All batch sizes are same except for the last batch
				// pssID = 1, i = 97 => index = 1*300 +97 = 397
				// batchSize := self.DefaultBatchSize()
				batchSize := 300
				pssIndex := common.GetIndexFromPSSID(msg.DPSSBatchRecDetails.PSSRoundDetails.PssID)
				keyIndex := (pssIndex * batchSize) + i
				// FIXME: this function needs to be created, where public key?
				self.StoreIndexToUser(keyIndex, id, publicKey, msg.curveName)
				state.UserIDs[id] = -1 // -1 to denote already done

			}
		}
	}
}
func (msg *LocalComputationMsg) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.Info("LocalComputationMsg: Process")

	n, _, t := self.Params()

	state, _ := self.State().PSSStore.Get(msg.DPSSBatchRecDetails.PSSRoundDetails.PssID)

	state.Lock()
	defer state.Unlock()

	go msg.ProcessUserIDData(sender, self)

	hash := getHash(msg.coefficients)
	_, ok := state.LocalComp[hash]
	if !ok {
		state.LocalComp[hash] = 0
	}

	state.LocalComp[hash] = state.LocalComp[hash] + 1
	// Check if t + 1 have sent similar message
	if state.LocalComp[hash] != t+1 {
		return
	}

	numShares := msg.DPSSBatchRecDetails.PSSRoundDetails.BatchSize

	matrixSize := int(math.Ceil(float64(numShares)/float64(n-2*t))) * (n - t)

	hiMatrix := sharing.CreateHIM(matrixSize, common.CurveFromName(msg.curveName))

	curve := common.CurveFromName(msg.curveName)

	// msg.coefficients = s + r

	shares := state.GetSharesFromT(msg.T, numShares, curve)

	globalRandomR, err := sharing.HimMultiplication(hiMatrix, shares)

	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "error in HIM Matrix Multiplication",
			},
		).Error("LocalCompMessageHandler: Process")
	}

	rPrimeValues := globalRandomR[:numShares]

	refreshedShares := make([]curves.Scalar, 0)

	for i, sr := range msg.coefficients {
		sri, err := curve.Scalar.SetBytes(sr)
		if err != nil {
			refreshedShares = append(refreshedShares, curve.Scalar.Zero())
			continue
		}
		newShare := sri.Sub(rPrimeValues[i])
		refreshedShares = append(refreshedShares, newShare) // ((s + r) - r')
	}

	// Assumption: All batch sizes are same except for the last batch
	// pssID = 1, i = 97 => index = 1*300 +97 = 397
	// batchSize := self.DefaultBatchSize()
	batchSize := 300
	pssIndex := common.GetIndexFromPSSID(msg.DPSSBatchRecDetails.PSSRoundDetails.PssID)
	for i, share := range refreshedShares {
		keyIndex := (pssIndex * batchSize) + i
		self.StoreShare(keyIndex, share, msg.curveName)
	}
}
