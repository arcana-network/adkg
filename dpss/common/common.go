package dpsscommon

import "encoding/binary"

// Can be reused from ADKG
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
