package dacss

import (
	"encoding/binary"
	"math/big"
	"strings"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/sirupsen/logrus"

	"github.com/arcana-network/adkg-proto/common"
	batch "github.com/arcana-network/adkg-proto/dpssmessages/batchreconstruction"
	"github.com/arcana-network/adkg-proto/dpssmessages/keyset"
)

var multiAcssMessageType common.MessageType = "multi_dacss"

type multiAcssMessage struct {
	ID     common.ADKGID
	sender int
	kind   common.MessageType
	curve  *curves.Curve
	Ti     int
}

func NewmultiAcssMessage(id common.ADKGID, Ti int, curve *curves.Curve, sender int) common.DKGMessage {
	m := multiAcssMessage{
		ID:     id,
		sender: sender,
		kind:   multiAcssMessageType,
		curve:  curve,
		Ti:     Ti,
	}

	return m
}

func (m multiAcssMessage) Sender() int {
	return m.sender
}

func (m multiAcssMessage) Kind() common.MessageType {
	return m.kind
}

// Collecting the output from multiple dacss per node base on the batch size
func (m multiAcssMessage) Process(p common.DkgParticipant) {
	// Get state from node
	state := p.State().DacssStore
	// Create empty  state
	defaultDacssState := &common.DacssState{
		T_dacss: make(map[common.ADKGID]int, 0),
		Ended:   false,
	}

	DacssState, _ := state.GetOrSet(p.ID(), defaultDacssState)
	DacssState.Lock()
	defer DacssState.Unlock()

	DacssState.T_dacss[m.ID] = m.Ti

	B, _ := batch.GetBatchSize(&m.ID)
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
			ID := common.ADKGID(strings.Join([]string{string(dpssID), committeeID.Text(16)}, common.Delimiter3))

			//Initiate keyset proposal
			round := common.CreateRound(ID, p.ID(), "keyset")
			msg := keyset.NewKeysetInitMessage(round, output[:], m.curve, p.ID())
			go p.ReceiveMessage(msg)
		} else {
			logrus.Info("Not enough nodes agreed upon across multiple dacss per node ")
			return
		}
	}

}
