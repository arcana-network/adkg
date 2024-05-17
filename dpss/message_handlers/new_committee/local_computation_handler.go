package new_committee

import (
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

func (msg *LocalComputationMsg) Process(sender common.NodeDetails, self common.PSSParticipant) {
	log.Info("LocalComputationMsg: Process")

	n, _, t := self.Params()

	state, _ := self.State().PSSStore.Get(msg.DPSSBatchRecDetails.PSSRoundDetails.PssID)

	state.Lock()
	defer state.Unlock()

	// Store hash(msg.Coefficients) -> 1
	// FIXME: Add function to hash [][]byte
	hash := ""

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

				"Error": err,

				"Message": "error in HIM Matrix Multiplication",
			},
		).Error("HIMMessageHandler: Process")

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
		// Validate ??
	}

	// FIXME: Add actual functions here
	// i => some unused index
	// i => share (refreshedShare)
	// userid(msg.userId[k]) => i

	// HOW??
	// i => public key
}
