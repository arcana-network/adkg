package common

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/arcana-network/dkgnode/common/sharing"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
	"golang.org/x/crypto/sha3"
)

type MessageType string

type ADKGID string

func GenerateADKGID(index big.Int) ADKGID {
	return ADKGID(strings.Join([]string{"ADKG", index.Text(16)}, Delimiter3))
}
func NewADKGID(index big.Int, curve CurveName) ADKGID {
	baseStr := "ADKG"
	if curve == ED25519 {
		baseStr = strings.Join([]string{"ADKG", string(ED25519)}, Delimiter5)
	}
	return ADKGID(strings.Join([]string{baseStr, index.Text(16)}, Delimiter3))
}

func (id *ADKGID) GetCurve() (CurveName, error) {
	str := string(*id)
	substrs := strings.Split(str, Delimiter3)

	if len(substrs) != 2 {
		return "", errors.New("could not parse dkgid")
	}

	ids := strings.Split(substrs[0], Delimiter5)
	if len(ids) == 1 {
		return SECP256K1, nil
	}
	if len(ids) == 2 && ids[1] == string(ED25519) {
		return ED25519, nil
	}
	return "", errors.New("invalid curve")
}

func (id *ADKGID) GetIndex() (big.Int, error) {
	str := string(*id)
	substrs := strings.Split(str, Delimiter3)

	if len(substrs) != 2 {
		return *new(big.Int), errors.New("could not parse dkgid")
	}

	index, ok := new(big.Int).SetString(substrs[1], 16)
	if !ok {
		return *new(big.Int), errors.New("could not get back index from dkgid")
	}

	return *index, nil
}

func ADKGIDFromRoundID(r RoundID) (ADKGID, error) {
	d := RoundDetails{}
	err := d.FromID(r)
	if err != nil {
		return ADKGID(""), err
	}

	return d.ADKGID, nil
}

func CreateRound(ADKGID ADKGID, dealer int, kind string) RoundID {
	r := RoundDetails{
		ADKGID,
		dealer,
		kind,
	}
	return r.ID()
}

type RoundDetails struct {
	ADKGID ADKGID
	Dealer int
	Kind   string
}

func (d *RoundDetails) ID() RoundID {
	return RoundID(strings.Join([]string{string(d.ADKGID), d.Kind, strconv.Itoa(d.Dealer)}, Delimiter4))
}

func (d *RoundDetails) FromID(roundID RoundID) error {
	s := string(roundID)
	substrings := strings.Split(s, Delimiter4)

	if len(substrings) != 3 {
		return fmt.Errorf("expected length of 2, got=%d", len(substrings))
	}
	d.ADKGID = ADKGID(substrings[0])
	d.Kind = substrings[1]
	index, err := strconv.Atoi(substrings[2])
	if err != nil {
		return err
	}
	d.Dealer = index
	return nil
}

func (r *RoundID) Leader() (big.Int, error) {
	str := string(*r)
	substrs := strings.Split(str, Delimiter4)

	if len(substrs) != 3 {
		return *new(big.Int), errors.New("could not parse round id")
	}

	index, ok := new(big.Int).SetString(substrs[2], 16)
	if !ok {
		return *new(big.Int), errors.New("could not get back index from round id")
	}

	return *index, nil
}

type RoundID string

type NodeState struct {
	KeygenStore  *SharingStoreMap
	SessionStore *ADKGSessionStore
	ABAStore     *ABAStoreMap
}

type ADKGSessionStore struct {
	Map sync.Map
}

func (m *ADKGSessionStore) Get(r ADKGID) (session *ADKGSession, found bool) {
	inter, found := m.Map.Load(r)
	session, _ = inter.(*ADKGSession)
	return
}

func (store *ADKGSessionStore) GetOrSetIfNotComplete(r ADKGID, input *ADKGSession) (*ADKGSession, bool) {
	inter, found := store.GetOrSet(r, input)
	if found {
		if inter == nil {
			return inter, true
		}
	}
	return inter, false
}

func (store *ADKGSessionStore) GetOrSet(r ADKGID, input *ADKGSession) (session *ADKGSession, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	session, _ = inter.(*ADKGSession)
	return
}

func (store *ADKGSessionStore) Complete(r ADKGID) {
	store.Map.Store(r, nil)
}

type SharingStoreMap struct {
	Map sync.Map
}

func (m *SharingStoreMap) Get(r RoundID) (keygen *SharingStore, found bool) {
	inter, found := m.Map.Load(r)
	keygen, _ = inter.(*SharingStore)
	return
}

func (store *SharingStoreMap) GetOrSetIfNotComplete(r RoundID, input *SharingStore) (keygen *SharingStore, complete bool) {
	inter, found := store.GetOrSet(r, input)
	if found {
		if inter == nil {
			return inter, true
		}
	}
	return inter, false
}

func (store *SharingStoreMap) GetOrSet(r RoundID, input *SharingStore) (keygen *SharingStore, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	keygen, _ = inter.(*SharingStore)
	return
}
func (store *SharingStoreMap) Complete(r RoundID) {
	store.Map.Store(r, nil)
}

func GetCStore(keygen *SharingStore, s string) *CStore {
	c, found := keygen.CStore[s]
	if !found {
		keygen.CStore[s] = &CStore{}
		c = keygen.CStore[s]
	}
	return c
}

type CStore struct {
	EC        int
	RC        int
	ReadySent bool
}

type SharingStore struct {
	sync.Mutex
	RoundID    RoundID
	State      RBCState
	CStore     map[string]*CStore
	ReadyStore []infectious.Share
	Started    bool
}

type ADKGSession struct {
	sync.Mutex
	// All keysets
	T          map[int]int
	TProposals map[int]int
	TPrime     int
	// Share mapping of acss dealer -> share
	S                      map[int]sharing.ShamirShare
	C                      map[int][]curves.Point
	PubKeyShares           map[int]curves.Point
	PubKeySharesUnverified map[int]PubKeyShare
	Over                   bool
	BFTDecided             bool
	Share                  *big.Int
	Commitments            ADKGMetadata
	Decisions              map[int]int
	ABAComplete            bool
	ABAStarted             []int
	KeyderivationStarted   bool
}

type PubKeyShare struct {
	R     []byte
	S     []byte
	Share []byte
}

type RBCState struct {
	Phase         phase
	ReceivedEcho  map[int]bool
	ReceivedReady map[int]bool
}

func Keccak256(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

type KeyPair struct {
	PublicKey  curves.Point
	PrivateKey curves.Scalar
}

type KeygenDetails struct {
	CurrentKeyIndex int
}

func GetADKGIDFromRoundID(roundID RoundID) (ADKGID, error) {
	r := &RoundDetails{}
	err := r.FromID(roundID)
	if err != nil {
		log.WithError(err).Infof("ParseRoundID()")
		return ADKGID(""), err
	}

	return r.ADKGID, nil
}

type DkgParticipant interface {
	// For ADKG state
	ParticipantState
	// Get Protocol n, k and f
	Params() (n int, k int, t int)
	// Node Index
	ID() int
	// Get self details
	Details() KeygenNodeDetails
	// Send message to a node
	Send(n KeygenNodeDetails, msg DKGMessage) error
	// Send message to all connected nodes
	Broadcast(msg DKGMessage)
	// Receive message to self
	ReceiveMessage(sender KeygenNodeDetails, msg DKGMessage)
	// Get public key of a node
	PublicKey(index int) curves.Point
	// Get map of connected nodes
	Nodes() map[NodeDetailsID]KeygenNodeDetails
	// Get self private key
	PrivateKey() curves.Scalar
	// Get public params for a curve, say g1 and g2
	CurveParams(name string) (curves.Point, curves.Point)
	// Receiving BFT message to broadcast
	ReceiveBFTMessage(DKGMessage)
	// Cleanup session store
	Cleanup(id ADKGID)
	// Store completed share
	StoreCompletedShare(index big.Int, si big.Int, c CurveName)
	// Store commitment to shares
	StoreCommitment(index big.Int, metadata ADKGMetadata, c CurveName)
}

type ParticipantState interface {
	State() *NodeState
}

func DefaultADKGSession() *ADKGSession {
	s := ADKGSession{
		C:                      make(map[int][]curves.Point),
		S:                      make(map[int]sharing.ShamirShare),
		PubKeyShares:           make(map[int]curves.Point),
		PubKeySharesUnverified: make(map[int]PubKeyShare),
		Decisions:              make(map[int]int),
		T:                      make(map[int]int),
		TProposals:             make(map[int]int),
		TPrime:                 0,
		ABAStarted:             []int{},
	}

	return &s
}

type ABAStoreMap struct {
	Map sync.Map
}

func (m *ABAStoreMap) Get(r RoundID) (keygen *ABAState, found bool) {
	inter, found := m.Map.Load(r)
	keygen, _ = inter.(*ABAState)
	return
}

func (store *ABAStoreMap) GetOrSetIfNotComplete(r RoundID, input *ABAState) (keygen *ABAState, complete bool) {
	inter, found := store.GetOrSet(r, input)
	if found {
		if inter == nil {
			return inter, true
		}
	}
	return inter, false
}

func (store *ABAStoreMap) GetOrSet(r RoundID, input *ABAState) (keygen *ABAState, found bool) {
	inter, found := store.Map.LoadOrStore(r, input)
	keygen, _ = inter.(*ABAState)
	return
}

func (store *ABAStoreMap) Complete(r RoundID) {
	store.Map.Store(r, nil)
}

type ABAState struct {
	sync.Mutex
	Started      map[int]bool
	Round        int
	CoinShares   map[int]curves.Point
	EstValues    map[int]map[int][]int
	AuxValues    map[int]map[int][]int
	EstValues2   map[int]map[int][]int
	AuxValues2   map[int]map[int][]int
	AuxsetValues map[int]map[int][]int
	EstSent2     map[int]map[int]bool
	EstSent      map[int]map[int]bool
	AuxsetSent   map[int]bool
	BinValues    map[int][]int
	BinValues2   map[int][]int
}

func DefaultABAStore() *ABAState {
	s := ABAState{
		Started:      make(map[int]bool),
		CoinShares:   make(map[int]curves.Point),
		EstValues:    make(map[int]map[int][]int),
		AuxValues:    make(map[int]map[int][]int),
		EstValues2:   make(map[int]map[int][]int),
		AuxValues2:   make(map[int]map[int][]int),
		AuxsetValues: make(map[int]map[int][]int),
		EstSent2:     make(map[int]map[int]bool),
		EstSent:      make(map[int]map[int]bool),
		AuxsetSent:   make(map[int]bool),
		BinValues:    make(map[int][]int),
		BinValues2:   make(map[int][]int),
	}
	return &s
}

func (s *ABAState) GetCoinShares() map[int]curves.Point {
	return s.CoinShares
}

func (s *ABAState) SetCoinShare(i int, p curves.Point) {
	s.CoinShares[i] = p
}
func (s *ABAState) IncrementRound() {
	s.Round = s.Round + 1
}

func (s *ABAState) GetRound() int {
	return s.Round
}
func (s *ABAState) SetStarted(r int) {
	s.Started[r] = true
}

func (s *ABAState) GetStarted(r int) bool {
	return s.Started[r]
}

// kind can  be "est" or "est2" or "auxset", will panic otherwise
func (s *ABAState) Sent(kind string, r, v int) bool {
	switch kind {
	case "est":
		return s.EstSent[r][v]
	case "est2":
		return s.EstSent2[r][v]
	case "auxset":
		return s.AuxsetSent[r]
	// case "est1"
	default:
		panic(fmt.Sprintf("Invalid values set to store.GetSent(%s)", kind))
	}
}

// kind can  be "est" or "est2" or "auxset", will panic otherwise
func (s *ABAState) SetSent(kind string, r, v int) {
	switch kind {
	case "est":
		if s.EstSent[r] == nil {
			s.EstSent[r] = make(map[int]bool)
		}
		s.EstSent[r][v] = true
		return
	case "est2":
		if s.EstSent2[r] == nil {
			s.EstSent2[r] = make(map[int]bool)
		}
		s.EstSent2[r][v] = true
		return
	case "auxset":
		s.AuxsetSent[r] = true
		return
	default:
		panic(fmt.Sprintf("Invalid values set to store.SetSent(%s)", kind))
	}
}

// kind can  be "bin" or "bin2", will panic otherwise
func (s *ABAState) GetBin(kind string, r int) []int {
	switch kind {
	case "bin":
		return s.BinValues[r]
	case "bin2":
		return s.BinValues2[r]
	default:
		panic(fmt.Sprintf("Invalid values set to store.GetBin(%s)", kind))
	}
}

// kind can  be "bin" or "bin2", will panic otherwise
func (s *ABAState) SetBin(kind string, r, v int) {
	switch kind {
	case "bin":
		if s.BinValues[r] == nil {
			s.BinValues[r] = []int{}
		}
		s.BinValues[r] = append(s.BinValues[r], v)
		return
	case "bin2":
		if s.BinValues2[r] == nil {
			s.BinValues2[r] = []int{}
		}
		s.BinValues2[r] = append(s.BinValues2[r], v)
		return
	}
}

// kind can be "est", "est2", "aux", "aux2", "auxset", will panic otherwise
func (s *ABAState) Values(kind string, r, v int) []int {
	switch kind {
	case "est":
		return s.EstValues[r][v]
	case "est2":
		return s.EstValues2[r][v]
	case "aux":
		return s.AuxValues[r][v]
	case "aux2":
		return s.AuxValues2[r][v]
	case "auxset":
		return s.AuxsetValues[r][v]
	default:
		panic(fmt.Sprintf("Invalid values set to store.Get(%s)", kind))
	}
}

// kind can be "est", "est2", "aux", "aux2", "auxset", will panic otherwise
func (s *ABAState) SetValues(kind string, r, v, node int) {
	switch kind {
	case "est":
		if s.EstValues[r] == nil {
			s.EstValues[r] = map[int][]int{}
		}
		s.EstValues[r][v] = append(s.EstValues[r][v], node)
		return
	case "est2":
		if s.EstValues2[r] == nil {
			s.EstValues2[r] = map[int][]int{}
		}
		s.EstValues2[r][v] = append(s.EstValues2[r][v], node)
		return
	case "aux":
		if s.AuxValues[r] == nil {
			s.AuxValues[r] = map[int][]int{}
		}
		s.AuxValues[r][v] = append(s.AuxValues[r][v], node)
		return
	case "aux2":
		if s.AuxValues2[r] == nil {
			s.AuxValues2[r] = map[int][]int{}
		}
		s.AuxValues2[r][v] = append(s.AuxValues2[r][v], node)
		return
	case "auxset":
		if s.AuxsetValues[r] == nil {
			s.AuxsetValues[r] = map[int][]int{}
		}
		s.AuxsetValues[r][v] = append(s.AuxsetValues[r][v], node)
		return
	default:
		panic(fmt.Sprintf("Invalid values set to store.Set(%s)", kind))
	}
}
