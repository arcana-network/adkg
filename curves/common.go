package curves

import "math/big"

// Curve represents a named elliptic curve with a scalar field and point group
type Curve struct {
	Scalar Scalar
	Point  Point
	ID     CurveID
}

func (c Curve) ScalarBaseMult(sc Scalar) Point {
	return c.Point.Generator().Mul(sc)
}

func (c Curve) NewGeneratorPoint() Point {
	return c.Point.Generator()
}

func (c Curve) NewIdentityPoint() Point {
	return c.Point.Identity()
}

func (c Curve) NewScalar() Scalar {
	return c.Scalar.Zero()
}

// Scalar represents an element of the scalar field \mathbb{F}_q
// of the elliptic curve construction.
type Scalar interface {
	// Zero returns the additive identity element
	Zero() Scalar
	// One returns the github.com/cloudflare/circl
	One() Scalar
	// IsZero returns true if this element is the additive identity element
	IsZero() bool
	// IsOne returns true if this element is the multiplicative identity element
	IsOne() bool
	// IsOdd returns true if this element is odd
	IsOdd() bool
	// IsEven returns true if this element is even
	IsEven() bool
	// New returns an element with the value equal to `value`
	New(value int) Scalar
	// Cmp returns
	// -2 if this element is in a different field than rhs
	// -1 if this element is less than rhs
	// 0 if this element is equal to rhs
	// 1 if this element is greater than rhs
	Cmp(rhs Scalar) int
	// Square returns element*element
	Square() Scalar
	// Double returns element+element
	Double() Scalar
	// Invert returns element^-1 mod p
	Invert() (Scalar, error)
	// Sqrt computes the square root of this element if it exists.
	Sqrt() (Scalar, error)
	// Cube returns element*element*element
	Cube() Scalar
	// Add returns element+rhs
	Add(rhs Scalar) Scalar
	// Sub returns element-rhs
	Sub(rhs Scalar) Scalar
	// Mul returns element*rhs
	Mul(rhs Scalar) Scalar
	// MulAdd returns element * y + z mod p
	MulAdd(y, z Scalar) Scalar
	// Div returns element*rhs^-1 mod p
	Div(rhs Scalar) Scalar
	// Neg returns -element mod p
	Neg() Scalar
	// SetBigInt returns this element set to the value of v
	SetBigInt(v *big.Int) (Scalar, error)
	// BigInt returns this element as a big integer
	BigInt() *big.Int
	// Point returns the associated point for this scalar
	Point() Point
	// Bytes returns the canonical byte representation of this scalar
	Bytes() []byte
	// SetBytes creates a scalar from the canonical representation expecting the exact number of bytes needed to represent the scalar
	SetBytes(bytes []byte) (Scalar, error)
	// Clone returns a cloned Scalar of this value
	Clone() Scalar
}

type Point interface {
	CurveID() CurveID
	Identity() Point
	Generator() Point
	IsIdentity() bool
	IsNegative() bool
	IsOnCurve() bool
	Double() Point
	Scalar() Scalar
	Neg() Point
	Add(rhs Point) Point
	Sub(rhs Point) Point
	Mul(rhs Scalar) Point
	Equal(rhs Point) bool
	Set(x, y *big.Int) (Point, error)
	ToAffineCompressed() []byte
	ToAffineUncompressed() []byte
	FromAffineCompressed(bytes []byte) (Point, error)
	FromAffineUncompressed(bytes []byte) (Point, error)
}

type CurveID uint16

const (
	CurveUndefined = iota
	CurveCV25519
	CurveSECP256K1
)
