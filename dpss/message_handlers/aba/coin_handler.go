package aba

import (
	"crypto/sha256"
	"encoding/json"
	"strconv"
	"time"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/him"
	keygencommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/common/aba"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var CoinMessageType common.DPSSMessageType = "coin_aba"

type CoinMessage struct {
	RoundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	data    []byte
}

func NewCoinMessage(id common.DPSSRoundID, data []byte, curve *curves.Curve, sender int) (*common.DPSSMessage, error) {
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

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m *CoinMessage) Process(sender common.KeygenNodeDetails, self dpsscommon.DPSSParticipant) {
	u, err := unpack(m.curve, m.data)
	if err != nil {
		log.WithError(err).Info("Could not unpack data in aba_coin_share")
		return
	}
	n, k, _ := self.Params(false)

	roundLeader, err := m.RoundID.Leader()
	if err != nil {
		log.WithError(err).Info("Could not get round leader in aba_coin_share")
		return
	}

	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.RoundID, common.DefaultABAStore())
	if complete {
		log.Debugf("Keygen already complete: %s", m.RoundID)
		return
	}
	store.Lock()
	coinID := string(m.RoundID) + strconv.Itoa(store.Round())
	store.Unlock()

	gTilde := m.curve.Point.Hash([]byte(coinID))

	adkgid, err := common.DPSSIDFromRoundID(m.RoundID)
	if err != nil {
		log.Debugf("Could not get leader from roundID, err=%s", err)
		return
	}
	sessionStore, complete := self.State().SessionStore.GetOrSetIfNotComplete(adkgid, dpsscommon.DefaultDPSSSession())
	if complete {
		log.Debugf("Keygen already complete: %s", adkgid)
		return
	}

	var TiSet []int
	start := time.Now()

	for {
		sessionStore.Lock()

		TiSet := keygencommon.GetSetBits(n, sessionStore.T[int(roundLeader.Int64())])

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
		if time.Since(start) > time.Second*10 {
			sessionStore.Unlock()
			return
		}

		sessionStore.Unlock()

		time.Sleep(200 * time.Millisecond)
	}

	sessionStore.Lock()
	defer sessionStore.Unlock()

	TiSet = keygencommon.GetSetBits(n, sessionStore.T[int(roundLeader.Int64())])

	if len(TiSet) == 0 {
		log.Debugf("TiSet == 0 for round: %s, self: %d", m.RoundID, self.ID())
		return
	}

	gI := aba.DerivePublicKey(sender.Index, k, m.curve, TiSet, sessionStore.C)

	log.WithFields(log.Fields{
		"self":      self.ID(),
		"sender":    sender.Index,
		"round":     m.RoundID,
		"publicKey": gI.ToAffineCompressed(),
		"T":         sessionStore.T,
		"C":         sessionStore.C,
		"verified":  verify(u, gTilde, gI, m.curve),
	}).Debug("aba_coin_msg_before_verified")

	if verify(u, gTilde, gI, m.curve) {
		store.SetCoinShare(sender.Index, u.GiTilde)
	}

	coinShares := store.CoinShares()
	log.WithFields(log.Fields{
		"self":             self.ID(),
		"sender":           sender.Index,
		"round":            m.RoundID,
		"coinsharesLength": len(coinShares),
		"k":                k,
		"decisions":        sessionStore.Decisions,
	}).Debug("aba_coin")

	_, ok := sessionStore.Decisions[int(roundLeader.Int64())]

	if len(coinShares) >= k && !ok {
		identities := make([]int, 0)

		for i := range coinShares {
			identities = append(identities, i)
		}

		coeff, err := aba.LagrangeCoeffs(identities[0:k], m.curve)
		if err != nil {
			return
		}
		log.WithFields(log.Fields{
			"self":   self.ID(),
			"sender": sender.Index,
			"round":  m.RoundID,
			"coeff":  coeff,
		}).Info("aba_coin")

		g0Tilde := m.curve.Point.Identity()

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
				adkgid, err := common.DPSSIDFromRoundID(m.RoundID)
				if err != nil {
					log.Debug("Could not get ADKGIDf from roundID")
					return
				}
				index, _ := adkgid.GetIndex()
				log.Debugf("ADKGID=%d, decisions=%v,self=%d", index.Int64(), sessionStore.Decisions, self.ID())

				for i := 1; i <= n; i++ {
					if !Contains(sessionStore.ABAStarted, i) {
						go func(id int) {
							round := common.CreateDPSSRound(adkgid, id, "keyset")
							msg, err := NewInitMessage(round, 0, 0, m.curve)
							if err != nil {
								return
							}
							self.ReceiveMessage(*msg)
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
		if n == len(sessionStore.Decisions) && sessionStore.ABAComplete {
			msg, err := him.NewInitMessage(m.RoundID, false, m.curve)
			if err != nil {
				log.Error(err)
				return
			}
			go self.ReceiveMessage(*msg)
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

func verify(u *Unpack, gTilde, gI curves.Point, curve *curves.Curve) bool {
	g := curve.Point.Generator()
	cBar := aba.Hash(g, u.H, gTilde, u.HTilde, gI, u.GiTilde, curve)

	hBar := g.Mul(u.Z).Sub(gI.Mul(cBar))

	hTildeBar := gTilde.Mul(u.Z).Sub(u.GiTilde.Mul(cBar))

	if u.H.Equal(hBar) && u.HTilde.Equal(hTildeBar) {
		return true
	}
	return false
}
