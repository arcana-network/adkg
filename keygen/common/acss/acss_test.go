package acss

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"math/rand"
	"reflect"
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/vivint/infectious"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestEncryptAES(t *testing.T) {
	type args struct {
		key       []byte
		plaintext []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"invalid keysize", args{[]byte(randSeq(5)), []byte("abc")}, true},
		{"correct keysize", args{[]byte("XVlBzgbaiCMRAjWw"), []byte("abc")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptAES(tt.args.key, tt.args.plaintext)

			if (err != nil) != tt.wantErr {
				t.Errorf("encryptAES() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestEncryptAndDecryptAES(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		plaintext string
	}{
		{"16-key-size", randSeq(16), randSeq(20)},
		{"24-key-size", randSeq(24), randSeq(30)},
		{"32-key-size", randSeq(32), randSeq(40)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("len=%d", len(tt.key))
			cipher, err := encryptAES([]byte(tt.key), []byte(tt.plaintext))
			if err != nil {
				t.Errorf("encryptAES() error = %v", err)
				return
			}

			plain, err := decryptAES([]byte(tt.key), cipher)
			if err != nil {
				t.Errorf("decryptAES() error = %v", err)
				return
			}
			if string(plain) != tt.plaintext {
				t.Errorf("encryptAES() = %v, want %v", string(plain), tt.plaintext)
			}
		})
	}
}

func TestEncodeAndDecode(t *testing.T) {
	type args struct {
		n     int
		k     int
		input string
	}
	tests := []struct {
		name string
		args args
	}{
		{"n=4,f=2", args{4, 2, randSeq(30)}},
		{"n=5,f=3", args{5, 3, randSeq(30)}},
		{"n=8,f=2", args{8, 3, randSeq(30)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			f, err := infectious.NewFEC(tt.args.k, tt.args.n)
			if err != nil {
				t.Errorf("NewFEC(k,n) error = %v", err)
				return
			}
			shares, err := Encode(f, []byte(tt.args.input))
			if err != nil {
				t.Errorf("Encode(k,n) error = %v", err)
				return
			}

			out, err := Decode(f, shares[0:tt.args.k])
			if err != nil {
				t.Errorf("Decode(k,n) error = %v", err)
				return
			}

			if string(out) != tt.args.input {
				t.Errorf("Encode and decode didn't match want=%s got=%s", string(out), tt.args.input)
			}
		})
	}
}

func Test_pkcs7Pad(t *testing.T) {
	type args struct {
		b         []byte
		blocksize int
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"odd number of padding", args{[]byte{0}, 4}, []byte{0, 3, 3, 3}, false},
		{"even number of padding", args{[]byte{0, 0}, 4}, []byte{0, 0, 2, 2}, false},
		{"more than 1 digit padding", args{[]byte{0}, 11}, []byte{0, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pkcs7Pad(tt.args.b, tt.args.blocksize)
			if (err != nil) != tt.wantErr {
				t.Errorf("pkcs7Pad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pkcs7Pad() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pkcs7Unpad(t *testing.T) {
	type args struct {
		b         []byte
		blocksize int
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"odd number of padding", args{[]byte{0, 3, 3, 3}, 4}, []byte{0}, false},
		{"even number of padding", args{[]byte{0, 0, 2, 2}, 4}, []byte{0, 0}, false},
		{"more than 1 digit padding", args{[]byte{0, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}, 11}, []byte{0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pkcs7Unpad(tt.args.b, tt.args.blocksize)
			if (err != nil) != tt.wantErr {
				t.Errorf("pkcs7Unpad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pkcs7Unpad() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSharedKey(t *testing.T) {

	curve := curves.K256()
	k256 := curve.Point.Generator().(*curves.PointK256)
	dealerPoint := k256.Generator()
	var scalar = new(curves.ScalarK256).New(8)
	key := SharedKey(scalar, dealerPoint)
	if hex.EncodeToString(key[:]) != "be2b01947193835b2a70e0bed841b4dd8926e75f6a7427ba3d90a1774beacac6" {
		t.Fatal("Should be able to generate a shared key")
	}
}

func TestPredicate(t *testing.T) {
	curve := curves.K256()
	k256 := curve.Point.Generator().(*curves.PointK256)
	dealerPoint := k256.Generator()
	var scalar = new(curves.ScalarK256).New(8)
	key := SharedKey(scalar, dealerPoint)
	commit := dealerPoint.ToAffineCompressed()
	_, _, b := Predicate(key[:], []byte("foobarfoobarfoobarfoobarfoobarfoobar"), commit, 1, curve)
	t.Log(b)
}

func TestSplit(t *testing.T) {
	curve := curves.K256()
	var scalar = new(curves.ScalarK256).New(8)

	split, shares, err := Split(scalar, 1, 1, curve, cryptorand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(shares[0].Value) != "0000000000000000000000000000000000000000000000000000000000000008" {
		t.Fatal("Should be able to split")
	}
	if hex.EncodeToString(split.Commitments[0].ToAffineUncompressed()) != "042f01e5e15cca351daff3843fb70f3c2f0a1bdd05e5af888a67784ef3e10a2a015c4da8a741539949293d082a132d13b4c2e213d6ba5b7617b5da2cb76cbde904" {
		t.Fatal("Should be able to split")
	}

}
