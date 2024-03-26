package keyset

import "encoding/binary"

func Predicate(Ti, Tj []byte) bool {
	b := binary.BigEndian.Uint64(Tj)
	a := binary.BigEndian.Uint64(Ti)
	if b&a == b { // Checking if T_j is a subset of T_i
		return true
	} else {
		return false
	}
}
