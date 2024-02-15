package keyderivation

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/curves"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/common/aba"

	log "github.com/sirupsen/logrus"
)

var ShareMessageType string = "key_derivation_pk_share"

type ShareMessage struct {
	RoundID common.RoundID
	Kind    string
	Curve   common.CurveName
	Share   []byte
	R       []byte
	S       []byte
}

func NewShareMessage(id common.RoundID, curve common.CurveName, share, r, s []byte) (*common.DKGMessage, error) {
	m := ShareMessage{
		id,
		ShareMessageType,
		curve,
		share,
		r, s,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m ShareMessage) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {

	n, k, _ := self.Params()
	curve := common.CurveFromName(m.Curve)

	adkgid, err := common.ADKGIDFromRoundID(m.RoundID)
	if err != nil {
		return
	}

	sessionStore, complete := self.State().SessionStore.GetOrSetIfNotComplete(adkgid, common.DefaultADKGSession())
	if complete {
		log.Debugf("Keygen already complete: %s", adkgid)
		return
	}
	sessionStore.Lock()
	defer sessionStore.Unlock()

	if !(n == len(sessionStore.Decisions) && sessionStore.ABAComplete) {
		sessionStore.PubKeySharesUnverified[sender.Index] = common.PubKeyShare{
			R:     m.R,
			S:     m.S,
			Share: m.Share,
		}
		return
	}

	keysets := make([][]int, 0)
	for k, v := range sessionStore.Decisions {
		if v == 1 {
			keysets = append(keysets, kcommon.GetSetBits(n, sessionStore.T[k]))
		}
	}

	T := Union(keysets...)

	log.WithFields(log.Fields{
		"keysets":       keysets,
		"sessionStore":  sessionStore.T,
		"decisions":     sessionStore.Decisions,
		"completeCount": len(sessionStore.Decisions),
		"self":          self.ID(),
		"adkgid":        adkgid,
	}).Debugf("len(T)==%d", len(T))

	if len(T) == 0 {
		return
	}

	gZj := aba.DerivePublicKey(sender.Index, k, curve, T, sessionStore.C) //y1

	hZj, verified := VerifyShare(common.PubKeyShare{
		S:     m.S,
		R:     m.R,
		Share: m.Share,
	}, curve, gZj, self)

	if !verified {
		return
	}
	sessionStore.PubKeyShares[sender.Index] = hZj

	ProcessUnverifiedShares(sessionStore, curve, k, T, self)

	if len(sessionStore.PubKeyShares) >= k && !sessionStore.Over { // t+1
		identities := make([]int, 0)

		for i := range sessionStore.PubKeyShares {
			identities = append(identities, i)
		}

		coeff, err := aba.LagrangeCoeffs(identities, curve)
		if err != nil {
			log.Errorf("Error aba.LagrangeCoeffs: %s", err)
			return
		}

		hZ := curve.Point.Identity()
		log.Debugf("length=%d, val=%v", len(coeff), coeff)
		log.Debugf("PubKeyShares=%d", sessionStore.PubKeyShares)
		for i := range coeff {
			hZ = hZ.Add(sessionStore.PubKeyShares[i].Mul(coeff[i]))
		}

		log.Infof("Finished keysharing for id: %s", adkgid)

		log.WithFields(log.Fields{
			"hz":     hZ.ToAffineUncompressed(),
			"shares": sessionStore.PubKeyShares,
			"adkgid": adkgid,
		}).Debug("key_derivation_share")

		adkgid, err := common.ADKGIDFromRoundID(m.RoundID)
		if err != nil {
			return
		}

		keyIndex, err := adkgid.GetIndex()
		if err != nil {
			return
		}

		zI := curve.Scalar.Zero()

		for _, j := range T {
			// 1,2,3
			shareScalar, err := curve.Scalar.SetBytes(sessionStore.S[j].Value)
			if err != nil {
				log.Errorf("Share set byte failed: err=%s", err)
				continue
			}
			zI = zI.Add(shareScalar) //x
		}
		if sessionStore.BFTDecided {
			c, err := adkgid.GetCurve()
			if err != nil {
				return
			}
			self.StoreCompletedShare(keyIndex, *zI.BigInt(), c)
			self.StoreCommitment(keyIndex, common.ADKGMetadata{Commitments: sessionStore.C, T: T}, c)
			self.Cleanup(adkgid)
		} else {
			sessionStore.Share = zI.BigInt()
			sessionStore.Commitments = common.ADKGMetadata{Commitments: sessionStore.C, T: T}
			sessionStore.Over = true
		}

		msg, err := NewPubKeygenMessage(m.RoundID, m.Curve, hZ)
		if err != nil {
			return
		}

		go self.ReceiveBFTMessage(*msg)
	}
}

func VerifyShare(s common.PubKeyShare,
	curve *curves.Curve, gZj curves.Point, self common.DkgParticipant) (curves.Point, bool) {

	length := 33
	if curve.Name == "ed25519" {
		length = 32
	}
	sBar, err := curve.Scalar.SetBytes(s.S)
	if err != nil {
		return nil, false
	}
	aBar, err := curve.Point.FromAffineCompressed(s.R[:length]) //g^k
	if err != nil {
		return nil, false
	}
	bBar, err := curve.Point.FromAffineCompressed(s.R[length:]) //h^k
	if err != nil {
		return nil, false
	}
	hZj, err := curve.Point.FromAffineCompressed(s.Share) //y2
	if err != nil {
		return nil, false
	}

	g, h := self.CurveParams(curve.Name)

	//c=Hash(g,g_zi,h,h_zi,A,B)
	cBar := aba.Hash(g, gZj, h, hZj, aBar, bBar, curve)

	//A′=s*g + c*(g_zj)
	aDash := g.Mul(sBar).Add(gZj.Mul(cBar))

	//B′=s*h +c*h_zj
	bDash := h.Mul(sBar).Add(hZj.Mul(cBar))

	if aBar.Equal(aDash) && bBar.Equal(bDash) {
		return hZj, true
	}
	return nil, false
}

func ProcessUnverifiedShares(sessionStore *common.ADKGSession, curve *curves.Curve,
	k int, T []int, self common.DkgParticipant) {
	for nodeIndex, share := range sessionStore.PubKeySharesUnverified {
		gZj := aba.DerivePublicKey(nodeIndex, k, curve, T, sessionStore.C) //y1

		hZj, verified := VerifyShare(common.PubKeyShare{
			S:     share.S,
			R:     share.R,
			Share: share.Share,
		}, curve, gZj, self)

		if !verified {
			continue
		}
		delete(sessionStore.PubKeySharesUnverified, nodeIndex)
		sessionStore.PubKeyShares[nodeIndex] = hZj
	}
}
