package common

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"strings"
	"sync"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var himCache [][]curves.Scalar
var himMaxSize int
var himCacheLock sync.Mutex

// Vandermonde matrix
func CreateHIM(size int, curve *curves.Curve) [][]curves.Scalar {
	himCacheLock.Lock()
	if himMaxSize >= size {
		himCacheLock.Unlock()
		return copyHIM(himCache)
	}
	himCacheLock.Unlock()

	him := make([][]curves.Scalar, size)

	for i := 1; i <= size; i++ {
		him[i-1] = make([]curves.Scalar, size)
		him[i-1][0] = curve.Scalar.New(1)
		element := curve.Scalar.New(i)

		for j := 1; j < size; j++ {

			prev := him[i-1][j-1]
			him[i-1][j] = prev.Mul(element)
		}
	}

	himCacheLock.Lock()
	himMaxSize = size
	himCache = copyHIM(him)
	himCacheLock.Unlock()

	return him
}

func copyHIM(him [][]curves.Scalar) [][]curves.Scalar {
	himCopy := make([][]curves.Scalar, len(him))
	for i, subSlice := range him {
		var oneSlice = make([]curves.Scalar, len(subSlice))
		copy(oneSlice, subSlice)
		himCopy[i] = oneSlice
	}
	return himCopy
}

func DotProduct(first, second []curves.Scalar, curve *curves.Curve) curves.Scalar {
	product := curve.Scalar.Zero()

	for i := range first {
		product = first[i].MulAdd(second[i], product)
	}

	return product
}

func GetSessionStoreFromRoundID(roundID common.DPSSRoundID, p DPSSParticipant) (*DPSSSession, error) {
	r := &common.DPSSRoundDetails{}
	err := r.FromID(roundID)
	if err != nil {
		log.Debugf("Error parsing round id, err=%s", err)
		return nil, err
	}
	sessionStore := p.State().SessionStore
	// create default session to use below
	s, _ := sessionStore.GetOrSet(r.DPSSID, DefaultDPSSSession())
	return s, nil
}

func GenerateDPSSID(rindex, noOfRandoms big.Int) common.DPSSID {
	index := strings.Join([]string{rindex.Text(16), noOfRandoms.Text(16)}, common.Delimiter2)
	return common.DPSSID(strings.Join([]string{"DPSS", index}, common.Delimiter3))
}

func Hash(msg []byte) []byte {
	sum := sha256.Sum256(msg)
	return sum[:]
}

func IntToByteValue(val int) []byte {
	var byteVal [8]byte
	binary.BigEndian.PutUint64(byteVal[:], uint64(val))
	return byteVal[:]
}

type CommitmentsForCommittees struct {
	sync.Mutex
	Ti     int //For counting the acss
	TjOld  int //For counting the opposite acss when in the old committee
	TjNew  int //For counting the opposite acss when in the new committee
	DPSSID common.DPSSID
}

type DacssState struct {
	sync.Mutex
	Ended   bool
	T_dacss map[common.DPSSID]int
}
