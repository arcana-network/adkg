package common

import (
	"encoding/hex"
	"io"
	"math/big"
	"reflect"
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

func TestBitKeeper(t *testing.T) {
	tests := []struct {
		name   string
		values []int
	}{
		{"works as expected", []int{1, 2, 3}},
		{"works as expected", []int{1, 2, 3, 4, 5, 6}},
		{"works as expected", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"works as expected", []int{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := 0
			for _, v := range tt.values {
				n = SetBit(n, v)
			}
			for _, v := range tt.values {
				if !HasBit(n, v) {
					t.Errorf("Bit set but not found value=%d", v)
				}
			}

			if !reflect.DeepEqual(GetSetBits(len(tt.values), n), tt.values) {
				t.Errorf("values did not match value=%x", GetSetBits(3, n))
			}

			if CountBit(n) != len(tt.values) {
				t.Errorf("length mismatch got=%d want=%d", CountBit(n), len(tt.values))
			}
		})
	}
}
func TestCurvePointToPoint(t *testing.T) {
	eP := EPoint{}
	pp := CurvePointToPoint(eP)
	if hex.EncodeToString(pp.X.Bytes()) != "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798" ||
		hex.EncodeToString(pp.Y.Bytes()) != "483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8" {
		t.Fatal("Should be able to convert to point")
	}
}

type EPoint struct{}

func (E EPoint) Random(reader io.Reader) curves.Point {
	panic("implement me")
}

func (E EPoint) Hash(bytes []byte) curves.Point {
	panic("implement me")
}

func (E EPoint) Identity() curves.Point {
	panic("implement me")
}

func (E EPoint) Generator() curves.Point {
	panic("implement me")
}

func (E EPoint) IsIdentity() bool {
	panic("implement me")
}

func (E EPoint) IsNegative() bool {
	panic("implement me")
}

func (E EPoint) IsOnCurve() bool {
	panic("implement me")
}

func (E EPoint) Double() curves.Point {
	panic("implement me")
}

func (E EPoint) Scalar() curves.Scalar {
	panic("implement me")
}

func (E EPoint) Neg() curves.Point {
	panic("implement me")
}

func (E EPoint) Add(rhs curves.Point) curves.Point {
	panic("implement me")
}

func (E EPoint) Sub(rhs curves.Point) curves.Point {
	panic("implement me")
}

func (E EPoint) Mul(rhs curves.Scalar) curves.Point {
	panic("implement me")
}

func (E EPoint) Equal(rhs curves.Point) bool {
	panic("implement me")
}

func (E EPoint) Set(x, y *big.Int) (curves.Point, error) {
	panic("implement me")
}

func (E EPoint) ToAffineCompressed() []byte {
	panic("implement me")
}

func (E EPoint) ToAffineUncompressed() []byte {
	ret, err := hex.DecodeString("0479be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8")
	if err != nil {
		panic(err)
	}
	return ret
}

func (E EPoint) FromAffineCompressed(bytes []byte) (curves.Point, error) {
	panic("implement me")
}

func (E EPoint) FromAffineUncompressed(bytes []byte) (curves.Point, error) {
	panic("implement me")
}

func (E EPoint) CurveName() string {
	panic("implement me")
}

func (E EPoint) SumOfProducts(points []curves.Point, scalars []curves.Scalar) curves.Point {
	panic("implement me")
}
