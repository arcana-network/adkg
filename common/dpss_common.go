package common

import (
	"sync"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

type PSSParticipant interface {
	IsOldNode() bool
	IsNewNode() bool
	PublicKey(idx int, fromNewCommittee bool) curves.Point
	Params(fromNewCommittee bool) (n int, k int, t int)
	Broadcast(toNewCommittee bool, msg DKGMessage)
	Send(n NodeDetails, msg DKGMessage) error
	Details(fromNewCommittee bool) NodeDetails
	PrivateKey() curves.Scalar
	ReceiveMessage(sender NodeDetails, msg DKGMessage)
	Nodes(fromNewCommittee bool) map[NodeDetailsID]NodeDetails
}

type PSSNodeState struct {
	Shares   *PSSShareStoreMap
	RbcStore *RBCStateMap
}

type PSSShareStore struct {
	sync.Mutex
	roundID RoundID
}

// Stores the information of the shares for a given round ID.
type PSSShareStoreMap struct {
	store sync.Map // Key: roundID, Value: SharingStore
}
