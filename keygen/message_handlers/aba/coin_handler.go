package aba

import (
	"crypto/sha256"
	"encoding/json"
	"strconv"
	"time"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/common"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/common/aba"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/keyderivation"
)

var CoinMessageType common.MessageType = "aba_coin"

type CoinMessage struct {
	RoundID common.RoundID
	Kind    common.MessageType
	Curve   common.CurveName
	Data    []byte
}

func NewCoinMessage(id common.RoundID, data []byte, curve common.CurveName) (*common.DKGMessage, error) {
	m := CoinMessage{
		id,
		CoinMessageType,
		curve,
		data,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m *CoinMessage) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {
	curve := common.CurveFromName(m.Curve)
	u, err := unpack(curve, m.Data)
	if err != nil {
		log.WithError(err).Error("Could not unpack data in aba_coin_share")
		return
	}
	n, k, f := self.Params(false)

	roundLeader, err := m.RoundID.Leader()
	if err != nil {
		log.WithError(err).Error("Could not get round leader in aba_coin_share")
		return
	}

	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.RoundID, common.DefaultABAStore())
	if complete {
		log.Infof("Keygen already complete: %s", m.RoundID)
		return
	}
	store.Lock()
	coinID := string(m.RoundID) + strconv.Itoa(store.GetRound())
	store.Unlock()

	gTilde := curve.Point.Hash([]byte(coinID))

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
			log.Errorf("timeout coin_share message, round=%s", m.RoundID)
			return
		}

		sessionStore.Unlock()

		time.Sleep(200 * time.Millisecond)
	}

	sessionStore.Lock()
	defer sessionStore.Unlock()

	TiSet = kcommon.GetSetBits(n, sessionStore.T[int(roundLeader.Int64())])

	if len(TiSet) == 0 {
		log.Infof("TiSet == 0 for round: %s, self: %d", m.RoundID, self.ID())
		return
	}

	gI := aba.DerivePublicKey(sender.Index, k, curve, TiSet, sessionStore.C)

	log.WithFields(log.Fields{
		"self":      self.ID(),
		"sender":    sender.Index,
		"round":     m.RoundID,
		"publicKey": gI.ToAffineCompressed(),
		"T":         sessionStore.T,
		"C":         sessionStore.C,
		"verified":  verify(u, gTilde, gI, curve, self),
	}).Debug("aba_coin_msg_before_verified")

	if verify(u, gTilde, gI, curve, self) {
		store.SetCoinShare(sender.Index, u.GiTilde)
	} else {
		log.Error("Coin share not verified, returning")
		return
	}

	coinShares := store.GetCoinShares()
	log.WithFields(log.Fields{
		"self":             self.ID(),
		"sender":           sender.Index,
		"round":            m.RoundID,
		"coinsharesLength": len(coinShares),
		"k":                k,
		"decisions":        sessionStore.Decisions,
	}).Debug("aba_coin")

	_, ok := sessionStore.Decisions[int(roundLeader.Int64())]

	if len(coinShares) == f+1 && !ok {
		identities := make([]int, 0)

		for i := range coinShares {
			identities = append(identities, i)
		}

		coeff, err := aba.LagrangeCoeffs(identities[0:k], curve)
		if err != nil {
			return
		}
		log.WithFields(log.Fields{
			"self":   self.ID(),
			"sender": sender.Index,
			"round":  m.RoundID,
			"coeff":  coeff,
		}).Info("aba_coin")

		g0Tilde := curve.Point.Identity()

		for i := range coeff {
			share := coinShares[i]
			if share != nil {
				g0Tilde = g0Tilde.Add(share.Mul(coeff[i]))
			}
		}

		roundLeader, err := m.RoundID.Leader()
		if err != nil {
			return
		}

		log.WithFields(log.Fields{
			"self":                 self.ID(),
			"sender":               sender.Index,
			"round":                m.RoundID,
			"sessionStoreDecision": sessionStore.Decisions,
			"Inequality(1==true)":  int(sha256.Sum256(g0Tilde.ToAffineCompressed())[31]) % 2,
		}).Info("aba_coin")

		if int(sha256.Sum256(g0Tilde.ToAffineCompressed())[31])%2 == 1 {
			sessionStore.Decisions[int(roundLeader.Int64())] = 1
			if !sessionStore.ABAComplete {
				sessionStore.ABAComplete = true
				adkgid, err := common.ADKGIDFromRoundID(m.RoundID)
				if err != nil {
					log.Info("Could not get ADKGIDf from roundID")
					return
				}
				index, _ := adkgid.GetIndex()
				log.Debugf("ADKGID=%d, decisions=%v,self=%d", index.Int64(), sessionStore.Decisions, self.ID())

				for i := 1; i <= n; i++ {
					if !Contains(sessionStore.ABAStarted, i) {
						go func(id int) {
							round := common.CreateRound(adkgid, id, "keyset")
							msg, err := NewInitMessage(round, 0, 0, m.Curve)
							if err != nil {
								log.WithError(err).Error("Could not create init message")
								return
							}
							self.ReceiveMessage(self.Details(), *msg)
						}(i)
					}
				}
			}
		} else {
			sessionStore.Decisions[int(roundLeader.Int64())] = 0
		}

		log.WithFields(log.Fields{
			"self":                 self.ID(),
			"sender":               sender.Index,
			"round":                m.RoundID,
			"Decision":             sessionStore.Decisions[int(roundLeader.Int64())],
			"sessionStoreDecision": sessionStore.Decisions,
			"CompleteCount":        len(sessionStore.Decisions),
			"ABAComplete":          sessionStore.ABAComplete,
		}).Info("aba_coin")

		// If all rounds ABA'd to 0 or 1, set ABA complete to true and start key derivation
		if n == len(sessionStore.Decisions) && !sessionStore.KeyderivationStarted {
			sessionStore.KeyderivationStarted = true
			msg, err := keyderivation.NewInitMessage(m.RoundID, m.Curve)
			if err != nil {
				return
			}
			go self.ReceiveMessage(self.Details(), *msg)
		}
	}
}

type Unpack struct {
	Z       curves.Scalar
	H       curves.Point
	HTilde  curves.Point
	GiTilde curves.Point
}

func unpack(curve *curves.Curve, msg []byte) (*Unpack, error) {
	d := Unpack{}

	z, err := curve.Scalar.SetBytes(msg[:32])
	if err != nil {
		return nil, err
	}
	d.Z = z

	h, err := curve.Point.FromAffineCompressed(msg[32:65])
	if err != nil {
		return nil, err
	}
	d.H = h

	hTilde, err := curve.Point.FromAffineCompressed(msg[65:98])
	if err != nil {
		return nil, err
	}
	d.HTilde = hTilde

	giTilde, err := curve.Point.FromAffineCompressed(msg[98:])
	if err != nil {
		return nil, err
	}
	d.GiTilde = giTilde

	return &d, nil
}

func verify(u *Unpack, gTilde, gI curves.Point, curve *curves.Curve, self common.DkgParticipant) bool {
	g, _ := self.CurveParams(curve.Name)

	cBar := aba.Hash(g, u.H, gTilde, u.HTilde, gI, u.GiTilde, curve)

	hBar := g.Mul(u.Z).Sub(gI.Mul(cBar))

	hTildeBar := gTilde.Mul(u.Z).Sub(u.GiTilde.Mul(cBar))

	if u.H.Equal(hBar) && u.HTilde.Equal(hTildeBar) {
		return true
	}
	return false
}
