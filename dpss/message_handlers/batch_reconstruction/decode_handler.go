package batchreconstruction

import (
	"encoding/json"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var DecodeMessageType common.DPSSMessageType = "decode"

type InitDecodeMessage struct {
	RoundID common.DPSSRoundID
	u_share []curves.Scalar
	T       []int
	kind    common.DPSSMessageType
	curve   *curves.Curve
}

func NewInitDecodeMessage(id common.DPSSRoundID, T []int, u_share []curves.Scalar, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := InitDecodeMessage{
		id,
		u_share,
		T,
		DecodeMessageType,
		curve,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m *InitDecodeMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {

	dpssID, _ := common.DPSSIDFromRoundID(m.RoundID)

	// 4) ECC DECODE: u1...uT...uN |-> s1...sT
	// Get state from node
	state := p.State().UshareStore
	// Create empty  state
	defaultUshareState := &dpsscommon.UshareState{
		ReceivedUshare: make(map[int]bool),
		ReceivedU_i:    make(map[int]bool),
		Ushares:        make(map[int][]curves.Scalar),
		U_i:            make(map[int][]curves.Scalar),
		CountUshare:    0,
		CountU_i:       0,
	}

	UshareState, _ := state.GetOrSet(dpssID, defaultUshareState)
	UshareState.Lock()
	defer UshareState.Unlock()

	// Check if the share has already been received
	receivedUshare, found := UshareState.ReceivedUshare[sender.Index]
	if receivedUshare && found {
		log.Debugf("Already received ushare share for %s from %d", dpssID, sender.Index)
		return
	}
	UshareState.ReceivedUshare[sender.Index] = true
	//Verification needs to done
	UshareState.Ushares[sender.Index] = m.u_share
	UshareState.CountUshare++

	var s_i_dash []curves.Scalar
	n2, _, f2 := p.Params(false)
	batchSize, _ := GetBatchSize(&dpssID)
	T := n2 - 2*f2
	b := math.Ceil(float64(batchSize.Int64()) / float64(T))

	si_ri := make([]curves.Scalar, 0)

	if UshareState.CountUshare >= T && !UshareState.EndedUshare {
		UshareState.EndedUshare = true

		for i := 0; i < int(b); i++ {

			ushares := make(map[int]curves.Scalar)
			for k, v := range UshareState.Ushares {
				ushares[k] = v[i]
			}

			threshold := T
			if i == int(b-1) && int(batchSize.Int64())%T != 0 {
				threshold = int(batchSize.Int64()) % T
			}

			coeff, err := dpsscommon.RecoverPriPoly(m.curve, ushares, threshold)
			if err != nil {
				log.Error(err)
			}
			si_ri = append(si_ri, coeff.Coeffs...)
		}

		// generate r'_i
		var r_i_dash []curves.Scalar
		n1, _, f1 := p.Params(false)
		if UshareState.T[IntToString(m.T)] >= f1+1 {
			r_i_dash = m.Getr_i(n1, f1, p)

		}
		//Update key shares
		for i := 0; i < int(batchSize.Int64()); i += 1 {
			s_i_dash = append(s_i_dash, si_ri[i].Sub(r_i_dash[i]))
		}
		//Update commitment
		Timestamp = time.Now()
		//Just for testing

		// TODO: Fix this
		// for i := 0; i < int(batchSize.Int64()); i += 1 {
		// 	adkgID := common.GenerateADKGID(*new(big.Int).SetInt64(int64(i)))
		// 	msg := testhandler.NewTestMessage(adkgID, true, s_i_dash[i], m.curve, p.ID())
		// 	p.Send(msg, p.Nodes(false))
		// }

	}

}

var Timestamp time.Time

func (m *InitDecodeMessage) Getr_i(n, t int, p dpsscommon.DPSSParticipant) []curves.Scalar {

	dpssID, _ := common.DPSSIDFromRoundID(m.RoundID)

	//Create HIM
	him := dpsscommon.CreateHIM(n-t, m.curve)

	batchSize, _ := GetBatchSize(&dpssID)

	//Get shares
	var globalRandoms []curves.Scalar

	committeeID := *new(big.Int).SetInt64(1)

	sharesPerNode := int(batchSize.Int64()) / (n - 2*t)
	mod := int(batchSize.Int64()) % (n - 2*t)
	if mod > 0 {
		mod = 1
	}
	sharesPerNode = sharesPerNode + mod

	for b := 0; b < sharesPerNode; b++ {
		var localRandoms []curves.Scalar
		rIndex := *new(big.Int).SetInt64(int64(b))
		index := strings.Join([]string{rIndex.Text(16), batchSize.Text(16)}, common.Delimiter2)
		adkgID := common.DPSSID(strings.Join([]string{"DPSS", index, committeeID.Text(16)}, common.Delimiter3))
		shareStore, found := p.State().SessionStore.GetOrSet(adkgID, dpsscommon.DefaultDPSSSession())
		if !found {
			log.Errorf("Store empty")
		}

		for _, i := range m.T {
			shareScalar, err := m.curve.Scalar.SetBytes(shareStore.S[i].Value)
			if err != nil {
				continue
			}
			localRandoms = append(localRandoms, shareScalar)
		}
		for i := 0; i < (n - 2*t); i++ {

			globalRandoms = append(globalRandoms, dpsscommon.DotProduct(him[i][:n-t], localRandoms, m.curve))
		}

	}

	//r_i values
	globalRandoms = globalRandoms[:batchSize.Int64()]
	return globalRandoms

}
