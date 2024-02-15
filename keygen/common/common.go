package common

import (
	"encoding/binary"
	"math/big"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/curves"
)

func GetSetBits(n, val int) []int {
	l := make([]int, 0)
	for i := 1; i <= n; i++ {
		if HasBit(val, i) {
			l = append(l, i)
		}
	}
	return l
}

func IntToByteValue(val int) []byte {
	var byteVal [8]byte
	binary.BigEndian.PutUint64(byteVal[:], uint64(val))
	return byteVal[:]
}

func ByteToIntValue(val []byte) int {
	intVal := binary.BigEndian.Uint64(val)
	return int(intVal)
}

func HasBit(n int, pos int) bool {
	val := n & (1 << pos)
	return (val > 0)
}

func SetBit(n int, pos int) int {
	n |= (1 << pos)
	return n
}

func CountBit(n int) int {
	count := 0
	for n > 0 {
		n &= (n - 1)
		count++
	}
	return count
}

func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func CurvePointToPoint(p curves.Point, c common.CurveName) common.Point {
	bytes := p.ToAffineUncompressed()
	if c == common.ED25519 {
		xBytes := reverse(bytes[:32])
		yBytes := reverse(bytes[32:])
		return common.Point{
			X: *new(big.Int).SetBytes(xBytes),
			Y: *new(big.Int).SetBytes(yBytes),
		}
	} else {
		xBytes := bytes[1:33]
		yBytes := bytes[33:]
		return common.Point{
			X: *new(big.Int).SetBytes(xBytes),
			Y: *new(big.Int).SetBytes(yBytes),
		}
	}
}

func Contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
