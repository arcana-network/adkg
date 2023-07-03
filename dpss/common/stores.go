package common

import (
	"sync"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/vivint/infectious"
)

type DPSSStoreMap struct {
	Map sync.Map
}

func (m *DPSSStoreMap) Get(r common.DPSSRoundID) (keygen *common.SharingStore, found bool) {
	inter, found := m.Map.Load(r)
	keygen, _ = inter.(*common.SharingStore)
	return
}

func (store *DPSSStoreMap) GetOrSetIfNotComplete(r common.DPSSRoundID, input *common.SharingStore) (keygen *common.SharingStore, complete bool) {
	inter, found := store.GetOrSet(r, input)
	if found {
		if inter == nil {
			return inter, true
		}
	}
	return inter, false
}

func (store *DPSSStoreMap) GetOrSet(r common.DPSSRoundID, input *common.SharingStore) (keygen *common.SharingStore, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	keygen, _ = inter.(*common.SharingStore)
	return
}

type ABAStoreMap struct {
	Map sync.Map
}

type DecisionStoreMap struct {
	Map sync.Map
}

type CommitmentStoreMap struct {
	Map sync.Map
}

// dpssid=>commitment=>count
func (store *CommitmentStoreMap) GetOrSet(r common.DPSSRoundID, input *CommitmentState) (session *CommitmentState, found bool) {

	inter, found := store.Map.LoadOrStore(r, input)
	session, _ = inter.(*CommitmentState)
	return
}

type OutputCommitmentStoreMap struct {
	Map sync.Map
}

func (store *OutputCommitmentStoreMap) GetOrSet(r common.DPSSRoundID, input *CommitmentsForCommittees) (commitstate *CommitmentsForCommittees, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	commitstate, _ = inter.(*CommitmentsForCommittees)
	return
}

type UShareStoreMap struct {
	Map sync.Map
}

type DacssStoreMap struct {
	Map sync.Map
}

func (store *DacssStoreMap) GetOrSet(r int, input *DacssState) (session *DacssState, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	session, _ = inter.(*DacssState)
	return
}

type TestStoreMap struct {
	Map sync.Map
}

func (store *TestStoreMap) GetOrSet(r common.DPSSRoundID, input *TestState) (session *TestState, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	session, _ = inter.(*TestState)
	return
}

func (store *ABAStoreMap) GetOrSetIfNotComplete(r common.DPSSRoundID, input *common.ABAState) (keygen *common.ABAState, complete bool) {
	inter, found := store.GetOrSet(r, input)
	if found {
		if inter == nil {
			return inter, true
		}
	}
	return inter, false
}

func (store *ABAStoreMap) GetOrSet(r common.DPSSRoundID, input *common.ABAState) (keygen *common.ABAState, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	keygen, _ = inter.(*common.ABAState)
	return
}

type DPSSSessionStore struct {
	Map sync.Map
}

func (m *DPSSSessionStore) Get(r common.DPSSID) (session *DPSSSession, found bool) {
	inter, found := m.Map.Load(r)
	session, _ = inter.(*DPSSSession)
	return
}

func (store *DPSSSessionStore) GetOrSetIfNotComplete(r common.DPSSID, input *DPSSSession) (*DPSSSession, bool) {
	inter, found := store.GetOrSet(r, input)
	if found {
		if inter == nil {
			return inter, true
		}
	}
	return inter, false
}

func (store *DPSSSessionStore) GetOrSet(r common.DPSSID, input *DPSSSession) (session *DPSSSession, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	session, _ = inter.(*DPSSSession)
	return
}

func (store *UShareStoreMap) GetOrSet(r common.DPSSID, input *UshareState) (session *UshareState, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	session, _ = inter.(*UshareState)
	return
}

type UshareState struct {
	sync.Mutex
	ReceivedUshare map[int]bool
	ReceivedU_i    map[int]bool
	Ushares        map[int][]curves.Scalar
	U_i            map[int][]curves.Scalar
	CountUshare    int
	CountU_i       int
	T              map[string]int
	EndedU_i       bool
	EndedUshare    bool
	Ui_Commit      map[int][][]curves.Point
}

type SharingStore struct {
	sync.Mutex
	RoundID    common.DPSSRoundID
	State      common.RBCState
	CStore     map[string]*common.CStore
	ReadyStore []infectious.Share
	Started    bool
}

type SharingStoreMap struct {
	Map sync.Map
}

func (m *SharingStoreMap) Get(r common.DPSSRoundID) (keygen *SharingStore, found bool) {
	inter, found := m.Map.Load(r)
	keygen, _ = inter.(*SharingStore)
	return
}

func (store *SharingStoreMap) GetOrSetIfNotComplete(r common.DPSSRoundID, input *SharingStore) (keygen *SharingStore, complete bool) {
	inter, found := store.GetOrSet(r, input)
	if found {
		if inter == nil {
			return inter, true
		}
	}
	return inter, false
}

func (store *SharingStoreMap) GetOrSet(r common.DPSSRoundID, input *SharingStore) (keygen *SharingStore, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	keygen, _ = inter.(*SharingStore)
	return
}
func (store *SharingStoreMap) Complete(r common.DPSSRoundID) {
	store.Map.Store(r, nil)
}

func GetCStore(keygen *SharingStore, s string) *common.CStore {
	c, found := keygen.CStore[s]
	if !found {
		keygen.CStore[s] = &common.CStore{}
		c = keygen.CStore[s]
	}
	return c
}

type CStore struct {
	EC        int
	RC        int
	ReadySent bool
}
