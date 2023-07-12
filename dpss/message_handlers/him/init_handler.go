package him

import (
	"encoding/json"
	"math/big"
	"sort"
	"strings"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	batchreconstruction "github.com/arcana-network/dkgnode/dpss/message_handlers/batch_reconstruction"
	keygencommon "github.com/arcana-network/dkgnode/keygen/common"

	key "github.com/arcana-network/dkgnode/keygen/message_handlers/keyderivation"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var InitMessageType common.DPSSMessageType = "him_init"

type InitMessage struct {
	RoundID      common.DPSSRoundID
	newCommittee bool
	kind         common.DPSSMessageType
	curve        *curves.Curve
}

func NewInitMessage(id common.DPSSRoundID, newCommittee bool, curve *curves.Curve) (*common.DPSSMessage, error) {
	m := InitMessage{
		id,
		newCommittee,
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

func (m *InitMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	if sender.Index != p.ID() {
		return
	}
	curve := m.curve
	n, _, t := p.Params(m.newCommittee)

	// create a large him to extract submatrices later from cache
	dpsscommon.CreateHIM(len(p.Nodes(true)), curve)
	//Create HIM
	him := dpsscommon.CreateHIM(n-t, curve)

	dpssID, _ := common.DPSSIDFromRoundID(m.RoundID)

	//Get T
	store, _ := p.State().SessionStore.GetOrSet(dpssID, dpsscommon.DefaultDPSSSession())

	keysets := make([][]int, 0)
	for k, v := range store.Decisions {
		if v == 1 {
			keysets = append(keysets, keygencommon.GetSetBits(n, store.T[k]))
		}
	}
	T := key.Union(keysets...)
	sort.Ints(T)
	T = T[:(n - t)]

	batchSize, _ := batchreconstruction.GetBatchSize(&dpssID)

	//Get shares
	var globalRandoms []curves.Scalar

	committeeID := *new(big.Int).SetInt64(int64(0))

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
		dpssID := common.DPSSID(strings.Join([]string{"DPSS", index, committeeID.Text(16)}, common.Delimiter3))
		shareStore, found := p.State().SessionStore.GetOrSet(dpssID, dpsscommon.DefaultDPSSSession())
		if !found {
			log.Errorf("Store empty")
		}

		for _, i := range T {
			shareScalar, err := m.curve.Scalar.SetBytes(shareStore.S[i].Value)
			if err != nil {
				continue
			}
			localRandoms = append(localRandoms, shareScalar)
		}
		for i := 0; i < (n - 2*t); i++ {

			globalRandoms = append(globalRandoms, dpsscommon.DotProduct(him[i][:n-t], localRandoms, curve))
		}

	}

	//r_i values
	globalRandoms = globalRandoms[:batchSize.Int64()]

	si_ri := make([]curves.Scalar, int(batchSize.Int64()))

	for b := 0; b < int(batchSize.Int64()); b++ {

		//Get s_i values ,Accessing key share
		dpssID := common.GenerateDPSSID(*new(big.Int).SetInt64(int64(b)))
		privKeyShare, _ := p.State().SessionStore.GetOrSet(dpssID, dpsscommon.DefaultDPSSSession())
		si_ri[b] = globalRandoms[b].Add(privKeyShare.Z)

	}

	msg, err := batchreconstruction.NewInitBatchMessage(m.RoundID, si_ri, T, m.curve)
	if err != nil {
		return
	}
	p.ReceiveMessage(*msg)

}
