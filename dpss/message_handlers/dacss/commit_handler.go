package dacss

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var CommitMessageType common.DPSSMessageType = "dacss_commit"

type CommitMessage struct {
	RoundID       common.DPSSRoundID
	committeeType int
	kind          common.DPSSMessageType
	curve         *curves.Curve
	Commitment    []byte
	newCommittee  bool
}

func NewCommitMessage(id common.DPSSRoundID, data []byte, curve *curves.Curve, newCommittee bool) (*common.DPSSMessage, error) {
	m := CommitMessage{
		RoundID:      id,
		curve:        curve,
		Commitment:   data,
		newCommittee: newCommittee,
	}
	if newCommittee {
		m.committeeType = 1
	} else {
		m.committeeType = 0
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m CommitMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {

	_, _, t := p.Params(m.newCommittee)

	commitmentStore := p.State().CommitmentStore
	dpssID, err := common.DPSSIDFromRoundID(m.RoundID)
	if err != nil {
		log.Errorf("Error getting ADKG ID: %s", err)
		return
	}

	defaultCommitCount := dpsscommon.CommitmentState{
		CommitmentsForADKGID: make(map[string]int),
		ReceivedCommit:       make(map[string][]int),
		Ended:                make(map[string]bool),
	}
	commitmentCount, _ := commitmentStore.GetOrSet(dpssID, &defaultCommitCount)
	commitmentCount.Lock()
	defer commitmentCount.Unlock()

	// Check if it has already been received
	received, found := commitmentCount.ReceivedCommit[string(m.Commitment)]
	if contains(received, sender.Index) && found {

		log.Debugf("Already received share for %s from %d", m.RoundID, sender.Index)
		return

	}

	commitmentCount.ReceivedCommit[string(m.Commitment)] = append(commitmentCount.ReceivedCommit[string(m.Commitment)], sender.Index)
	commitmentCount.CommitmentsForADKGID[string(m.Commitment)]++

	if commitmentCount.CommitmentsForADKGID[string(m.Commitment)] >= t+1 && !commitmentCount.Ended[string(m.Commitment)] {
		commitmentCount.Ended[string(m.Commitment)] = true
		outputMsg, err := NewOutputMessage(m.RoundID, m.Commitment, m.curve, "commit", m.newCommittee)
		if err != nil {
			return
		}
		go p.ReceiveMessage(*outputMsg)
	}
}
func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
