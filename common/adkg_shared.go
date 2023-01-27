package common

import (
	"crypto/sha256"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

func HashByte(msg []byte) []byte {
	sum := sha256.Sum256(msg)
	return sum[:]
}

type ADKGMetadata struct {
	Commitments map[int][]curves.Point
	T           []int
}
