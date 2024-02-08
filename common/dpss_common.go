package common

import (
	"math/big"
	"strings"

	log "github.com/sirupsen/logrus"
)

// PSSParticipant represents a party in the DPSS protocol.
// type PSSParticipant interface {
// 	// Returns if the participant is from the old committee.
// 	IsOldNode() bool
// 	// Returns if the participant is from the new committee.
// 	IsNewNode() bool
// 	// Sends a given message to the given node.
// 	Send(msg PSSMessage, node PSSParticipant)
// 	// Returns the ID of the participant.
// 	ID() int
// 	// Returns the private key of the participant.
// 	PrivateKey() curves.Scalar
// 	// Returns the public key of the given participant.
// 	PublicKey(index int) curves.Point
// 	// Returns the params of the network in which the participant is connected.
// 	Params(newCommittee bool) (n int, k int, f int)
// }

// Represents a message in the DPSS protocol
// type PSSMessage interface {
// 	Kind() MessageType
// 	Process(sender KeygenNodeDetails, self DkgParticipant)
// }

func GenerateDPSSID(rindex, noOfRandoms big.Int) ADKGID {
	index := strings.Join([]string{rindex.Text(16), noOfRandoms.Text(16)}, Delimiter2)
	return ADKGID(strings.Join([]string{"DPSS", index}, Delimiter3))
}

// GetSessionStoreFromRoundID extracts the session store of a given participant
// for the specified round ID.
func GetSessionStoreFromRoundID(roundID RoundID, p DkgParticipant) (*ADKGSession, error) {
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
