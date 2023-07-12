package batchreconstruction

import (
	"encoding/json"
	"math"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

var InitMessageType common.DPSSMessageType = "br_init"

type InitBatchMessage struct {
	RoundID common.DPSSRoundID
	s       []curves.Scalar
	T       []int
	kind    common.DPSSMessageType
	curve   *curves.Curve
}

func NewInitBatchMessage(id common.DPSSRoundID, s []curves.Scalar, T []int, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := InitBatchMessage{
		id,
		s,
		T,
		InitMessageType,
		curve,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m *InitBatchMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	// 1) EXPAND: [s1]...[sT] |-> [u1]...[uT]...[uN]    (polnomial evaluation)
	u_shares := make([][]curves.Scalar, len(p.Nodes(true)))
	vandMatrix := dpsscommon.CreateHIM(len(p.Nodes(true)), m.curve)

	n, _, t := p.Params(false)
	T := n - 2*t
	b := math.Ceil(float64(len(m.s)) / float64(T))

	for j := 0; j < int(b); j++ {
		end := (j * T) + T

		for i := 0; i < len(p.Nodes(true)); i++ {
			if j == int(b)-1 {
				end = len(m.s[j*T:])
				u_shares[i] = append(u_shares[i], dpsscommon.DotProduct(vandMatrix[i][:end], m.s[j*T:], m.curve))

			} else {
				u_shares[i] = append(u_shares[i], dpsscommon.DotProduct(vandMatrix[i][:T], m.s[j*T:end], m.curve))
			}
		}
	}

	// 2) PRIVATE RECONSTRUCT: [u1]...[uT]...[uN] -> u1...uT...uN
	for _, n := range p.Nodes(true) {
		go func(node common.KeygenNodeDetails) {
			msg, err := NewInitReconsMessage(m.RoundID, m.T, u_shares[node.Index], m.curve)
			if err != nil {
				return
			}
			p.Send(*msg, node)
		}(n)
	}

}
