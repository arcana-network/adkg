package aba

import (
	"encoding/hex"
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

func TestHash(t *testing.T) {
	curve := curves.K256()

	g := curve.Point.Identity()
	h := curve.Point.Generator()
	type args struct {
		g       curves.Point
		h       curves.Point
		gTilde  curves.Point
		hTilde  curves.Point
		gI      curves.Point
		gITilde curves.Point
		curve   *curves.Curve
	}

	tests := []struct {
		name string
		args args
		want string
	}{

		{"test-1", args{g, g, g, g, g, g, curve}, "184c4ce04b0b6bf5f17b13a931e695f0157fa9f90fcfc71090f6ea5f61eb760c"},
		{"test-2", args{g, h, g, h, g, h, curve}, "1b696f79ef23c62c16c8e34d9862c67f40021aa1b4c344f9c2aa4c5f52a184d9"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Hash(tt.args.g, tt.args.h, tt.args.gTilde, tt.args.hTilde, tt.args.gI, tt.args.gITilde, tt.args.curve)
			if hex.EncodeToString(got.Bytes()) != tt.want {
				t.Errorf("Hash() = %v, want %v", hex.EncodeToString(got.Bytes()), tt.want)
			}
		})
	}
}

func TestLagrangeCoeffs(t *testing.T) {

	curve := curves.K256()
	type args struct {
		identities []int
		curve      *curves.Curve
	}
	tests := []struct {
		name string
		args args
		want map[int]string
	}{
		{
			"test-1",
			args{[]int{1, 12312, 3}, curve},
			map[int]string{
				1:     "1993be8a088694d8ad82319d8f2b65df0f3875cecedddabbd6675d2f86e6ca5e",
				12312: "ad553b888f56fa38bae0235d43189c5ef870c99401d41e1dcb6e516b2e78166b",
				3:     "391705ed682270ee979dab052dbbfdc0b3059d83de96a7621dfcaff21ad76079",
			},
		},
		{
			"test-2",
			args{[]int{1212, 12312, 3}, curve},
			map[int]string{
				1212:  "aab02cdd0a414ec73563ab386cbafa48f9dd524a6c5b3731e14ee0cf08ee884e",
				12312: "832b6798adc1ae5381e66cc7cb645e95d40ddbe5e17afaf492be700bbef1e608",
				3:     "d2246b8a47fd02e548b5e7ffc7e0a71ea7728b9d10bb0e510b976c3ed88c142d",
			},
		},
		{
			"test-3",
			args{[]int{1322, 12312, 1123}, curve},
			map[int]string{
				1322:  "10b24b4afcc6acfa3c69b4c2ad5688067a518a093edc8d04e00ed8a30170a033",
				12312: "168f4de49b3ff03f8c3507f34b535829233e81c45a782f94fa49f4b5dd81af83",
				1123:  "d8be66d067f962c63761434a07561fcf1d1ed11915f3e3a1e5799133f143f18c",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := LagrangeCoeffs(tt.args.identities, tt.args.curve)
			for k, v := range got {
				s := hex.EncodeToString(v.Bytes())
				t.Log(k, s)
			}

			for k, v := range got {
				g := hex.EncodeToString(v.Bytes())
				if g != tt.want[k] {
					t.Errorf("LagrangeCoeffs()[%d] = %v, want %v", k, g, tt.want[k])
				}
			}
		})
	}
}
