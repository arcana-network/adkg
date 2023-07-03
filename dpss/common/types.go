package common

import (
	"sync"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/sharing"
)

type DPSSParticipant interface {
	ID() int
	ReceiveMessage(common.DPSSMessage)
	State() *NodeState
	Params(bool) (int, int, int)
	Transport
	CurveParams(*curves.Curve) (curves.Point, curves.Point)
	PublicKey(int) curves.Point
	SelfPrivateKey() curves.Scalar
	Contract() SmartContract
}

type Transport interface {
	Broadcast(newNodes bool, msg common.DPSSMessage)
	Send(msg common.DPSSMessage, node common.KeygenNodeDetails)
	Nodes(newNodes bool) map[common.NodeDetailsID]common.KeygenNodeDetails

	// NewCommittee(committee *[]DPSSParticipant)
}

type SmartContract interface {
	Register(int, curves.Point)
	PublicKey(int) curves.Point
	KeyDetails() common.KeygenDetails
	PublicParams(name string) (curves.Point, curves.Point)
}

type NodeState struct {
	KeygenStore           *SharingStoreMap
	SessionStore          *DPSSSessionStore
	ABAStore              *ABAStoreMap
	DecisionStore         *DecisionStoreMap
	CommitmentStore       *CommitmentStoreMap
	OutputCommitmentStore *OutputCommitmentStoreMap
	UshareStore           *UShareStoreMap
	DacssStore            *DacssStoreMap
	TestStore             *TestStoreMap
}

type TestState struct {
	sync.Mutex
	Old      map[int]curves.Scalar
	New      map[int]curves.Scalar
	EndedOld bool
	EndedNew bool
}

type DPSSSession struct {
	sync.Mutex
	// All keysets
	T          map[int]int
	TProposals map[int]int
	TPrime     int
	IsTSet     bool
	// Share mapping of acss dealer -> share
	S                      map[int]sharing.ShamirShare
	C                      map[int][]curves.Point
	PubKeyShares           map[int]curves.Point
	PubKeySharesUnverified map[int]common.PubKeyShare
	Decisions              map[int]int
	ABAComplete            bool
	ABAStarted             []int
	KeyderivationStarted   bool
	Z                      curves.Scalar
}

func DefaultDPSSSession() *DPSSSession {
	s := DPSSSession{
		C:                      make(map[int][]curves.Point),
		S:                      make(map[int]sharing.ShamirShare),
		PubKeyShares:           make(map[int]curves.Point),
		PubKeySharesUnverified: make(map[int]common.PubKeyShare),
		Decisions:              make(map[int]int),
		T:                      make(map[int]int),
		TProposals:             make(map[int]int),
		TPrime:                 0,
		ABAStarted:             []int{},
	}

	return &s
}

type CommitmentState struct {
	CommitmentsForADKGID map[string]int
	ReceivedCommit       map[string][]int
	Ended                map[string]bool
	sync.Mutex
}
