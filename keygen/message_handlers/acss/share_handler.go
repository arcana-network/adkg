package acss

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"
	"github.com/arcana-network/dkgnode/telemetry"
)

var ShareMessageType string = "acss_share"

type ShareMessage struct {
	RoundID common.RoundID
	Kind    string
	Curve   common.CurveName
}

func NewShareMessage(id common.RoundID, curve common.CurveName) (*common.DKGMessage, error) {
	m := ShareMessage{
		id,
		ShareMessageType,
		curve,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}
func (m ShareMessage) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {
	log.Debugf("sender=%d, self=%d", sender.Index, self.ID())
	if sender.Index != self.ID() {
		return
	}

	state := self.State().KeygenStore
	defaultKeygen := &common.SharingStore{
		RoundID: m.RoundID,
		State: common.RBCState{
			Phase:         common.Initial,
			ReceivedReady: make(map[int]bool),
			ReceivedEcho:  make(map[int]bool),
		},
		EchoStore: make(map[string]*common.EchoStore),
		Started:   false,
	}
	keygen, complete := state.GetOrSetIfNotComplete(m.RoundID, defaultKeygen)

	if complete {
		log.Infof("Keygen already complete: %s", m.RoundID)
		return
	}
	keygen.Lock()
	defer keygen.Unlock()

	if keygen.Started {
		log.Warnf("Tried to start already started keygen: %s", m.RoundID)
		return
	}

	log.Infof("Starting keygen: %s", m.RoundID)

	keygen.Started = true

	// TODO FIX this has to be commented out for the tests to run
	telemetry.IncrementKeysGenerated()

	curve := common.CurveFromName(m.Curve)
	// Generate secret
	secret := acss.GenerateSecret(curve)

	// Generate share and commitments
	n, k, f := self.Params()

	log.Debugf("keygenid=%s;n=%d;k=%d;f=%d", m.RoundID, n, k, f)
	commitments, shares, err := acss.GenerateCommitmentAndShares(secret,
		uint32(k), uint32(n), curve)

	if err != nil {
		log.Errorf("acss.GenerateCommitmentAndShares():err=%v", err)
		return
	}
	// Compress commitments
	compressedCommitments := acss.CompressCommitments(commitments)

	// Init share map
	shareMap := make(map[uint32][]byte, n)

	// encrypt each share with node respective generated symmetric key, add to share map
	for _, share := range shares {
		nodePublicKey := self.PublicKey(int(share.Id))

		cipherShare, err := acss.Encrypt(share.Bytes(), nodePublicKey,
			self.PrivateKey())
		if err != nil {
			log.Errorf("acss.Encrypt():err=%v", err)
			return
		}
		shareMap[share.Id] = cipherShare
	}

	// Create message data
	messageData := &messages.MessageData{
		Commitments: compressedCommitments,
		ShareMap:    shareMap,
	}

	data, err := messageData.Serialize()
	if err != nil {
		log.Errorf("MessageData.Serialize():err=%v", err)
		return
	}

	// Create propose message & broadcast
	msg, err := NewAcssProposeMessage(m.RoundID, data, m.Curve)
	if err != nil {
		log.Errorf("NewAcssPropose:err=%v", err)
		return
	}

	go self.Broadcast(*msg)
}
