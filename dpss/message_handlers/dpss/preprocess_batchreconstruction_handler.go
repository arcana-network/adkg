package dpss

import (
	"math"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var PreprocessBatchRecMessageType string = "dpss_batchreconstruction_preprocess"

type PreprocessBatchRecMessage struct {
	PSSRoundDetails common.PSSRoundDetails // Details of the PSS instance.
	Kind            string                 // Type of the message.
	RValues         []byte                 // B compressed (random) scalars
	CurveName       common.CurveName
}

func NewPreprocessBatchReconstructionMessage(pssRoundDetails common.PSSRoundDetails, rValues []byte, curveName common.CurveName) (*common.PSSMessage, error) {
	msg := PreprocessBatchRecMessage{
		PSSRoundDetails: pssRoundDetails,
		Kind:            PreprocessBatchRecMessageType,
		RValues:         rValues,
		CurveName:       curveName,
	}

	msgBytes, err := bijson.Marshal(msg)
	if err != nil {
		return nil, err
	}

	pssMsg := common.CreatePSSMessage(
		pssRoundDetails,
		msg.Kind,
		msgBytes,
	)

	return &pssMsg, nil
}

func (msg *PreprocessBatchRecMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	self.State().ShareStore.Lock()
	defer self.State().ShareStore.Unlock()

	// This message should only be sent by self
	if !sender.IsEqual(self.Details()) {
		log.WithFields(
			log.Fields{
				"Sender":   sender,
				"Receiver": self.Details(),
				"Message":  "Sender and receiver should be the same",
			},
		).Error("PreprocessBatchRecMessage: Process")
		return
	}

	// locally compute (s_i+r_i) for i in B; shares s_i and shares of random values r_i
	numShares := len(self.State().ShareStore.OldShares)
	r_scalars, err := sharing.DecompressScalars(msg.RValues, common.CurveFromName(msg.CurveName), numShares)
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while decompressing the random values",
			},
		).Error("PreprocessBatchRecMessage: Process")

		return
	}
	n, _, t := self.Params()
	ai_values := make([]curves.Scalar, 0)
	for i := range numShares {
		ai_values = append(ai_values, self.State().ShareStore.OldShares[i].Add(r_scalars[i]))
	}
	// run B/(n-2t) BatchReconstruction
	batchSize := n - 2*t
	nrBatches := int(math.Ceil(float64(numShares) / float64(batchSize)))
	for i := range nrBatches {
		dpssBatchDetails := common.DPSSBatchRecDetails{
			PSSRoundDetails: msg.PSSRoundDetails,
			BatchRecCount:   i,
		}

		startIdx := i * batchSize
		endIdx := startIdx + batchSize
		if endIdx > (numShares - 1) {
			endIdx = numShares - 1
		}
		shareBatch := ai_values[startIdx:endIdx]
		compressedBatch := sharing.CompressScalars(shareBatch)

		// Create msg
		initMsg, err := NewInitRecMessage(
			dpssBatchDetails,
			compressedBatch,
			msg.CurveName)

		if err != nil {
			log.WithFields(
				log.Fields{
					"Error":   err,
					"Message": "Error while creating the InitRecMessage",
				},
			).Error("PreprocessBatchRecMessage: Process")
			return
		}

		// Send msg
		go self.ReceiveMessage(self.Details(), *initMsg)
	}
}