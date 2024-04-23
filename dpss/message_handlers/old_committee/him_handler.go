// Handler in charge of processing the messages that use an hyperinvertible
// matrix.
package old_committee

import (
	"math"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

// Type of the message for the HIM handler.
var DpssHimHandlerType string = "dpss_him"

// Represents a message for the hyper-invertible matrix computation in Line 103 of the DPSS protocol.
// We assume the following representation for the r_i shares: [ r_1 | r_2 | r_3 | ... | r_(B / (n - 2*t)) * (n - t) ].
// This representation is done in batches.
type DpssHimMessage struct {
	PSSRoundDetails common.PSSRoundDetails // Details of the PSS instance.
	Kind            string                 // Type of the message.
	CurveName       common.CurveName       // Curve that is being used in the protocol.
	Shares          []byte                 // Byte representation of the r_i shares.
}

// Creates a new message to handle the Line 103 of the DPSS protocol.
func NewDpssHimMessage(
	pssRoundDetails common.PSSRoundDetails,
	shares []curves.Scalar,
	hash []byte,
	curve common.CurveName,
) (*common.PSSMessage, error) {
	sharesBytes := sharing.CompressScalars(shares)
	msg := DpssHimMessage{
		PSSRoundDetails: pssRoundDetails,
		Kind:            DpssHimHandlerType,
		CurveName:       curve,
		Shares:          sharesBytes,
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

// Process the message that executes Line 103 of the DPSS protocol.
func (msg *DpssHimMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	// The message should sent by self.
	if sender.Index != self.Details().Index {
		log.WithFields(
			log.Fields{
				"SelfIndex":   self.Details().Index,
				"SenderIndex": sender.Index,
				"Message":     "Indexes should be equal",
			},
		).Error("DacssHimMessage: Process")
	}

	n, _, t := self.Params()

	// Number of old shares that will be transformed, i.e. B.
	numShares := msg.PSSRoundDetails.BatchSize

	// Matrix is square. The matrix will be of dimensions matrixSize x matrixSize.
	matrixSize := int(math.Ceil(float64(numShares)/float64(n-2*t))) * (n - t)

	// Decompress the shares comming from the MVBA protocol.
	shares, err := sharing.DecompressScalars(
		msg.Shares,
		common.CurveFromName(msg.CurveName),
		matrixSize,
	)

	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while decompressing the shares",
			},
		)
	}
	hiMatrix := sharing.CreateHIM(
		matrixSize,
		common.CurveFromName(msg.CurveName),
	)

	// Multiplies the HI matrix by the shares outputted by MVBA. We provide
	// (B / (n - 2*t)) * (n - t) shares and obtain again (B / (n - 2*t)) * (n - t)
	// shares with the following property:
	// (B / (n - 2*t)) * (n - t) - t of such shares are hidding truly random values.
	// But notice that (B / (n - 2*t)) * (n - t) - t >= B if and only if B >= n - 2t
	// which is a reasonable assumption.
	globalRandomR, err := sharing.HimMultiplication(hiMatrix, shares)

	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "error in HIM Matrix Multiplication",
			},
		).Error("HIMMessageHandler: Process")
	}

	// From the trully random values, we select B of them to be the masks for
	// the values {s_i}. The rValues are the ones that will be used to mask the
	// secrets.
	rValues := globalRandomR[:numShares]
	rValuesBytes := sharing.CompressScalars(rValues)

	reconstructionMsg, err := NewPreprocessBatchReconstructionMessage(
		msg.PSSRoundDetails,
		rValuesBytes,
		msg.CurveName,
	)
	if err != nil {
		log.WithFields(
			log.Fields{
				"Error":   err,
				"Message": "Error while creating the reconstruction message",
			},
		).Error("HIMMessageHandler: Process")
	}

	go self.Send(self.Details(), *reconstructionMsg)
}
