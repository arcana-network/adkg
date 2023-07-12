package common

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

func GenerateDPSSID(rindex, noOfRandoms big.Int) DPSSID {
	index := strings.Join([]string{rindex.Text(16), noOfRandoms.Text(16)}, Delimiter2)
	return DPSSID(strings.Join([]string{"DPSS", index}, Delimiter3))
}

func (id *DPSSID) GetIndex() (big.Int, error) {
	str := string(*id)
	substrs := strings.Split(str, Delimiter3)

	if len(substrs) != 2 {
		return *new(big.Int), errors.New("could not parse dkgid")
	}

	index, ok := new(big.Int).SetString(substrs[1], 16)
	if !ok {
		return *new(big.Int), errors.New("could not get back index from dkgid")
	}

	return *index, nil
}

func DPSSIDFromRoundID(r DPSSRoundID) (DPSSID, error) {
	d := DPSSRoundDetails{}
	err := d.FromID(r)
	if err != nil {
		return DPSSID(""), err
	}

	return d.DPSSID, nil
}

func CreateDPSSRound(id DPSSID, dealer int, kind string) DPSSRoundID {
	r := DPSSRoundDetails{
		id,
		dealer,
		kind,
	}
	return r.ID()
}

func (d *DPSSRoundDetails) ID() DPSSRoundID {
	return DPSSRoundID(strings.Join([]string{string(d.DPSSID), d.Kind, strconv.Itoa(d.Dealer)}, Delimiter4))
}

func (d *DPSSRoundDetails) FromID(roundID DPSSRoundID) error {
	s := string(roundID)
	substrings := strings.Split(s, Delimiter4)

	if len(substrings) != 3 {
		return fmt.Errorf("expected length of 2, got=%d", len(substrings))
	}
	d.DPSSID = DPSSID(substrings[0])
	d.Kind = substrings[1]
	index, err := strconv.Atoi(substrings[2])
	if err != nil {
		return err
	}
	d.Dealer = index
	return nil
}

func (r *DPSSRoundID) Leader() (big.Int, error) {
	str := string(*r)
	substrs := strings.Split(str, Delimiter4)

	if len(substrs) != 3 {
		return *new(big.Int), errors.New("could not parse round id")
	}

	index, ok := new(big.Int).SetString(substrs[2], 16)
	if !ok {
		return *new(big.Int), errors.New("could not get back index from round id")
	}

	return *index, nil
}
