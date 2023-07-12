package batchreconstruction

import (
	"errors"
	"math/big"
	"strings"

	"github.com/arcana-network/dkgnode/common"
)

func GetBatchSize(id *common.DPSSID) (big.Int, error) {

	str := string(*id)
	substrs1 := strings.Split(str, common.Delimiter3)
	if len(substrs1) != 3 {
		return *new(big.Int), errors.New("could not parse dkgid")
	}

	substrs2 := strings.Split(substrs1[1], common.Delimiter2)
	if len(substrs2) != 2 {
		return *new(big.Int), errors.New("could not parse dkgid")
	}

	index, ok := new(big.Int).SetString(substrs2[1], 16)
	if !ok {
		return *new(big.Int), errors.New("could not get back index from dkgid")
	}

	return *index, nil
}
