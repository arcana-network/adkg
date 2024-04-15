package common

import (
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"

	ownsharing "github.com/arcana-network/dkgnode/common/sharing"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/sharing"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
	"github.com/vivint/infectious"
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
	// Broker needed to start the next batch of DPSS
	// In all the Test Nodes it is mocked to simply return nil
	// In the PSSNode, it returns the MessageBroker
	GetMessageBroker() *MessageBroker
	// Get node details by id, only checks in list of old nodes
	OldNodeDetailsByID(id int) (NodeDetails, error)
	// Get B
	GetBatchCount() int
	// Get (G, H) for specific curve
	CurveParams(curveName string) (curves.Point, curves.Point)
}

// Constants for state naming. Used for loging.
const (
	AcssStateType     = "ACSS"
	BatchRecStateType = "BatchRec"
	RbcStateType      = "RBC"
)

/*
B is fixed, we dont want to start too many dpss
at once since P2P network will start dropping messages

PSS-1 => 0 - 51 => B = 51 / 3 => 17 ACSS
PSS-2 => 52 -103 => B = 51 / 3 => 17 ACSS
*/

// PSSNodeState represents the internal state of a node that participates in
// possibly multiple DPSS protocol. There is an storage for the different
// sub-protocols in the DPSS: ACSS, RBC
type PSSNodeState struct {
	BatchReconStore *BatchRecStoreMap // State for the separate batch reconstruction rounds
	AcssStore       *AcssStateMap     // State for the separate ACSS rounds
	ShareStore      *PSSShareStore    // Storage of shares for the DPSS protocol.
	ABAStore        *AbaStateMap
	KeysetStore     *KeysetStateMap // Could have reused ACSS here
	PSSStore        *PSSStateMap
}

// Clean completely cleans the state of a node for a given PSSRound.
func (state *PSSNodeState) Clean(pssRound PSSRoundDetails) error {
	err := cleanMap(&state.AcssStore.AcssStateForRound, pssRound, Delimiter1)
	if err != nil {
		return err
	}

	err = cleanMap(&state.BatchReconStore.BatchReconStateForRound, pssRound, Delimiter2)
	if err != nil {
		return err
	}

	state.ShareStore.NewShares = make([]curves.Scalar, 0)
	state.ShareStore.OldShares = make([]PrivKeyShare, 0)

	return nil
}

// cleanMap removes the entries of the syncMap that contains the given PSSRoundDetails
// as part of the key, which is separated with the given delimiter.
func cleanMap(mapStore *sync.Map, pssRound PSSRoundDetails, delimiter string) error {
	var err error
	mapStore.Range(
		func(key, value any) bool {
			// Parses the key into PSSRoundDetails || AcssCount.
			keyAcssRoundID := key.(ACSSRoundID)
			splittedKey := strings.Split(
				string(keyAcssRoundID), delimiter,
			)

			if len(splittedKey) != 2 {
				err = errors.New("the split process was not done correctly")
				return false
			}

			// Takes the PSSRoundDetails from the key.
			pssDetailsKey := splittedKey[0]
			if pssDetailsKey == pssRound.ToString() {
				mapStore.Delete(
					key,
				)
			}
			return true
		},
	)

	return err
}

// Stores all the information for each separate batch reconstruction.
type BatchRecStoreMap struct {
	sync.Mutex
	// For a specific batch reconstructoin round, we have a BatchRecState
	BatchReconStateForRound sync.Map // key: DPSSBatchRecID, value: BatchRecState
}

// Stores all the information related to the batch reconstruction protocol for a given round.
type BatchRecState struct {
	UStore              map[int]curves.Scalar // Stores the shares [u_i] sent by a given party.
	ReconstructedUStore map[int]curves.Scalar // Stores the restructured u_i sent by given party
	SentLocalCompMsg    bool                  // Tells wether the node has sent a PrivRecMsg.
	SentPubMsg          bool                  // Tells wether the node has sent a PubRecMsg.
}

// Counts the ammount of U shares received by the party.
func (batchState *BatchRecState) CountReceivedU() int {
	counter := 0
	for range batchState.UStore {
		counter++
	}

	return counter
}

// Counts the ammount of ReconstructedU shares received by the party.
func (batchState *BatchRecState) CountReconstructedReceivedU() int {
	counter := 0
	for range batchState.ReconstructedUStore {
		counter++
	}

	return counter
}

// Obtains the batch reconstruction information for a given batch reconstruction ID.
func (store *BatchRecStoreMap) Get(id BatchRecID) (*BatchRecState, bool, error) {
	value, found := store.BatchReconStateForRound.Load(id)
	if !found {
		return nil, false, nil
	}

	state, ok := value.(*BatchRecState)
	if !ok {
		return nil, true, errors.New("error while retrieving the batch state")
	}

	return state, true, nil
}

// UpdateBatchRecState updates the state stored under the provided ID. The updating
// is defined by the updater function provided in the parameters. If the ID is not
// stored in the
func (store *BatchRecStoreMap) UpdateBatchRecState(batchID BatchRecID, updater func(*BatchRecState)) error {
	value, found := store.BatchReconStateForRound.Load(batchID)
	var state *BatchRecState
	if found {
		var ok bool
		state, ok = value.(*BatchRecState)
		if !ok {
			return errors.New("error retrieving the batch state to update it")
		}
	} else {
		state = &BatchRecState{
			UStore:              make(map[int]curves.Scalar),
			ReconstructedUStore: make(map[int]curves.Scalar),
		}
	}

	updater(state)
	store.BatchReconStateForRound.Store(batchID, state)
	return nil
}

/* ------- PSS state start ------- */
// Stores the information of the shares for a given PSS Round
type PSSStateMap struct {
	Map sync.Map // key:PSSRoundID, value: PSSState
}

type ACSSKeysetMap struct {
	// Own super keyset, not limited to n-f
	TPrime int
	// nodeIndex => Commitments (used in ABA coin tossing)
	CommitmentStore map[int][]curves.Point
	// nodeIndex => share
	ShareStore map[int]*sharing.ShamirShare
}

// Shared data for a PSS Round
type PSSState struct {
	sync.Mutex
	T              map[int]int // nodeIndex => verified keyset (limited to n-f)
	TProposals     map[int]int // nodeIndex => unverified keyset
	PSSID          PSSID
	KeysetMap      map[int]*ACSSKeysetMap // acssCount => ACCKeysetMap
	KeysetProposed bool
	ABAStarted     []int
	ABAComplete    bool
	Decisions      map[int]int
	HIMStarted     bool
}

func (state *PSSState) GetTSet(n, t int) []int {
	keysets := make([][]int, 0)
	for k, v := range state.Decisions {
		if v == 1 {
			keysets = append(keysets, GetSetBits(n, state.T[k]))
		}
	}

	T := Union(keysets...)
	sort.Ints(T)
	T = T[:(n - t)]
	return T
}

func (state *PSSState) GetSharesFromT(T []int, alpha int, curve *curves.Curve) []curves.Scalar {
	shares := []curves.Scalar{}
	for i := range alpha {
		val := state.KeysetMap[i]
		for _, j := range T {
			log.Debug("val.ShareStore", val.ShareStore, "curve", curve)
			s := val.ShareStore[j]
			log.Debug("s.value", s, j, val.ShareStore)
			share, err := curve.Scalar.SetBytes(s.Value)
			if err != nil {
				log.Error("scalar.setBytes:", err)
				continue
			}
			shares = append(shares, share)
		}
	}
	return shares
}

func Contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func HasBit(n int, pos int) bool {
	val := n & (1 << pos)
	return (val > 0)
}

func GetSetBits(n, val int) []int {
	l := make([]int, 0)
	for i := 1; i <= n; i++ {
		if HasBit(val, i) {
			l = append(l, i)
		}
	}
	return l
}

func Union(args ...[]int) []int {
	if len(args) == 0 {
		return []int{}
	}

	a := args[0]
	m := make(map[int]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, s := range args[1:] {
		for _, item := range s {
			if _, ok := m[item]; !ok {
				a = append(a, item)
				m[item] = true
			}
		}
	}
	return a
}

/*
PSS-1-Dealer-1-Alpha-1
Alpha = 0 -> B/n-2t

keysetMap Structure:
Alpha-0 => {
	Node - 1 => {

	},
	Node - X => {
		TPrime => int
		CommmitmentStore => {
			NodeIndex => Commitment
		}
		ShareStore => {
			NodeIndex => share
		}
	}
},
.
.
.
Alpha-X => {
	Node - 1 => {

	},
	Node - X => {
		TPrime => int
		CommmitmentStore => {
			NodeIndex => Commitment
		}
		ShareStore => {
			NodeIndex => share
		}
	}
}
*/

func (state *PSSState) GetKeysetMap(id int) *ACSSKeysetMap {
	_, ok := state.KeysetMap[id]
	if !ok {
		state.KeysetMap[id] = &ACSSKeysetMap{
			TPrime:          0,
			CommitmentStore: make(map[int][]curves.Point),
			ShareStore:      make(map[int]*sharing.ShamirShare),
		}
	}

	return state.KeysetMap[id]
}

// Threshold = n-t, alpha = B/n-2t
func (state *PSSState) CheckForThresholdCompletion(alpha, threshold int) (int, bool) {
	T := 0
	Tset := make([]int, 0)
	if len(state.KeysetMap) == alpha {
		for _, v := range state.KeysetMap {
			Tset = append(Tset, v.TPrime)
		}

		T = Tset[0]
		for i := 1; i < alpha; i += 1 {
			T = T & Tset[i]
		}
		if CountBit(T) >= threshold {
			return T, true
		}
	}
	return 0, false
}

// FIXME: Deduplicate this v
func CountBit(n int) int {
	count := 0
	for n > 0 {
		n &= (n - 1)
		count++
	}
	return count
}

func GetDefaultPSSState(id PSSID) *PSSState {
	s := PSSState{
		PSSID:      id,
		KeysetMap:  make(map[int]*ACSSKeysetMap),
		T:          make(map[int]int),
		TProposals: make(map[int]int),
		Decisions:  make(map[int]int),
	}
	return &s
}

func (m *PSSStateMap) Get(r PSSID) (state *PSSState, found bool) {
	inter, found := m.Map.Load(r)
	state, _ = inter.(*PSSState)
	return
}

func (store *PSSStateMap) GetOrSetIfNotComplete(r PSSID) (keygen *PSSState, complete bool) {
	inter, found := store.GetOrSet(r, GetDefaultPSSState(r))
	if found {
		if inter == nil {
			return inter, true
		}
	}
	return inter, false
}

func (store *PSSStateMap) GetOrSet(r PSSID, input *PSSState) (keygen *PSSState, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	keygen, _ = inter.(*PSSState)
	return
}
func (store *PSSStateMap) Complete(r PSSID) {
	store.Map.Store(r, nil)
}

func (store *PSSStateMap) Delete(r PSSID) {
	store.Map.Delete(r)
}

/* ------- PSS state end ------- */

/* ------- Keyset start ------- */
// Stores the information of the shares for a given Keyset Round
type KeysetStateMap struct {
	Map sync.Map // key:PSSRoundID, value: KeysetState
}

type KeysetState struct {
	sync.Mutex
	RoundID  PSSRoundID
	RBCState NewRBCState
}

type NewRBCState struct {
	Phase         phase
	ReceivedEcho  map[int]bool
	ReceivedReady map[int]bool
	ReadySent     bool
	EchoStore     map[string]*EchoStore
	ReadyStore    []infectious.Share
}

// Get or create echo store according to the id
func (s *KeysetState) GetEchoStore(id string, share infectious.Share, hash []byte) *EchoStore {
	if _, ok := s.RBCState.EchoStore[id]; !ok {
		s.RBCState.EchoStore[id] = &EchoStore{
			Shard:       share,
			HashMessage: hash,
			Count:       0,
		}
	}
	return s.RBCState.EchoStore[id]
}

func (s *KeysetState) FindThresholdEchoStore(threshold int) *EchoStore {
	for _, v := range s.RBCState.EchoStore {
		if v.Count >= threshold {
			return v
		}
	}
	return nil
}

func (m *KeysetStateMap) Get(r PSSRoundID) (keygen *KeysetState, found bool) {
	inter, found := m.Map.Load(r)
	keygen, _ = inter.(*KeysetState)
	return
}

func (store *KeysetStateMap) GetOrSetIfNotComplete(r PSSRoundID, input *KeysetState) (keygen *KeysetState, complete bool) {
	inter, found := store.GetOrSet(r, input)
	if found {
		if inter == nil {
			return inter, true
		}
	}
	return inter, false
}

func (store *KeysetStateMap) GetOrSet(r PSSRoundID, input *KeysetState) (keygen *KeysetState, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	keygen, _ = inter.(*KeysetState)
	return
}
func (store *KeysetStateMap) Complete(r PSSRoundID) {
	store.Map.Store(r, nil)
}

func (store *KeysetStateMap) Delete(r PSSRoundID) {
	store.Map.Delete(r)
}

/* ------- Keyset end ------- */

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
	ReceivedShare *sharing.ShamirShare
	//random secret shared by the dealers in the start of the protocol
	//only to be stored by the dealer
	RandomSecretShared map[ACSSRoundID]*curves.Scalar
	// Hex representation of the has of the own commitment computed in the RBC (see Line 203, DACSS protocol)
	OwnCommitmentsHash string
	// Received commitments
	ReceivedCommitments map[int]bool
	// Commitment database that counts how many times a commitment has been received.
	CommitmentCount map[string]int
	// Tells wether the node has broadcasted the commitment
	CommitmentSent bool
}

func (state *AccsState) FindThresholdCommitment(threshold int) (string, bool) {
	for hash, count := range state.CommitmentCount {
		if count >= threshold {
			return hash, true
		}
	}
	return "", false
}

type AccsStateUpdater func(*AccsState)

type AcssData struct {
	// Compressed commitments computed when the share was created.
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
			VerifiedRecoveryShares:    make(map[int]*sharing.ShamirShare),
			RandomSecretShared:        make(map[ACSSRoundID]*curves.Scalar),
			ReceivedCommitments:       make(map[int]bool),
			CommitmentCount:           make(map[string]int),
			ImplicateInformationSlice: make([]ImplicateInformation, 0),
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

// Stores the shares that the node uses during the DPSS protocol. That means that
// This storage stores the old shares (for the old nodes) but also the new shares
// at the end of the DPSS protocol (for the new nodes).
type PSSShareStore struct {
	sync.Mutex
	NewShares []curves.Scalar // Map of shares at the end of the protocol.
	OldShares []PrivKeyShare  // Map of shares at the beginning of the protocol.
}

func (store *PSSShareStore) Initialize(storeSize int) {
	store.NewShares = make([]curves.Scalar, storeSize)
	store.OldShares = make([]PrivKeyShare, storeSize)
}

// Returns the UserID for the owners of the shares in the same order that they
// were provided at the beggining of the protocol.
func (store *PSSShareStore) GetUserIDs() []string {
	result := make([]string, len(store.OldShares))
	for i, privKeyShare := range store.OldShares {
		result[i] = privKeyShare.UserIdOwner
	}
	return result
}

// PSSRoundDetails represents all the details in a round for the DPSS protocol.
type PSSRoundDetails struct {
	PssID     string      // ID for the PSS.
	Dealer    NodeDetails // Index & PubKey of the dealer Node.
	BatchSize int         // Number of shares that will be converted from one degree to another in one batch.
}

func (pssRoundDetails PSSRoundDetails) ToString() string {
	return strings.Join([]string{
		string(pssRoundDetails.Dealer.ToNodeDetailsID()),
		string(pssRoundDetails.PssID),
	}, Delimiter1)
}

func CreatePSSRound(pssID string, dealer NodeDetails, batchSize int) PSSRoundDetails {
	return PSSRoundDetails{
		pssID,
		dealer,
		batchSize,
	}
}

type PSSID string

func GeneratePSSID(index big.Int) PSSID {
	return PSSID(strings.Join([]string{"PSS", index.Text(16)}, Delimiter3))
}

func (round *PSSRoundDetails) ToRoundID() PSSRoundID {
	return PSSRoundID(strings.Join([]string{
		string(round.PssID),
		strconv.Itoa(round.Dealer.Index),
	}, Delimiter1))
}

type PSSRoundID string

// ACSSRoundID defines the ID of a single ACSS that can be running within the DPSS process
type ACSSRoundID string

type ACSSRoundDetails struct {
	PSSRoundDetails PSSRoundDetails // PSSRoundDetails represented in a string
	ACSSCount       int             // number of ACSS round this is in the PSS
}

func (acssRoundDetails *ACSSRoundDetails) ToACSSRoundID() ACSSRoundID {
	// Convert ACSSRoundDetails to a string representation to be used as an ID
	return ACSSRoundID(strings.Join([]string{
		acssRoundDetails.PSSRoundDetails.ToString(),
		strconv.Itoa(acssRoundDetails.ACSSCount),
	}, Delimiter1))
}

// BatchRecID represents the ID for a single round of the batch reconstruction in the DPSS protocol.
type BatchRecID string

// Defines the information for a round in the batch reconstruction.
type DPSSBatchRecDetails struct {
	PSSRoundDetails PSSRoundDetails // PSS instance to which the Batch reconstruction belongs.
	BatchRecCount   int             // ID for the batch reconstruction round.
}

// ToBathcRecID transforms the details of a batch reocnstruction round into an
// ID by encoding its fields.
func (details *DPSSBatchRecDetails) ToBatchRecID() BatchRecID {
	return BatchRecID(
		strings.Join(
			[]string{
				details.PSSRoundDetails.ToString(),
				strconv.Itoa(details.BatchRecCount),
			},
			Delimiter2,
		),
	)
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
func NewPssID(index big.Int) PSSID {
	return PSSID(strings.Join([]string{"PSS", index.Text(16)}, Delimiter3))
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

// PrivKeyShare represents a share of a private key.
type PrivKeyShare struct {
	UserIdOwner string                 // Owner of the private key.
	Share       ownsharing.ShamirShare // Share of the private key.
}
