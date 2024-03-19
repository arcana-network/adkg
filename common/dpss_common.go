package common

import (
	"errors"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/sharing"
	"github.com/torusresearch/bijson"
)

// PSSParticipant is the interface that covers all the participants inside the
// DPSS protocol
type PSSParticipant interface {
	// For PSS state
	State() *PSSNodeState
	// Defines if the current node belongs to the old or new committee.
	IsNewNode() bool
	// Obtains the public key from a node in the old or new committee. The
	// committee is defined by the flag fromNewCOmmittee.
	GetPublicKeyFor(idx int, fromNewCommittee bool) curves.Point
	// Obtains the parameters of the protocols for the committee for which the
	// current node belongs. n = number of nodes, k = reconstruction threshold
	// t = max number of malicious nodes.
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
	AcssStore       *AcssStateMap // State for the separate ACSS rounds
	DualAcssStarted bool          // Flag whether the DualAcss part of the protocol has started.
}

// Stores the information of the shares for a given ACSS Round
type AcssStateMap struct {
	sync.Mutex
	// For each specific ACSS round, we have a DacssState
	AcssStateForRound sync.Map // key:ACSSRoundID, value: AccsState
}

type AccsState struct {
	// Hash of Commitments, encrypted shares & ephemeral public key of the dealer
	// in order to be able to compare it with received data
	AcssDataHash []byte
	// Storage for RBC protocol that is executed in this ACSS round
	RBCState RBCState
	// TODO do we want to extract ImplicateInformationSlice, VerifiedRecoveryShares & ShareRecoveryOngoing to a separate state field?
	// Information about all the possible implicate flows that can be in progress
	ImplicateInformationSlice []ImplicateInformation
	// Received verified shares from other nodes, needed to recover Node's share in Implicate phase
	VerifiedRecoveryShares map[int]*sharing.ShamirShare
	// Indicates per ACSS round whether the Share Recovery is in process
	ShareRecoveryOngoing bool
	// Indicates wether the shares held by the current party are valid at the end of the ACSS protocol.
	ValidShareOutput bool
	// Shares received from each dealer
	ReceivedShares map[string]*sharing.ShamirShare
	// Own commitment received in the RBC (see Line 203, DACSS protocol)
	OwnCommitments *sharing.FeldmanVerifier
	// Received commitments
	ReceivedCommitments map[int]bool
	// Commitment database that counts how many times a commitment has been received.
	CommitmentDatabase map[string]*CommitmentStore
}

func (state *AccsState) GetStoreForCommitment(
	fingerprint string,
	commitments []Point,
) *CommitmentStore {
	_, found := state.CommitmentDatabase[fingerprint]
	if !found {
		return &CommitmentStore{
			Commitments: commitments,
			Count:       0,
		}
	}
	return state.CommitmentDatabase[fingerprint]
}

func (state *AccsState) FindThresholdCommitment(threshold int) *CommitmentStore {
	for _, commitment := range state.CommitmentDatabase {
		if commitment.Count >= threshold {
			return commitment
		}
	}
	return nil
}

type CommitmentStore struct {
	Commitments []Point
	Count       int
}

type AccsStateUpdater func(*AccsState)

type AcssData struct {
	Commitments []byte
	// Key = hex of PubKey receiving node
	// Value = encrypted share for receiving node
	ShareMap map[string][]byte
	// hex of ephemeral public key of the dealer
	DealerEphemeralPubKey string
}

func (a *AcssData) IsUninitialized() bool {
	// Check if all the fields are in their zero value state.
	// For Commitments and DealerEphemeralPubKey, this means checking if they are empty.
	// For ShareMap, check if it is nil or has no entries.
	return len(a.Commitments) == 0 && len(a.DealerEphemeralPubKey) == 0 && (a.ShareMap == nil || len(a.ShareMap) == 0)
}

func (m *AcssStateMap) Get(acssRoundID ACSSRoundID) (*AccsState, bool, error) {
	value, found := m.AcssStateForRound.Load(acssRoundID)
	if !found {
		return nil, false, nil
	}

	dacssState, ok := value.(*AccsState)
	if !ok {
		// If the value is found but is not of type *AccsState, return an error.
		return nil, true, errors.New("value found, but type is not *AccsState")
	}

	// If everything is ok, return the dacssState, true for found, and no error.
	return dacssState, true, nil
}

// Generic updater for a AcssState within the AcssStateMap
func (m *AcssStateMap) UpdateAccsState(acssRoundID ACSSRoundID, updater AccsStateUpdater) error {
	// Attempt to load the existing AccsState for the given acssRoundID
	value, found := m.AcssStateForRound.Load(acssRoundID)
	var existingState *AccsState

	if found {
		// If found, type assert the value to *AccsState
		var ok bool
		existingState, ok = value.(*AccsState)
		if !ok {
			return errors.New("value found, but type is not *AccsState")
		}
	} else {
		// If not found, create a new AccsState
		existingState = &AccsState{
			RBCState: RBCState{
				ReceivedEcho:   make(map[int]bool),
				ReceivedReady:  make(map[int]bool),
				EchoDatabase:   make(map[string]*EchoStore),
				IsReadyMsgSent: false,
			},
			VerifiedRecoveryShares: make(map[int]*sharing.ShamirShare),
			ReceivedShares:         make(map[string]*sharing.ShamirShare),
		}
	}

	// Apply the updater function to the existing or new AccsState
	updater(existingState)

	// Store the updated or new AccsState back into the map
	m.AcssStateForRound.Store(acssRoundID, existingState)

	return nil
}

// Data that is needed to execute Implicate flow
// This flow might have a waiting period in between until it has access to the shareMap
// to be able to continue the flow once the shareMap has been received, this information is stored
type ImplicateInformation struct {
	SymmetricKey    []byte // Compressed Affine Point
	Proof           []byte // Contains d, R, S
	SenderPubkeyHex string // Hex of compressed Affine Point
	AcssDataHash    []byte // Hash of received AcssData
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

func (acssRoundDetails *ACSSRoundDetails) ToACSSRoundID() ACSSRoundID {
	// Convert ACSSRoundDetails to a string representation to be used as an ID
	return ACSSRoundID(strings.Join([]string{
		string(acssRoundDetails.PSSRoundDetails.Dealer.ToNodeDetailsID()),
		strconv.Itoa(acssRoundDetails.ACSSCount),
	}, Delimiter1))
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

func HashAcssData(data AcssData) ([]byte, error) {

	//TODO: is there a better way to convert it into bytes?
	// is this conversion unique??
	bytes, err := bijson.Marshal(data)

	if err != nil {
		return nil, err
	}

	return HashByte(bytes), nil
}
