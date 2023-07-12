package keyset

import "encoding/binary"

func Predicate(Tj, Ti []byte) bool {
	a := binary.BigEndian.Uint64(Ti)
	b := binary.BigEndian.Uint64(Tj)
	if b&a == b { // Checking if T_j is a subset of T_i
		return true
	} else {
		return false
	}
}
