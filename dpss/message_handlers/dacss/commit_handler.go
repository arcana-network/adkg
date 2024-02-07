package dacss

import (
	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/adkg-proto/common"
)

var AcssCommitMessageType common.MessageType = "dacss_commit"

type AcssCommitMessage struct {
	roundID       common.RoundID
	sender        int
	committeeType int
	kind          common.MessageType
	curve         *curves.Curve
	Commitment    []byte
	newCommittee  bool
}

func NewAcssCommitMessage(id common.RoundID, data []byte, curve *curves.Curve, sender int, newCommittee bool) common.DKGMessage {
	m := AcssCommitMessage{
		roundID:      id,
		sender:       sender,
		curve:        curve,
		Commitment:   data,
		newCommittee: newCommittee,
	}
	if newCommittee {
		m.committeeType = 1
	} else {
		m.committeeType = 0
	}
	return m
}

func (m AcssCommitMessage) Sender() int {
	return m.sender
}

func (m AcssCommitMessage) Kind() common.MessageType {
	return m.kind
}

func (m AcssCommitMessage) Process(p common.DkgParticipant) {

	_, _, t := p.Params(m.newCommittee)

	commitmentStore := p.State().CommitmentStore
	adkgID, err := common.ADKGIDFromRoundID(m.roundID)
	if err != nil {
		log.Errorf("Error getting ADKG ID: %s", err)
		return
	}

	defaultCommitCount := common.CommitmentState{
		CommitmentsForADKGID: make(map[string]int),
		ReceivedCommit:       make(map[string][]int),
		Ended:                make(map[string]bool),
	}
	commitmentCount, _ := commitmentStore.GetOrSet(adkgID, &defaultCommitCount)
	commitmentCount.Lock()
	defer commitmentCount.Unlock()

	// Check if it has already been received
	received, found := commitmentCount.ReceivedCommit[string(m.Commitment)]
	if contains(received, m.sender) && found {

		log.Debugf("Already received share for %s from %d", m.roundID, m.Sender())
		return

	}

	commitmentCount.ReceivedCommit[string(m.Commitment)] = append(commitmentCount.ReceivedCommit[string(m.Commitment)], m.Sender())
	commitmentCount.CommitmentsForADKGID[string(m.Commitment)]++

	if commitmentCount.CommitmentsForADKGID[string(m.Commitment)] >= t+1 && !commitmentCount.Ended[string(m.Commitment)] {
		commitmentCount.Ended[string(m.Commitment)] = true
		outputMsg := NewAcssOutputMessage(m.roundID, m.Commitment, m.curve, p.ID(), "commit", m.newCommittee)
		go p.ReceiveMessage(outputMsg)
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
