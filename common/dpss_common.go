package common

import "github.com/coinbase/kryptology/pkg/core/curves"

// PSSParticipant represents a party in the DPSS protocol.
type PSSParticipant interface {
	// Returns if the participant is from the old committee.
	IsOldNode() bool
	// Returns if the participant is from the new committee.
	IsNewNode() bool
	// Sends a given message to the given node.
	Send(msg PSSMessage, node PSSParticipant)
	// Returns the ID of the participant.
	ID() int
	// Returns the private key of the participant.
	PrivateKey() curves.Scalar
	// Returns the public key of the given participant.
	PublicKey(index int) curves.Point
	// Returns the params of the network in which the participant is connected.
	Params(newCommittee bool) (n int, k int, f int)
}

// Represents a message in the DPSS protocol
type PSSMessage interface {
	Kind() MessageType
	Process(sender KeygenNodeDetails, self PSSParticipant)
}
