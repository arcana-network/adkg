package aba

import (
	"encoding/json"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/keyderivation"
)

var Aux2MessageType string = "aba_aux2"

type Aux2Message struct {
	RoundID common.RoundID
	Kind    string
	Curve   common.CurveName
	V       int
	R       int
}

func NewAux2Message(id common.RoundID, v, r int, curve common.CurveName) (*common.DKGMessage, error) {
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

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m Aux2Message) Process(sender common.NodeDetails, self common.DkgParticipant) {
	v, r := m.V, m.R

	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.RoundID, common.DefaultABAStore())
	if complete {
		log.Infof("Keygen already complete: %s", m.RoundID)
		return
	}
	store.Lock()
	defer store.Unlock()

	n, _, f := self.Params()

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
	bin2 := store.GetBin("bin2", r)

	log.WithFields(log.Fields{
		"round":    m.RoundID,
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
		"round":   m.RoundID,
	}).Debugf("Process():%s", Aux2MessageType)
	if len(values2) > 0 {
		if len(values2) == 1 {
			w := values2[0]
			if w == 2 {
				// Create a coin
				coinID := string(m.RoundID) + strconv.Itoa(m.R)
				log.Debugf("Node=%d, Round=%s,CoinID=%s", self.ID(), m.RoundID, coinID)
				msg, err := NewCoinInitMessage(m.RoundID, coinID, m.Curve)
				if err != nil {
					return
				}
				go self.ReceiveMessage(self.Details(), *msg)
			} else {
				round := common.RoundDetails{}
				err := round.FromID(m.RoundID)
				if err != nil {
					log.Infof("Could not get leader from roundID, err=%s", err)
					return
				}
				sessionStore, complete := self.State().SessionStore.GetOrSetIfNotComplete(round.ADKGID, common.DefaultADKGSession())
				if complete {
					log.Infof("Keygen already complete: %s", round.ADKGID)
					return
				}

				sessionStore.Lock()
				defer sessionStore.Unlock()
				log.WithFields(log.Fields{
					"completed":     sessionStore.ABAComplete,
					"round":         m.RoundID,
					"self":          self.ID(),
					"Decisions":     sessionStore.Decisions,
					"ABAStarted":    sessionStore.ABAStarted,
					"completeCount": len(sessionStore.Decisions),
				}).Debug("SessionStore:Aux2Handler")

				log.WithFields(log.Fields{
					"Decisions": sessionStore.Decisions,
				}).Debugf("Node %d decided on round %s=%d", self.ID(), m.RoundID, w)
				// Set decision to 0 or 1
				if _, ok := sessionStore.Decisions[round.Dealer]; ok {
					return
				}
				sessionStore.Decisions[round.Dealer] = w

				// If one ABA has outputted 1, then any ABA hasn't started yet, vote 0 for that ABA
				if w == 1 && !sessionStore.ABAComplete {
					sessionStore.ABAComplete = true
					adkgid, err := common.ADKGIDFromRoundID(m.RoundID)
					if err != nil {
						log.Debug("Could not get ADKGIDf from roundID")
						return
					}
					index, _ := adkgid.GetIndex()
					log.Debugf("ADKGID=%d, decisions=%v,self=%d", index.Int64(), sessionStore.Decisions, self.ID())

					for i := 1; i <= n; i++ {
						if !Contains(sessionStore.ABAStarted, i) {
							// go func(id int) {
							round := common.CreateRound(adkgid, i, "keyset")
							msg, err := NewInitMessage(round, 0, 0, m.Curve)
							if err != nil {
								log.WithError(err).Error("Could not create init message")
								return
							}
							go self.ReceiveMessage(self.Details(), *msg)
							// }(i)
						}
					}
				}

				// If all rounds ABA'd to 0 or 1, set ABA complete to true and start key derivation
				if n == len(sessionStore.Decisions) && !sessionStore.KeyderivationStarted {
					log.WithFields(log.Fields{
						"roundID":   m.RoundID,
						"f":         f,
						"node":      self.ID(),
						"Decisions": sessionStore.Decisions,
					}).Debug("starting_key_derivation")
					msg, err := keyderivation.NewInitMessage(m.RoundID, m.Curve)
					if err != nil {
						return
					}
					sessionStore.KeyderivationStarted = true
					go self.ReceiveMessage(self.Details(), *msg)
				}
			}
		} else {
			log.Debugf("Round::Current: %d, Next: %d", store.GetRound(), store.GetRound()+1)
			w := values2[0]
			msg, err := NewInitMessage(m.RoundID, w, store.GetRound()+1, m.Curve)
			if err != nil {
				return
			}
			go self.ReceiveMessage(self.Details(), *msg)
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
