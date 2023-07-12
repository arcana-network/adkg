package aba

import (
	"crypto/rand"
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	keygencommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/common/aba"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var CoinInitMessageType common.DPSSMessageType = "coin_init_aba"

type CoinInitMessage struct {
	roundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	coinID  string
}

func NewCoinInitMessage(id common.DPSSRoundID, coinID string, curve *curves.Curve, sender int) (*common.DPSSMessage, error) {
	m := CoinInitMessage{
		roundID: id,
		kind:    CoinInitMessageType,
		curve:   curve,
		coinID:  coinID,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.roundID, m.kind, bytes)
	return &msg, nil
}

func (m *CoinInitMessage) Process(sender common.KeygenNodeDetails, self dpsscommon.DPSSParticipant) {

	gTilde := m.curve.Point.Hash([]byte(m.coinID))

	uJi := m.curve.Scalar.Zero()

	n, _, _ := self.Params(false)
	roundLeader, err := m.roundID.Leader()
	if err != nil {
		return
	}

	adkgid, err := common.DPSSIDFromRoundID(m.roundID)
	if err != nil {
		log.Debugf("Could not get leader from roundID, err=%s", err)
		return
	}

	sessionStore, complete := self.State().SessionStore.GetOrSetIfNotComplete(adkgid, dpsscommon.DefaultDPSSSession())
	if complete {
		log.Debugf("Keygen already complete: %s", adkgid)
		return
	}

	sessionStore.Lock()
	defer sessionStore.Unlock()

	TiSet := keygencommon.GetSetBits(n, sessionStore.T[int(roundLeader.Int64())])
	for _, i := range TiSet {
		share, err := m.curve.Scalar.SetBytes(sessionStore.S[i].Value)
		if err != nil {
			continue
		}
		uJi = uJi.Add(share)
	}
	// Create proof
	proof := generateProof(m.curve, gTilde, uJi)

	// Create message data
	gITilde := gTilde.Mul(uJi)
	data := make([]byte, 0)
	data = append(data, proof[:]...)
	data = append(data, gITilde.ToAffineCompressed()...)

	msg, err := NewCoinMessage(m.roundID, data, m.curve, self.ID())
	if err != nil {
		return
	}
	go self.Broadcast(false, *msg)
}

func generateProof(
	curve *curves.Curve,
	gTilde curves.Point,
	xI curves.Scalar) []byte {

	s := curve.NewScalar().Random(rand.Reader)

	g := curve.Point.Generator()

	h := g.Mul(s)
	hTilde := gTilde.Mul(s)

	gI := g.Mul(xI)
	gITilde := gTilde.Mul(xI)

	c := aba.Hash(g, h, gTilde, hTilde, gI, gITilde, curve)

	//z = s + xi * c
	z := xI.MulAdd(c, s)

	proof := make([]byte, 0)
	proof = append(proof, z.Bytes()...)                   // z is 32 bytes
	proof = append(proof, h.ToAffineCompressed()...)      // 33 bytes
	proof = append(proof, hTilde.ToAffineCompressed()...) // 33 bytes

	// z, h, hTilde
	return proof
}
