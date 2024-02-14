package curves

import (
	"errors"
	"fmt"
	"math/big"

	secp256k12 "github.com/consensys/gnark-crypto/ecc/secp256k1"
	"github.com/consensys/gnark-crypto/ecc/secp256k1/fp"
)

func CurveK256() *Curve {
	tmp := Curve{
		Scalar: &ScalarK256{},
		Point:  &PointK256{},
		ID:     CurveSECP256K1,
	}
	return &tmp
}

type ScalarK256 struct {
	value fp.Element
}

type PointK256 struct {
	value secp256k12.G1Affine
}

func (s *ScalarK256) Zero() Scalar {
	return &ScalarK256{
		value: fp.NewElement(0),
	}
}

func (s *ScalarK256) One() Scalar {
	return &ScalarK256{
		value: fp.One(),
	}
}

func (s *ScalarK256) IsZero() bool {
	return s.value.IsZero()
}

func (s *ScalarK256) IsOne() bool {
	return s.value.IsOne()
}

func (s *ScalarK256) IsOdd() bool {
	return s.value.Bytes()[0]&1 == 1
}

func (s *ScalarK256) IsEven() bool {
	return s.value.Bytes()[0]&1 == 0
}

func (s *ScalarK256) New(value int) Scalar {
	return &ScalarK256{
		value: fp.NewElement(uint64(value)),
	}
}

func (s *ScalarK256) Cmp(rhs Scalar) int {
	r, ok := rhs.(*ScalarK256)
	if ok {
		return s.value.Cmp(&r.value)
	} else {
		return -2
	}
}

func (s *ScalarK256) Square() Scalar {
	tmp := fp.NewElement(0)
	tmp.Square(&s.value)
	return &ScalarK256{
		value: tmp,
	}
}

func (s *ScalarK256) Double() Scalar {
	tmp := fp.NewElement(0)
	tmp.Double(&s.value)
	return &ScalarK256{
		value: tmp,
	}
}

func (s *ScalarK256) Invert() (Scalar, error) {
	tmp := fp.NewElement(0)
	tmp.Inverse(&s.value)
	return &ScalarK256{
		value: tmp,
	}, nil
}

func (s *ScalarK256) Sqrt() (Scalar, error) {
	tmp := fp.NewElement(0)
	tmp.Sqrt(&s.value)
	return &ScalarK256{
		tmp,
	}, nil
}

func (s *ScalarK256) Cube() Scalar {
	tmp := fp.NewElement(0)
	tmp.Square(&s.value)
	tmp.Mul(&tmp, &s.value)
	return &ScalarK256{
		tmp,
	}
}

func (s *ScalarK256) Add(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarK256)
	if ok {
		tmp := fp.NewElement(0)
		tmp.Add(&s.value, &r.value)
		return &ScalarK256{
			value: tmp,
		}
	} else {
		return nil
	}
}

func (s *ScalarK256) Sub(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarK256)
	if ok {
		tmp := fp.NewElement(0)
		tmp.Sub(&s.value, &r.value)
		return &ScalarK256{
			value: tmp,
		}
	} else {
		return nil
	}
}

func (s *ScalarK256) Mul(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarK256)
	if ok {
		tmp := fp.NewElement(0)
		tmp.Mul(&s.value, &r.value)
		return &ScalarK256{
			value: tmp,
		}
	} else {
		return nil
	}
}

func (s *ScalarK256) MulAdd(y, z Scalar) Scalar {
	return s.Mul(y).Add(z)
}

func (s *ScalarK256) Div(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarK256)
	if ok {
		inverted := fp.NewElement(0)
		inverted.Inverse(&r.value)
		tmp := fp.NewElement(0)
		tmp.Mul(&s.value, &inverted)
		return &ScalarK256{
			value: tmp,
		}
	} else {
		return nil
	}
}

func (s *ScalarK256) Neg() Scalar {
	tmp := fp.NewElement(0)
	tmp.Neg(&s.value)
	return &ScalarK256{
		value: tmp,
	}
}

func (s *ScalarK256) SetBigInt(v *big.Int) (Scalar, error) {
	if v == nil {
		return nil, fmt.Errorf("'v' cannot be nil")
	}
	tmp := fp.NewElement(0)
	tmp.SetBigInt(v)
	return &ScalarK256{
		value: tmp,
	}, nil
}

func (s *ScalarK256) BigInt() *big.Int {
	return s.value.BigInt(new(big.Int))
}

func (s *ScalarK256) Bytes() []byte {
	tmp := s.value.Bytes()
	return tmp[:]
}

func (s *ScalarK256) SetBytes(bytes []byte) (Scalar, error) {
	tmp := fp.NewElement(0)
	tmp.SetBytes(bytes)
	return &ScalarK256{
		value: tmp,
	}, nil
}

func (s *ScalarK256) Point() Point {
	return new(PointK256).Identity()
}

func (s *ScalarK256) Clone() Scalar {
	tmp := fp.NewElement(0)
	tmp.Set(&s.value)
	return &ScalarK256{
		value: tmp,
	}
}

func (p *PointK256) Identity() Point {
	return &PointK256{
		value: secp256k12.G1Affine{
			X: fp.NewElement(0),
			Y: fp.NewElement(0),
		},
	}
}

func (p *PointK256) Generator() Point {
	_, acc := secp256k12.Generators()
	return &PointK256{
		value: acc,
	}
}

func (p *PointK256) IsIdentity() bool {
	return p.value.X.IsZero() && p.value.Y.IsZero()
}

func (p *PointK256) IsNegative() bool {
	zero := fp.NewElement(0)
	return p.value.Y.Cmp(&zero) >= 1
}

func (p *PointK256) IsOnCurve() bool {
	return p.value.IsOnCurve()
}

func (p *PointK256) Double() Point {
	var tmp secp256k12.G1Affine
	tmp.Double(&p.value)
	return &PointK256{value: tmp}
}

func (p *PointK256) Scalar() Scalar {
	return new(ScalarK256).Zero()
}

func (p *PointK256) Neg() Point {
	var tmp secp256k12.G1Affine
	tmp.Neg(&p.value)
	return &PointK256{value: tmp}
}

func (p *PointK256) Add(rhs Point) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*PointK256)
	if ok {
		var tmp secp256k12.G1Affine
		tmp.Add(&p.value, &r.value)
		return &PointK256{value: tmp}
	} else {
		return nil
	}
}

func (p *PointK256) Sub(rhs Point) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*PointK256)
	if ok {
		var tmp secp256k12.G1Affine
		tmp.Sub(&p.value, &r.value)
		return &PointK256{value: tmp}
	} else {
		return nil
	}
}

func (p *PointK256) Mul(rhs Scalar) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*ScalarK256)
	if ok {
		var tmp secp256k12.G1Affine
		tmp.ScalarMultiplication(&p.value, r.BigInt())
		return &PointK256{value: tmp}
	} else {
		return nil
	}
}

func (p *PointK256) Equal(rhs Point) bool {
	r, ok := rhs.(*PointK256)
	if ok {
		return p.value.Equal(&r.value)
	} else {
		return false
	}
}

func (p *PointK256) Set(x, y *big.Int) (Point, error) {
	var tmp secp256k12.G1Affine
	tmp.X.SetBigInt(x)
	tmp.Y.SetBigInt(y)
	return &PointK256{value: tmp}, nil
}

func (p *PointK256) ToAffineCompressed() []byte {
	x := make([]byte, 33)
	{
		tmp := p.value.Y.BigInt(new(big.Int))
		if tmp.Mod(tmp, big.NewInt(2)).Cmp(big.NewInt(1)) == 0 {
			x[0] = 3
		} else {
			x[0] = 2
		}
	}
	xBytes := p.value.X.Bytes()
	copy(x[1:], xBytes[:])
	return x
}

func (p *PointK256) ToAffineUncompressed() []byte {
	buf := make([]byte, 65)
	buf[0] = 4
	tmp := p.value.RawBytes()
	copy(buf[1:], tmp[:])
	return buf
}

func (p *PointK256) FromAffineCompressed(bytes []byte) (Point, error) {
	if len(bytes) != 33 {
		return nil, errors.New("invalid sign byte")
	}
	ȳ := bytes[0] - 2
	_, b := secp256k12.CurveCoefficients()

	x := fp.NewElement(0)
	x.SetBytes(bytes[1:])

	// y² = x³ + ax + b, and we have to solve for y, so we start with x³ first
	y := fp.NewElement(0)

	// x³
	y.Square(&x)
	y.Mul(&y, &x)

	// + ax
	/*{
		tmp := fp.NewElement(0)
		tmp.Mul(&a, &x)
		y.Add(&y, &tmp)
	}*/

	// +b
	y.Add(&y, &b)

	// y = √(y²)
	y.Sqrt(&y)

	{
		tmp := y.BigInt(new(big.Int))
		tmp.Mod(tmp, big.NewInt(2))
		// ???, need to optimize this anyhow
		// if y % 2 != ȳ (i.e. the sign of y), set y = prime - y
		if tmp.Cmp(big.NewInt(int64(ȳ))) != 0 {
			mod := fp.NewElement(0)
			mod.SetBigInt(fp.Modulus())
			// y = p - y
			y.Sub(&mod, &y)
		}
	}

	return &PointK256{
		value: secp256k12.G1Affine{
			X: x,
			Y: y,
		},
	}, nil
}

func (p *PointK256) FromAffineUncompressed(bytes []byte) (Point, error) {
	if len(bytes) != 65 || bytes[0] != 4 {
		return nil, errors.New("invalid sign byte")
	}
	/*
		var tmp secp256k12.G1Affine
		_, err := tmp.SetBytes(bytes[1:])
		if err != nil {
			return nil, err
		}*/
	x := fp.NewElement(0)
	x.SetBytes(bytes[1:33])
	y := fp.NewElement(0)
	y.SetBytes(bytes[33:65])
	tmp := secp256k12.G1Affine{
		X: x,
		Y: y,
	}
	return &PointK256{value: tmp}, nil
}

func (p *PointK256) CurveID() CurveID {
	return CurveSECP256K1
}

/*
func (p *PointK256) CurveName() string {
	return elliptic.P256().Params().Name
}

func (p *PointK256) SumOfProducts(points []Point, scalars []Scalar) Point {
	nPoints := make([]*native.EllipticPoint, len(points))
	nScalars := make([]*native.Field, len(scalars))
	for i, pt := range points {
		ptv, ok := pt.(*PointK256)
		if !ok {
			return nil
		}
		nPoints[i] = ptv.value
	}
	for i, sc := range scalars {
		s, ok := sc.(*ScalarK256)
		if !ok {
			return nil
		}
		nScalars[i] = s.value
	}
	value := p256n.P256PointNew()
	_, err := value.SumOfProducts(nPoints, nScalars)
	if err != nil {
		return nil
	}
	return &PointK256{value}
}

func (p *PointK256) X() *native.Field {
	return p.value.GetX()
}

func (p *PointK256) Y() *native.Field {
	return p.value.GetY()
}

func (p *PointK256) Params() *elliptic.CurveParams {
	return elliptic.P256().Params()
}

func (p *PointK256) MarshalBinary() ([]byte, error) {
	return pointMarshalBinary(p)
}

func (p *PointK256) UnmarshalBinary(input []byte) error {
	pt, err := pointUnmarshalBinary(input)
	if err != nil {
		return err
	}
	ppt, ok := pt.(*PointK256)
	if !ok {
		return fmt.Errorf("invalid point")
	}
	p.value = ppt.value
	return nil
}

func (p *PointK256) MarshalText() ([]byte, error) {
	return pointMarshalText(p)
}

func (p *PointK256) UnmarshalText(input []byte) error {
	pt, err := pointUnmarshalText(input)
	if err != nil {
		return err
	}
	ppt, ok := pt.(*PointK256)
	if !ok {
		return fmt.Errorf("invalid point")
	}
	p.value = ppt.value
	return nil
}

func (p *PointK256) MarshalJSON() ([]byte, error) {
	return pointMarshalJson(p)
}

func (p *PointK256) UnmarshalJSON(input []byte) error {
	pt, err := pointUnmarshalJson(input)
	if err != nil {
		return err
	}
	P, ok := pt.(*PointK256)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	p.value = P.value
	return nil
}
*/
