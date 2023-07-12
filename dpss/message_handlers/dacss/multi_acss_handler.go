package dacss

import (
	"encoding/binary"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	batch "github.com/arcana-network/dkgnode/dpss/message_handlers/batch_reconstruction"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/keyset"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/sirupsen/logrus"
)

var multiAcssMessageType common.DPSSMessageType = "multi_dacss"

type multiAcssMessage struct {
	RoundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	Ti      int
}

func NewmultiAcssMessage(id common.DPSSRoundID, Ti int, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := multiAcssMessage{
		RoundID: id,
		kind:    multiAcssMessageType,
		curve:   curve,
		Ti:      Ti,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

// Collecting the output from multiple dacss per node base on the batch size
func (m multiAcssMessage) Process(p dpsscommon.DPSSParticipant) {
	state := p.State().DacssStore
	dpssID, _ := common.DPSSIDFromRoundID(m.RoundID)
	// Create empty  state
	defaultDacssState := &dpsscommon.DacssState{
		T_dacss: make(map[common.DPSSID]int, 0),
		Ended:   false,
	}

	DacssState, _ := state.GetOrSet(p.ID(), defaultDacssState)
	DacssState.Lock()
	defer DacssState.Unlock()

	DacssState.T_dacss[dpssID] = m.Ti

	B, _ := batch.GetBatchSize(&dpssID)
	n, _, f := p.Params(false)

	alpha := int(B.Int64()) / (n - 2*f)
	mod := int(B.Int64()) % (n - 2*f)
	if mod > 0 {
		mod = 1
	}
	alpha = alpha + mod

	Tset := make([]int, 0)
	T := 0
	if len(DacssState.T_dacss) == alpha {

		for _, v := range DacssState.T_dacss {
			Tset = append(Tset, v)
		}

		T = Tset[0]
		for i := 1; i < alpha; i += 1 {
			T = T & Tset[i]
		}

		if countSetBits(T) >= n-f && !DacssState.Ended {
			DacssState.Ended = true
			//Move to key set proposal
			var output [8]byte
			binary.BigEndian.PutUint64(output[:], uint64(T))

			//Generate ADKGID
			committeeID := *new(big.Int).SetInt64(int64(0))
			dpssID := common.GenerateDPSSID(*new(big.Int).SetInt64(int64(0)), B)
			ID := common.DPSSID(strings.Join([]string{string(dpssID), committeeID.Text(16)}, common.Delimiter3))

			//Initiate keyset proposal
			round := common.CreateDPSSRound(ID, p.ID(), "keyset")
			msg, err := keyset.NewInitMessage(round, output[:], m.curve)
			if err != nil {
				return
			}
			go p.ReceiveMessage(*msg)
		} else {
			logrus.Info("Not enough nodes agreed upon across multiple dacss per node ")
			return
		}
	}

}
