package common

import (
	"math/big"
	"strings"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

// PSSParticipant represents a party in the DPSS protocol.
// PssNode implement PSSParticipant, like KeygenNode implements DKGParticipant
type PSSParticipant interface {
	// TODO fix this to the state it should be for DPSS
	ParticipantState
	// Returns if the participant is from the old committee.
	IsOldNode() bool
	// Returns if the participant is from the new committee.
	IsNewNode() bool
	// Sends a given message to the given node.
	// TODO can we change the type of msg
	// TODO can KeygenNodeDetails be renamed? (has implications for keygen)
	Send(n KeygenNodeDetails, msg DKGMessage) error
	// Returns the ID of the participant.
	ID() int
	// Returns the private key of the participant.
	PrivateKey() curves.Scalar
	// Returns the public key of the given participant.
	PublicKey(index int, isNewCommittee bool) curves.Point
	// Returns the params of the network in which the participant is connected.
	Params(newCommittee bool) (n int, k int, f int)
	// Returns the nodes of the new or old committee
	Nodes(isNewCommittee bool) map[NodeDetailsID]KeygenNodeDetails
}

func GenerateDPSSID(rindex, noOfRandoms big.Int) ADKGID {
	index := strings.Join([]string{rindex.Text(16), noOfRandoms.Text(16)}, Delimiter2)
	return ADKGID(strings.Join([]string{"DPSS", index}, Delimiter3))
}

// GetSessionStoreFromRoundID extracts the session store of a given participant
// for the specified round ID.
func GetSessionStoreFromRoundID(roundID RoundID, p PSSParticipant) (*ADKGSession, error) {
	r := &RoundDetails{}
	err := r.FromID(roundID)
	if err != nil {
		log.Debugf("Error parsing round id, err=%s", err)
		return nil, err
	}
	sessionStore := p.State().SessionStore
	// create default session to use below
	s, _ := sessionStore.GetOrSet(r.ADKGID, DefaultADKGSession())
	return s, nil
}
