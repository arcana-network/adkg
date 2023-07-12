package aba

import (
	"encoding/json"
	"strconv"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/him"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var Aux2MessageType common.DPSSMessageType = "aba_aux2"

type Aux2Message struct {
	roundID common.DPSSRoundID
	kind    common.DPSSMessageType
	curve   *curves.Curve
	v       int
	r       int
}

func NewAux2Message(id common.DPSSRoundID, v, r int, curve *curves.Curve, sender int) (*common.DPSSMessage, error) {
	m := Aux2Message{
		id,
		Aux2MessageType,
		curve,
		v,
		r,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.roundID, m.kind, bytes)
	return &msg, nil
}

func (m *Aux2Message) Process(sender common.KeygenNodeDetails, self dpsscommon.DPSSParticipant) {
	v, r := m.v, m.r

	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.roundID, common.DefaultABAStore())
	if complete {
		log.Debugf("Keygen already complete: %s", m.roundID)
		return
	}
	store.Lock()
	defer store.Unlock()

	// If round has incremented ignore previous round messages
	if store.Round() != r {
		log.WithFields(log.Fields{
			"Got":      r,
			"Expected": store.Round(),
		}).Debugf("Process():%s", Aux2MessageType)
		return
	}
	n, _, f := self.Params(false)

	// Check if already present
	if Contains(store.Values("aux2", r, v), sender.Index) {
		log.Debugf("Got redundant AUX message from %d", sender.Index)
		return
	}

	//Otherwise, add sender
	store.SetValues("aux2", r, v, sender.Index)

	aux2Len0 := len(store.Values("aux2", r, 0))
	aux2Len1 := len(store.Values("aux2", r, 1))
	aux2Len2 := len(store.Values("aux2", r, 2))
	bin2 := store.Bin("bin2", r)

	log.WithFields(log.Fields{
		"round":    m.roundID,
		"aux2Len0": aux2Len0,
		"aux2-0":   store.Values("aux2", r, 0),
		"aux2Len1": aux2Len1,
		"aux2-1":   store.Values("aux2", r, 1),
		"aux2Len2": aux2Len2,
		"aux2-2":   store.Values("aux2", r, 2),
		"bin2":     bin2,
	}).Debugf("Process():%s", Aux2MessageType)

	var values2 []int

	if Contains(bin2, 1) && aux2Len1 >= n-f {
		values2 = append(values2, 1)
	} else if Contains(bin2, 0) && aux2Len0 >= n-f {
		values2 = append(values2, 0)
	} else if Contains(bin2, 2) && aux2Len2 >= n-f {
		values2 = append(values2, 2)
	} else if (aux2Len2+aux2Len0) >= n-f && Contains(bin2, 0) && Contains(bin2, 2) {
		values2 = append(values2, 0, 2)
	} else if (aux2Len2+aux2Len1) >= n-f && Contains(bin2, 1) && Contains(bin2, 2) {
		values2 = append(values2, 1, 2)
	}
	log.WithFields(log.Fields{
		"values2": values2,
		"round":   m.roundID,
	}).Debugf("Process():%s", Aux2MessageType)
	if len(values2) > 0 {
		if len(values2) == 1 {
			w := values2[0]
			if w == 2 {
				// Create a coin
				coinID := string(m.roundID) + strconv.Itoa(m.r)
				log.Debugf("PSSNode=%d, Round=%s,CoinID=%s", self.ID(), m.roundID, coinID)
				msg, err := NewCoinInitMessage(m.roundID, coinID, m.curve, self.ID())
				if err != nil {
					return
				}
				go self.ReceiveMessage(*msg)
			} else {
				round := common.DPSSRoundDetails{}
				err := round.FromID(m.roundID)
				if err != nil {
					log.Debugf("Could not get leader from roundID, err=%s", err)
					return
				}
				sessionStore, complete := self.State().SessionStore.GetOrSetIfNotComplete(round.DPSSID, dpsscommon.DefaultDPSSSession())
				if complete {
					log.Debugf("Keygen already complete: %s", round.DPSSID)
					return
				}

				sessionStore.Lock()
				defer sessionStore.Unlock()
				log.WithFields(log.Fields{
					"completed":     sessionStore.ABAComplete,
					"round":         m.roundID,
					"self":          self.ID(),
					"Decisions":     sessionStore.Decisions,
					"ABAStarted":    sessionStore.ABAStarted,
					"completeCount": len(sessionStore.Decisions),
				}).Debug("SessionStore:Aux2Handler")

				log.WithFields(log.Fields{
					"Decisions": sessionStore.Decisions,
				}).Debugf("PSSNode %d decided on round %s=%d", self.ID(), m.roundID, w)
				// Set decision to 0 or 1
				if _, ok := sessionStore.Decisions[round.Dealer]; ok {
					return
				}
				sessionStore.Decisions[round.Dealer] = w

				// If one ABA has outputted 1, then any ABA hasn't started yet, vote 0 for that ABA
				if w == 1 && !sessionStore.ABAComplete {
					sessionStore.ABAComplete = true
					dpssID, err := common.DPSSIDFromRoundID(m.roundID)
					if err != nil {
						log.Debug("Could not get ADKGIDf from roundID")
						return
					}
					index, _ := dpssID.GetIndex()
					log.Debugf("ADKGID=%d, decisions=%v,self=%d", index.Int64(), sessionStore.Decisions, self.ID())

					for i := 1; i <= n; i++ {
						if !Contains(sessionStore.ABAStarted, i) {
							// go func(id int) {
							round := common.CreateDPSSRound(dpssID, i, "keyset")
							msg, err := NewInitMessage(round, 0, 0, m.curve)
							if err != nil {
								return
							}
							go self.ReceiveMessage(*msg)
							// }(i)
						}
					}
				}

				// If all rounds ABA'd to 0 or 1, set ABA complete to true and start key derivation
				if n == len(sessionStore.Decisions) && !sessionStore.KeyderivationStarted {
					log.WithFields(log.Fields{
						"roundID":   m.roundID,
						"f":         f,
						"node":      self.ID(),
						"Decisions": sessionStore.Decisions,
					}).Debug("starting_key_derivation")
					msg, err := him.NewInitMessage(round.ID(), false, m.curve)
					if err != nil {
						return
					}
					go self.ReceiveMessage(*msg)

				}
			}
		} else {
			w := values2[0]
			store.IncrementRound()
			msg, err := NewInitMessage(m.roundID, w, store.Round(), m.curve)
			if err != nil {
				return
			}
			go self.ReceiveMessage(*msg)
		}
	}
}

func GetDecisionsCount(m map[int]int, vote int) int {
	count := 0
	for _, v := range m {
		if v == vote {
			count += 1
		}
	}
	return count
}
