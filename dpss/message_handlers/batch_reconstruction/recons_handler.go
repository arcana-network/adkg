package batchreconstruction

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/dkgnode/keygen/common/aba"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var ReconsMessageType common.DPSSMessageType = "recons"

type InitReconsMessage struct {
	RoundID common.DPSSRoundID
	u_i     []curves.Scalar
	T       []int
	kind    common.DPSSMessageType
	curve   *curves.Curve
}

func NewInitReconsMessage(id common.DPSSRoundID, T []int, u_i []curves.Scalar, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := InitReconsMessage{
		id,
		u_i,
		T,
		ReconsMessageType,
		curve,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m *InitReconsMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {

	dpssID, _ := common.DPSSIDFromRoundID(m.RoundID)
	batchSize, _ := GetBatchSize(&dpssID)
	n, _, f := p.Params(false)
	T := n - 2*f
	b := math.Ceil(float64(batchSize.Int64()) / float64(T))

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
		T:              make(map[string]int),
		EndedU_i:       false,
		EndedUshare:    false,
		//	Ui_Commit:      common.GenerateUiCommits(m.curve, p, m.T, batchSize),
	}

	UshareState, _ := state.GetOrSet(dpssID, defaultUshareState)
	UshareState.Lock()
	defer UshareState.Unlock()

	// Check if the share has already been received
	receivedU_i, found := UshareState.ReceivedU_i[sender.Index]
	if receivedU_i && found {
		log.Debugf("Already received u_i share for %s from %d", dpssID, sender.Index)
		return
	}
	//Verification needs to be done
	/*for j := 0; j < int(b); j++ {
		if UshareState.Ui_Commit[m.sender][p.ID()][j] != m.curve.ScalarBaseMult(m.u_i[j]) {
			fmt.Println("Here")
		} else {
			fmt.Println("Not here")
		}
	}*/
	UshareState.ReceivedU_i[sender.Index] = true
	UshareState.U_i[sender.Index] = m.u_i
	UshareState.CountU_i++
	UshareState.T[IntToString(m.T)]++

	if UshareState.CountU_i >= f+1 && !UshareState.EndedU_i {
		UshareState.EndedU_i = true
		ushare := Reconstruct(UshareState.U_i, m.curve, int(b), f+1)

		// 3) BROADCAST: u1...uT...uN   (rbc)
		for _, n := range p.Nodes(true) {
			go func(node common.KeygenNodeDetails) {
				msg, err := NewInitDecodeMessage(m.RoundID, m.T, ushare, m.curve)
				if err != nil {
					return
				}
				p.Send(*msg, node)
			}(n)
		}

	}

}

func Reconstruct(ushare map[int][]curves.Scalar, curve *curves.Curve, b, k int) []curves.Scalar {

	identities := make([]int, 0)
	u_i := make([]curves.Scalar, b)

	for i := range ushare {
		identities = append(identities, i)
	}

	coeff, _ := aba.LagrangeCoeffs(identities[0:k], curve)

	for j := 0; j < b; j++ {
		u_i[j] = curve.Scalar.Zero()

		for i := range coeff {
			u_i[j] = u_i[j].Add(ushare[i][j].Mul(coeff[i]))
		}
	}
	return u_i
}

func IntToString(a []int) string {
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.Itoa(v)
	}

	return strings.Join(b, ",")
}
