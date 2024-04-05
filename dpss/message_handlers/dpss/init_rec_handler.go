package dpss

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

var InitRecHandlerType string = "dpss_init_rec"

// Message that represents the initial step in the batch reconstruction protocol.
type InitRecMessage struct {
	DPSSBatchRecDetails common.DPSSBatchRecDetails // Information of the batch reconstruction round.
	ShareBatch          []byte                     // Share batch that will be reconstructed.
	Curve               common.CurveName           // Curve that is being used in the computations.
	Kind                string                     // Type of the message.
}

// NewInitRecMessage creates a new InitRecMessage.
func NewInitRecMessage(
	dpssBatchRecDetails common.DPSSBatchRecDetails,
	shareBatch []byte,
	curve common.CurveName,
) (*common.PSSMessage, error) {
	msg := InitRecMessage{
		Kind:       InitRecHandlerType,
		ShareBatch: shareBatch,
		Curve:      curve,
	}

	msgBytes, err := bijson.Marshal(msg)
	if err != nil {
		return nil, err
	}

	pssMsg := common.CreatePSSMessage(
		msg.DPSSBatchRecDetails.PSSRoundDetails,
		msg.Kind,
		msgBytes,
	)

	return &pssMsg, nil
}

// Process processes a received InitRecMessage
func (msg *InitRecMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	// Check if the sender and receiver are the same party.
	if !sender.IsEqual(self.Details()) {
		log.WithFields(
			log.Fields{
				"Sender":   sender,
				"Receiver": self.Details(),
				"Message":  "Sender and receiver should be the same",
			},
		).Error("InitRecMessage: Process")
		return
	}

	// Deserialize the shares.
	n, _, t := self.Params()
	batchSize := n - 2*t
	shareBatch, err := sharing.DecompressScalars(
		msg.ShareBatch,
		common.CurveFromName(msg.Curve),
		batchSize,
	)
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while deserializing the shares",
			},
		).Error("InitRecMessage: Process")
		return
	}

	// Get a Vandermonde matrix
	bigVandermondeMatrix := sharing.CreateHIM(n, common.CurveFromName(msg.Curve))
	vandermondeMatrix := sharing.GetFirstColumns(bigVandermondeMatrix, batchSize)

	uShares, err := sharing.HimMultiplication(vandermondeMatrix, shareBatch)
	if err != nil {
		log.WithFields(
			log.Fields{
				"Message": "Error while doing matrix multiplication",
				"Error":   err,
			},
		).Error("InitRecMessage: Process")
		return
	}

	for _, recvrNode := range self.Nodes(self.IsNewNode()) {
		shareBytes := uShares[recvrNode.Index-1].Bytes()
		pubRecMsg, err := NewPrivateRecMsg(
			msg.DPSSBatchRecDetails,
			msg.Curve,
			shareBytes,
		)
		if err != nil {
			log.WithFields(
				log.Fields{
					"Message": "Error while constructing the PSS Message",
					"Error":   err,
				},
			).Error("InitRecMessage: Process")
			return
		}

		go self.Send(recvrNode, *pubRecMsg)
	}
}
