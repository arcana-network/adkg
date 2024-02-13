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
	Send(n KeygenNodeDetails, msg DKGMessage) error
	Details() KeygenNodeDetails
	PrivateKey() curves.Scalar
	ReceiveMessage(sender KeygenNodeDetails, msg DKGMessage)
	Nodes(fromNewCommittee bool) map[NodeDetailsID]KeygenNodeDetails
}

type PSSNodeState struct {
	shares   PSSShareStoreMap
	rbcStore RBCStateMap
}

type PSSShareStore struct {
	sync.Mutex
	roundID RoundID
}

// Stores the information of the shares for a given round ID.
type PSSShareStoreMap struct {
	store sync.Map // Key: roundID, Value: SharingStore
}
