package aba

import (
	"crypto/rand"
	"encoding/json"
	"time"

	"github.com/arcana-network/dkgnode/common"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/common/aba"
	log "github.com/sirupsen/logrus"

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
	bytes, err := json.Marshal(m)
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
	roundLeader, err := m.RoundID.Leader()
	if err != nil {
		return
	}

	adkgid, err := common.ADKGIDFromRoundID(m.RoundID)
	if err != nil {
		log.Infof("Could not get leader from roundID, err=%s", err)
		return
	}

	sessionStore, complete := self.State().SessionStore.GetOrSetIfNotComplete(adkgid, common.DefaultADKGSession())
	if complete {
		log.Infof("Keygen already complete: %s", adkgid)
		return
	}

	var TiSet []int
	start := time.Now()

	for {
		sessionStore.Lock()

		TiSet := kcommon.GetSetBits(n, sessionStore.T[int(roundLeader.Int64())])

		log.WithFields(log.Fields{
			"self":   self.ID(),
			"sender": sender.Index,
			"round":  m.RoundID,
			"TiSet":  TiSet,
		}).Info("aba_coin")

		if len(TiSet) > 0 {
			sessionStore.Unlock()
			break
		}
		// Breakout if time since message received has exceeded 10s
		if time.Since(start) > time.Second*20 {
			sessionStore.Unlock()
			log.Errorf("timeout coin_init message, round=%s", m.RoundID)
			return
		}

		sessionStore.Unlock()

		time.Sleep(200 * time.Millisecond)
	}

	sessionStore.Lock()
	defer sessionStore.Unlock()

	TiSet = kcommon.GetSetBits(n, sessionStore.T[int(roundLeader.Int64())])
	for _, i := range TiSet {
		share, err := curve.Scalar.SetBytes(sessionStore.S[i].Value)
		if err != nil {
			continue
		}
		uJi = uJi.Add(share)
	}
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
	xI curves.Scalar, self common.DkgParticipant) []byte {

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
