package common

import (
	"errors"
	"math/big"
	"strings"
	"sync"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

// PSSParticipant is the interface that covers all the participants inside the
// DPSS protocol
type PSSParticipant interface {
	// For PSS state
	State() *PSSNodeState
	// Defines if the current node belongs to the old or new committee.
	IsOldNode() bool
	// Obtains the public key from a node in the old or new committee. The
	// committee is defined by the flag fromNewCOmmittee.
	GetPublicKeyFor(idx int, fromNewCommittee bool) curves.Point
	// Obtains the parameters of the protocols for the committee for which the
	// current node belongs.
	Params() (n int, k int, t int)
	// Broadcast a message to the old or new committee. The committee is defined
	// by the flag toNewCommittee.
	Broadcast(toNewCommittee bool, msg PSSMessage)
	// Send a message to a given node.
	Send(n NodeDetails, msg PSSMessage) error
	// Obtains the details of the current node.
	// The public key in the NodeDetails uniquely identifies the node.
	Details() NodeDetails
	// Returns the private key of the current node.
	PrivateKey() curves.Scalar
	// Receives a message from a given sender.
	ReceiveMessage(sender NodeDetails, msg PSSMessage)
	// Obtains the nodes from the new or old committee. The committee is defined
	// by the flag fromNewCommitte.
	Nodes(fromNewCommittee bool) map[NodeDetailsID]NodeDetails
}

// PSSNodeState represents the internal state of a node that participates in
// possibly multiple DPSS protocol. There is an storage for the different
// sub-protocols in the DPSS: ACSS, RBC
type PSSNodeState struct {
	DacssStore      *DacssStateMap // State for the separate ACSS rounds
	DualAcssStarted bool           // Flag whether the DualAcss part of the protocol has started.
}

// Stores the information of the shares for a given ACSS Round
type DacssStateMap struct {
	sync.Mutex
	// For each specific ACSS round, we have a DacssState
	DacssStateForRound sync.Map // key:ACSSRoundID, value: DaccsState
}

type DaccsState struct {
	DacssData            DacssData // Commitments, encrypted shares & ephemeral public key of the dealer
	RBCState             RBCState  // Storage for RBC protocol that is executed in this ACSS round
	ShareRecoveryOngoing bool      // Indicates per dACSS round whether the Share Recovery is in process
}

type DacssData struct {
	Commitments []byte
	// Key = hex of PubKey receiving node
	// Value = encrypted share for receiving node
	ShareMap map[string][]byte
	// hex of ephemeral public key of the dealer
	DealerEphemeralPubKey string
}

func (m *DacssStateMap) Get(acssRoundID ACSSRoundID) (*DaccsState, bool, error) {
	value, found := m.DacssStateForRound.Load(acssRoundID)
	if !found {
		return nil, false, nil
	}

	dacssState, ok := value.(*DaccsState)
	if !ok {
		// If the value is found but is not of type *DaccsState, return an error.
		return nil, true, errors.New("value found, but type is not *DaccsState")
	}

	// If everything is ok, return the dacssState, true for found, and no error.
	return dacssState, true, nil
}

func (m *DacssStateMap) UpdateDacssData(acssRoundID ACSSRoundID, dacssData DacssData) error {
	// Attempt to load the existing DaccsState for the given acssRoundID
	value, found := m.DacssStateForRound.Load(acssRoundID)
	if found {
		// If found, type assert the value to *DaccsState
		existingState, ok := value.(*DaccsState)
		if ok {
			// Update the DacssData field of the existing DaccsState
			existingState.DacssData = dacssData
			// Store the updated DaccsState back into the map
			m.DacssStateForRound.Store(acssRoundID, existingState)
			return nil
		} else {
			return errors.New("value found, but type is not *DaccsState")
		}
	} else {
		// If the DaccsState for the acssRoundID does not exist, create a new one and store it
		m.DacssStateForRound.Store(acssRoundID, &DaccsState{DacssData: dacssData})
		return nil
	}
}

// Stores the shares that the node receives during the DPSS protocol.
type PSSShareStore struct {
	sync.Mutex
	Shares map[int]curves.Scalar // Map of shares. K: index of the owner of the share, V: the actual share.
}

// PSSRoundDetails represents all the details in a round for the DPSS protocol.
type PSSRoundDetails struct {
	// ID for the PSS.
	PssID string
	// Index & PubKey of the dealer Node.
	Dealer NodeDetails
}

// ACSSRoundID defines the ID of a single ACSS that can be running within the DPSS process
type ACSSRoundID string

type ACSSRoundDetails struct {
	PSSRoundDetails PSSRoundDetails // PSSRoundDetails represented in a string
	ACSSCount       int             // number of ACSS round this is in the PSS
}

// FIXME
func (acssRoundDetails *ACSSRoundDetails) ToACSSRoundID() ACSSRoundID {
	// TODO implement
	return "test"
}

// n -> total number of nodes
// t = f -> number of *max* malicious nodes
// k = f + 1 > reconstruction threshold
type CommitteeParams struct {
	N int
	K int
	T int // = K - 1
}

// PSSMessage represents a message in the DPSS protocol
type PSSMessage struct {
	PSSRoundDetails PSSRoundDetails // Round ID of the current execution of the DPSS protocol.
	Type            string          // Phase of the protocol in which the message belongs to.
	Data            []byte          // Actual data in the message.
}

func CreatePSSMessage(pssRoundDetails PSSRoundDetails, phase string, data []byte) PSSMessage {
	return PSSMessage{
		PSSRoundDetails: pssRoundDetails,
		Type:            phase,
		Data:            data,
	}
}

// Generates a new PSSRoundID for a given index.
func NewPssID(index big.Int) string {
	return strings.Join([]string{"PSS", index.Text(16)}, Delimiter3)
}
