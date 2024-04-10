package aba

import (
	"crypto/sha256"
	"strconv"
	"time"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"

	"github.com/arcana-network/dkgnode/common"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/common/aba"
)

var CoinMessageType string = "aba_coin"

type CoinMessage struct {
	RoundID common.PSSRoundDetails
	Kind    string
	Curve   common.CurveName
	Data    []byte
}

func NewCoinMessage(id common.PSSRoundDetails, data []byte, curve common.CurveName) (*common.PSSMessage, error) {
	m := CoinMessage{
		id,
		CoinMessageType,
		curve,
		data,
	}
	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m *CoinMessage) Process(sender common.NodeDetails, self common.PSSParticipant) {
	curve := common.CurveFromName(m.Curve)
	u, err := unpack(curve, m.Data)
	if err != nil {
		log.WithError(err).Error("Could not unpack data in aba_coin_share")
		return
	}
	n, k, f := self.Params()

	roundLeader := m.RoundID.Dealer.Index

	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.RoundID.ToRoundID(), common.DefaultABAStore())
	if complete {
		log.Infof("Keygen already complete: %v", m.RoundID)
		return
	}
	store.Lock()
	coinID := string(m.RoundID.ToRoundID()) + strconv.Itoa(store.GetRound())
	store.Unlock()

	gTilde := curve.Point.Hash([]byte(coinID))

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
		// Breakout if time since message received has exceeded 10s
		if time.Since(start) > time.Second*20 {
			pssState.Unlock()
			log.Errorf("timeout coin_share message, round=%v", m.RoundID)
			return
		}

		pssState.Unlock()

		time.Sleep(200 * time.Millisecond)
	}

	pssState.Lock()
	defer pssState.Unlock()

	TiSet = kcommon.GetSetBits(n, pssState.T[roundLeader])

	if len(TiSet) == 0 {
		log.Infof("TiSet == 0 for round: %v, self: %d", m.RoundID, self.Details().Index)
		return
	}

	// TODO: Recheck this, using the first sample for coin tossing
	gI := aba.DerivePublicKey(sender.Index, k, curve, TiSet, pssState.KeysetMap[0].CommitmentStore)

	log.WithFields(log.Fields{
		"self":      self.Details().Index,
		"sender":    sender.Index,
		"round":     m.RoundID,
		"publicKey": gI.ToAffineCompressed(),
		"T":         pssState.T,
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
		"self":             self.Details().Index,
		"sender":           sender.Index,
		"round":            m.RoundID,
		"coinsharesLength": len(coinShares),
		"k":                k,
		"decisions":        pssState.Decisions,
	}).Debug("aba_coin")

	_, ok := pssState.Decisions[roundLeader]

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
			"self":   self.Details().Index,
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

		log.WithFields(log.Fields{
			"self":                 self.Details().Index,
			"sender":               sender.Index,
			"round":                m.RoundID,
			"sessionStoreDecision": pssState.Decisions,
			"Inequality(1==true)":  int(sha256.Sum256(g0Tilde.ToAffineCompressed())[31]) % 2,
		}).Info("aba_coin")

		if int(sha256.Sum256(g0Tilde.ToAffineCompressed())[31])%2 == 1 {
			pssState.Decisions[roundLeader] = 1
			if !pssState.ABAComplete {
				pssState.ABAComplete = true
				pssID := m.RoundID.PssID
				log.Debugf("PSSID=%s, decisions=%v,self=%d", pssID, pssState.Decisions, self.Details().Index)

				for i := 1; i <= n; i++ {
					if !Contains(pssState.ABAStarted, i) {
						go func(id int) {
							round := common.CreatePSSRound(pssID, m.RoundID.Dealer, "keyset")
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
			pssState.Decisions[roundLeader] = 0
		}

		log.WithFields(log.Fields{
			"self":                 self.Details().Index,
			"sender":               sender.Index,
			"round":                m.RoundID,
			"Decision":             pssState.Decisions[roundLeader],
			"sessionStoreDecision": pssState.Decisions,
			"CompleteCount":        len(pssState.Decisions),
			"ABAComplete":          pssState.ABAComplete,
		}).Info("aba_coin")

		// If all rounds ABA'd to 0 or 1, set ABA complete to true and start key derivation
		if n == len(pssState.Decisions) && !pssState.HIMStarted {
			// 1) Get list of Keysets voted as 1
			// 2) Get T[index] from each keyset and union to get T
			// T := sessionStore.GetTSet(n, f)

			// shares := sessionStore.GetSharesFromT(T)
			// 3) Get shares and compress
			// Len(share) = B/n-2t * n-t
			// Somehow sort and create array from shares
			// [(1,1), (1,2), (1,3), (2, 1) ....]
			// (nodeIndex, acssCount) or (acssCount, nodeIndex) ?

			// msg, err := dpss.NewDacssHimMessage(m.RoundID, shares, m.Curve)
			// if err != nil {
			// 	return
			// }
			// sessionStore.HIMStarted = true
			// go self.ReceiveMessage(self.Details(), *msg)
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

func verify(u *Unpack, gTilde, gI curves.Point, curve *curves.Curve, self common.PSSParticipant) bool {
	g, _ := self.CurveParams(curve.Name)

	cBar := aba.Hash(g, u.H, gTilde, u.HTilde, gI, u.GiTilde, curve)

	hBar := g.Mul(u.Z).Sub(gI.Mul(cBar))

	hTildeBar := gTilde.Mul(u.Z).Sub(u.GiTilde.Mul(cBar))

	if u.H.Equal(hBar) && u.HTilde.Equal(hTildeBar) {
		return true
	}
	return false
}
