package aba

import (
	"crypto/rand"
	"time"

	"github.com/arcana-network/dkgnode/common"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/common/aba"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

var CoinInitMessageType string = "aba_coin_init"

type CoinInitMessage struct {
	RoundID common.PSSRoundDetails
	Kind    string
	Curve   common.CurveName
	CoinID  string
}

func NewCoinInitMessage(id common.PSSRoundDetails, coinID string, curve common.CurveName) (*common.PSSMessage, error) {
	m := CoinInitMessage{
		id,
		CoinInitMessageType,
		curve,
		coinID,
	}
	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m CoinInitMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	curve := common.CurveFromName(m.Curve)

	gTilde := curve.Point.Hash([]byte(m.CoinID))

	uJi := curve.Scalar.Zero()

	n, _, _ := self.Params()
	roundLeader := m.RoundID.Dealer.Index

	pssID := m.RoundID.PssID

	pssState, complete := self.State().PSSStore.GetOrSetIfNotComplete(pssID)
	if complete {
		log.Infof("pss already complete: %s", pssID)
		return
	}

	var TiSet []int
	start := time.Now()

	for {
		pssState.Lock()

		TiSet := kcommon.GetSetBits(n, pssState.T[roundLeader])

		log.WithFields(log.Fields{
			"self":   self.Details().Index,
			"sender": sender.Index,
			"round":  m.RoundID,
			"TiSet":  TiSet,
		}).Info("aba_coin")

		if len(TiSet) > 0 {
			pssState.Unlock()
			break
		}
		// Breakout if time since message received has exceeded 20s
		if time.Since(start) > time.Second*20 {
			pssState.Unlock()
			log.Errorf("timeout coin_init message, round=%v", m.RoundID)
			return
		}

		pssState.Unlock()

		time.Sleep(200 * time.Millisecond)
	}

	pssState.Lock()
	defer pssState.Unlock()

	TiSet = kcommon.GetSetBits(n, pssState.T[roundLeader])
	for _, i := range TiSet {
		share, err := curve.Scalar.SetBytes(pssState.KeysetMap[0].ShareStore[i].Value)
		if err != nil {
			continue
		}
		uJi = uJi.Add(share)
	}
	// FIXME: Maybe something wrong here, in generate or verify
	// Create proof
	proof := generateProof(curve, gTilde, uJi, self)

	// Create message data
	gITilde := gTilde.Mul(uJi)
	data := make([]byte, 0)
	data = append(data, proof[:]...)
	data = append(data, gITilde.ToAffineCompressed()...)

	msg, err := NewCoinMessage(m.RoundID, data, m.Curve)
	if err != nil {
		return
	}

	go self.Broadcast(false, *msg)
}

func generateProof(
	curve *curves.Curve,
	gTilde curves.Point,
	xI curves.Scalar, self common.PSSParticipant) []byte {

	s := curve.NewScalar().Random(rand.Reader)

	g, _ := self.CurveParams(curve.Name)

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
