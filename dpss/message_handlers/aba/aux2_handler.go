package aba

import (
	"math"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/old_committee"
)

var Aux2MessageType string = "aba_aux2"

type Aux2Message struct {
	RoundID common.PSSRoundDetails
	Kind    string
	Curve   common.CurveName
	V       int
	R       int
}

func NewAux2Message(id common.PSSRoundDetails, v, r int, curve common.CurveName) (*common.PSSMessage, error) {
	m := Aux2Message{
		id,
		Aux2MessageType,
		curve,
		v,
		r,
	}
	bytes, err := bijson.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreatePSSMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m Aux2Message) Process(sender common.NodeDetails, self common.PSSParticipant) {
	v, r := m.V, m.R

	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.RoundID.ToRoundID(), common.DefaultABAStore())
	if complete {
		log.Infof("Keygen already complete: %v", m.RoundID)
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
				coinID := string(m.RoundID.ToRoundID()) + strconv.Itoa(m.R)
				log.Debugf("Node=%d, Round=%v,CoinID=%s", self.Details().Index, m.RoundID, coinID)
				msg, err := NewCoinInitMessage(m.RoundID, coinID, m.Curve)
				if err != nil {
					return
				}
				go self.ReceiveMessage(self.Details(), *msg)
			} else {
				pssID := m.RoundID.PssID
				pssState, complete := self.State().PSSStore.GetOrSetIfNotComplete(pssID)
				if complete {
					log.Infof("Keygen already complete: %s", pssID)
					return
				}

				pssState.Lock()
				defer pssState.Unlock()
				log.WithFields(log.Fields{
					"completed":     pssState.ABAComplete,
					"round":         m.RoundID,
					"self":          self.Details().Index,
					"Decisions":     pssState.Decisions,
					"ABAStarted":    pssState.ABAStarted,
					"completeCount": len(pssState.Decisions),
				}).Debug("pssState:Aux2Handler")

				log.WithFields(log.Fields{
					"Decisions": pssState.Decisions,
				}).Debugf("Node %d decided on round %v=%d", self.Details().Index, m.RoundID, w)
				leader := m.RoundID.Dealer.Index
				// Set decision to 0 or 1
				if _, ok := pssState.Decisions[leader]; ok {
					return
				}
				pssState.Decisions[leader] = w

				// If one ABA has outputted 1, then any ABA hasn't started yet, vote 0 for that ABA
				if w == 1 && !pssState.ABAComplete {
					pssState.ABAComplete = true
					pssID := m.RoundID.PssID
					log.Debugf("PSSID=%s, decisions=%v,self=%d", pssID, pssState.Decisions, self.Details().Index)

					for i := 1; i <= n; i++ {
						if !Contains(pssState.ABAStarted, i) {
							details, err := self.OldNodeDetailsByID(i)
							if err != nil {
								continue
							}
							round := common.CreatePSSRound(pssID, details, m.RoundID.BatchSize)
							msg, err := NewInitMessage(round, 0, 0, m.Curve)
							if err != nil {
								log.WithError(err).Error("Could not create init message")
								return
							}
							go self.ReceiveMessage(self.Details(), *msg)
						}
					}
				}

				// If all rounds ABA'd to 0 or 1, set ABA complete to true and send init HIM
				if n == len(pssState.Decisions) && !pssState.HIMStarted {
					log.WithFields(log.Fields{
						"roundID":   m.RoundID,
						"f":         f,
						"node":      self.Details().Index,
						"Decisions": pssState.Decisions,
						"T":         pssState.T,
					}).Debug("starting HIM")

					ch := pssState.WaitForTSet(n, f)
					pssState.Unlock()
					T := <-ch
					pssState.Lock()
					curve := common.CurveFromName(m.Curve)
					numShares := m.RoundID.BatchSize

					alpha := int(math.Ceil(float64(numShares) / float64((n - 2*f))))
					shares, err := pssState.GetSharesFromT(T, alpha, curve)
					if err != nil {
						log.Errorf("Error: AUX2: GetShares: %s", err)
						return
					}
					msg, err := old_committee.NewDpssHimMessage(m.RoundID, shares, []byte{}, m.Curve)
					if err != nil {
						return
					}
					pssState.HIMStarted = true
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
