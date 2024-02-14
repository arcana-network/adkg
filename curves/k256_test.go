package curves

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"testing"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
)

func TestP256PointCompression(t *testing.T) {
	kryptologyk256 := curves.K256()
	kryptologyP1, err := kryptologyk256.Point.FromAffineCompressed(hexutil.MustDecode("0x02f54ba86dc1ccb5bed0224d23f01ed87e4a443c47fc690d7797a13d41d2340e1a"))
	if err != nil {
		panic(err)
	}
	assert.True(t, kryptologyP1.IsOnCurve())

	var customP1 PointK256
	customP2, err := customP1.FromAffineUncompressed(kryptologyP1.ToAffineUncompressed())
	if err != nil {
		panic(err)
	}
	fmt.Printf("CustomP2 Uncompressed Serialized:\t%x\nOriginal Point Serialized:\t\t\t%x\n", customP2.ToAffineUncompressed(), kryptologyP1.ToAffineUncompressed())
	assert.True(t, bytes.Equal(customP2.ToAffineUncompressed(), kryptologyP1.ToAffineUncompressed()))
	fmt.Printf("CustomP2 Compressed Serialized:\t\t%x\nOriginal Point Serialized:\t\t\t%x\n", customP2.ToAffineCompressed(), kryptologyP1.ToAffineCompressed())
	assert.True(t, bytes.Equal(customP2.ToAffineCompressed(), kryptologyP1.ToAffineCompressed()))

	customP3, err := customP1.FromAffineCompressed(kryptologyP1.ToAffineCompressed())
	if err != nil {
		panic(err)
	}
	fmt.Printf("CustomP3 Uncompressed Serialized:\t%x\nOriginal Point Serialized:\t\t\t%x\n", customP3.ToAffineUncompressed(), kryptologyP1.ToAffineUncompressed())
	assert.True(t, bytes.Equal(customP3.ToAffineUncompressed(), kryptologyP1.ToAffineUncompressed()))
	fmt.Printf("CustomP3 Compressed Serialized:\t\t%x\nOriginal Point Serialized:\t\t\t%x\n", customP3.ToAffineCompressed(), kryptologyP1.ToAffineCompressed())
	assert.True(t, bytes.Equal(customP3.ToAffineCompressed(), kryptologyP1.ToAffineCompressed()))
}