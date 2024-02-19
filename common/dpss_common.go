package common

import (
	"errors"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

// PSSParticipant is the interface that covers all the participants inside the
// DPSS protocol
type PSSParticipant interface {
	// Returns the ID of the participant
	ID() int
	// Defines if the current node belongs to the old or new committee.
	IsOldNode() bool
	// Obtains the public key from a node in the old or new committee. The
	// committee is defined by the flag fromNewCOmmittee.
	PublicKey(idx int, fromNewCommittee bool) curves.Point
	// Obtains the parameters of the protocols for the committee for which the
	// current node belongs.
	Params() (n int, k int, t int)
	// Broadcast a message to the old or new committee. The committee is defined
	// by the flag toNewCommittee.
	Broadcast(toNewCommittee bool, msg DKGMessage)
	// Send a message to a given node.
	Send(n NodeDetails, msg DKGMessage) error
	// Obtains the details of the current node.
	Details() NodeDetails
	// Returns the private key of the current node.
	PrivateKey() curves.Scalar
	// Receives a message from a given sender.
	ReceiveMessage(sender NodeDetails, msg DKGMessage)
	// Obtains the nodes from the new or old committee. The committee is defined
	// by the flag fromNewCommitte.
	Nodes(fromNewCommittee bool) map[NodeDetailsID]NodeDetails
}

// PSSNodeState represents the internal state of a node that participates in
// possibly multiple DPSS protocol. There is an storage for the different
// sub-protocols in the DPSS: ACSS, RBC
type PSSNodeState struct {
	ShareStore *PSSShareStoreMap // Storage for the shares in the ACSS Protocol.
	RbcStore   *RBCStateMap      // Storage for the RBC protocol.
}

// Stores the shares tha the node receives during the DPSS protocol.
type PSSShareStore struct {
	sync.Mutex
	Shares map[int]curves.Scalar // Map of shares. K: index of the owner of the share, V: the actual share.
}

// PSSID defines the ID of an instance of the DPSS protocol.
type PSSID string

// PSSRoundDetails represents all the details in a round for the DPSS protocol.
type PSSRoundDetails struct {
	PSSID  PSSID  // ID for the PSS.
	Dealer int    // ID of the node that is dealing the information to other parties.
	Kind   string // Stage of the DPSS protocol in which the round is.
}

// Stores the information of the shares for a given round ID. Remember that
// RBC can be executed in multiple rounds at the same time. This storage saves
// the RBC information for all of the rounds.
type PSSShareStoreMap struct {
	Map sync.Map // Key: RoundID, Value: PSSSharingStore
}

// Obtains a sharing store for a PSS round given the round ID. Returns the
// corresponding share store, and a boolean telling if the key was found or not.
func (m *PSSShareStoreMap) Get(r RoundID) (shares *PSSShareStore, found bool) {
	inter, found := m.Map.Load(r)
	shares, _ = inter.(*PSSShareStore)
	return
}

func (store *PSSShareStoreMap) GetOrSetIfNotComplete(r RoundID, input *PSSShareStore) (keygen *PSSShareStore, complete bool) {
	inter, found := store.GetOrSet(r, input)
	if found {
		if inter == nil {
			return inter, true
		}
	}
	return inter, false
}

// Obtains a share store given a round ID. If such key is not in the map, it
// sotres the given share store using the provided key. If the key was not in
// the map, then the function returns a boolean flag.
func (store *PSSShareStoreMap) GetOrSet(r RoundID, input *PSSShareStore) (keygen *PSSShareStore, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	keygen, _ = inter.(*PSSShareStore)
	return
}

func (store *PSSShareStoreMap) Complete(r RoundID) {
	store.Map.Store(r, nil)
}

// Deletes a share store given the ID of its round.
func (store *PSSShareStoreMap) Delete(r RoundID) {
	store.Map.Delete(r)
}

// Obtains an round ID from the round details by appending the information
// together.
func (d *PSSRoundDetails) ID() RoundID {
	return RoundID(strings.Join([]string{string(d.PSSID), d.Kind, strconv.Itoa(d.Dealer)}, Delimiter4))
}

// Generates a new PSSID for a given index.
func NewPSSID(index big.Int) PSSID {
	return PSSID(strings.Join([]string{"PSS", index.Text(16)}, Delimiter3))
}

// Return the index of a PSSID.
func (id *PSSID) GetIndex() (big.Int, error) {
	str := string(*id)
	substrs := strings.Split(str, Delimiter3)

	if len(substrs) != 2 {
		return *new(big.Int), errors.New("could not parse DPSSID")
	}

	index, ok := new(big.Int).SetString(substrs[1], 16)
	if !ok {
		return *new(big.Int), errors.New("could not get back index from DPSSID")
	}

	return *index, nil
}
