package common

import (
	"crypto/sha256"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

// Computes the SHA256 hash of the message.
func HashByte(msg []byte) []byte {
	sum := sha256.Sum256(msg)
	return sum[:]
}

type ADKGMetadata struct {
	Commitments map[int][]curves.Point
	T           []int
}
