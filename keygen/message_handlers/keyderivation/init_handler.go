package keyderivation

import (
	"crypto/rand"
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/common/aba"
	log "github.com/sirupsen/logrus"
)

var InitMessageType string = "key_derivation_init"

type InitMessage struct {
	RoundID common.RoundID
	Kind    string
	Curve   common.CurveName
}

func NewInitMessage(id common.RoundID, curve common.CurveName) (*common.DKGMessage, error) {
	m := InitMessage{
		id,
		InitMessageType,
		curve,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m InitMessage) Process(sender common.NodeDetails, self common.DkgParticipant) {
	if sender.Index != self.ID() {
		return
	}
	curve := common.CurveFromName(m.Curve)

	adkgid, err := common.ADKGIDFromRoundID(m.RoundID)
	if err != nil {
		return
	}
	n, _, _ := self.Params()

	sessionStore, complete := self.State().SessionStore.GetOrSetIfNotComplete(adkgid, common.DefaultADKGSession())
	if complete {
		log.Infof("Keygen already complete: %s", adkgid)
		return
	}

	sessionStore.Lock()
	defer sessionStore.Unlock()

	keysets := make([][]int, 0)
	for nodeIndex, v := range sessionStore.Decisions {
		if v == 1 {
			keysets = append(keysets, kcommon.GetSetBits(n, sessionStore.T[nodeIndex]))
		}
	}

	T := Union(keysets...)
	if len(T) == 0 {
		return
	}

	log.WithFields(log.Fields{
		"len(T)":  len(T),
		"T":       T,
		"keysets": keysets,
	}).Debug("keyderivation_init")

	zI := curve.Scalar.Zero()
	for _, j := range T {
		shareScalar, err := curve.Scalar.SetBytes(sessionStore.S[j].Value)
		if err != nil {
			continue
		}
		zI = zI.Add(shareScalar) //x
	}

	g, h := self.CurveParams(curve.Name)

	gZi := g.Mul(zI) // y1
	hZi := h.Mul(zI) // y2

	kRand := curve.Scalar.Random(rand.Reader)
	A := g.Mul(kRand) //g^k
	B := h.Mul(kRand) //h^k

	//c=Hash(g,g_zi,h,h_zi,A,B)
	C := aba.Hash(g, gZi, h, hZi, A, B, curve)

	// S=k_rand − C*Z_i
	S := C.Mul(zI).Sub(kRand)
	S = S.Neg()

	//Send (S,A,B)
	r := make([]byte, 0)
	r = append(r, A.ToAffineCompressed()...) //33 bytes
	r = append(r, B.ToAffineCompressed()...) //33 bytes

	msg, err := NewShareMessage(m.RoundID, m.Curve, hZi.ToAffineCompressed(), r, S.Bytes())
	if err != nil {
		return
	}
	go self.Broadcast(*msg)
}

func Union(args ...[]int) []int {
	if len(args) == 0 {
		return []int{}
	}

	a := args[0]
	m := make(map[int]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, s := range args[1:] {
		for _, item := range s {
			if _, ok := m[item]; !ok {
				a = append(a, item)
				m[item] = true
			}
		}
	}
	return a
}
