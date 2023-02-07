package commitment

import (
	"github.com/coinbase/kryptology/pkg/core/curves"

	"github.com/arcana-network/dkgnode/common/sharing"
)

type Value struct {
	Points []curves.Point
}

type Commitment []byte
type Opening []byte

type SRS struct{}

type Scheme interface {
	GenerateCommitmentValue() Value
	Setup() SRS
	Commit(commitmentValue Value) Commitment
	Open() Opening
	Check(opening Opening) bool // == Verify()?
}

type Verifier interface {
	Verify(share *sharing.ShamirShare) error
	Commitments() []curves.Point
}
