package dacss

import (
	"encoding/binary"
	"math/big"
	"strings"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/adkg-proto/common"
	"github.com/arcana-network/adkg-proto/common/acss"
	"github.com/arcana-network/adkg-proto/messages"
	"github.com/arcana-network/adkg-proto/messages/keyset"
)

var AcssOutputMessageType common.MessageType = "dacss_output"

type AcssOutputMessage struct {
	roundID      common.RoundID
	sender       int
	kind         common.MessageType
	curve        *curves.Curve
	m            []byte
	newCommittee bool
	handlerType  string
}

func NewAcssOutputMessage(id common.RoundID, data []byte, curve *curves.Curve, sender int, handlerType string, newCommittee bool) common.DKGMessage {
	m := AcssOutputMessage{
		roundID:      id,
		sender:       sender,
		newCommittee: newCommittee,
		kind:         AcssOutputMessageType,
		curve:        curve,
		m:            data,
		handlerType:  handlerType,
	}

	return m
}

func (m AcssOutputMessage) Sender() int {
	return m.sender
}

func (m AcssOutputMessage) Kind() common.MessageType {
	return m.kind
}

func (m AcssOutputMessage) Process(p common.DkgParticipant) {
	log.Debugf("Received output message on %d, OVER!!!!", m.Sender())

	//extracting common adkgid for old and new committee
	commonadkgID, err := common.ADKGIDFromRoundID(m.roundID)
	if err != nil {
		log.Debugf("Could not get adkgid from roundID, err=%s", err)
		return
	}
	substrs := strings.Split(string(commonadkgID), common.Delimiter3)
	adkgIDBig := new(big.Int)
	adkgIDBig.SetString(substrs[1], 16)

	commonadkgID = common.GenerateADKGID(*adkgIDBig)

	if m.handlerType == "ready" {
		n, k, _ := p.Params(m.newCommittee)

		// create default session to use below
		s, err := common.GetSessionStoreFromRoundID(m.roundID, p)
		if err != nil {
			log.Debugf("Could not get session store from roundID, err=%s", err)
			return
		}
		s.Lock()
		defer s.Unlock()
		dealer, err := m.roundID.Leader()
		if err != nil {
			log.Debugf("Could not get leader from roundID, err=%s", err)
			return
		}

		// Check if this round completeness has been accounted in current adkg id
		if !hasBit(s.TPrime, int(dealer.Int64())) {
			log.Debugf("Inside !hasBit, roundID=%s", m.roundID)
			// Get dealer public key from round id
			dealerPublicKey := p.PublicKey(int(dealer.Int64()))

			// Recover shared symmetric key
			priv := p.SelfPrivateKey()
			key := acss.SharedKey(priv, dealerPublicKey)

			// Predicate from recovered output data
			msg := messages.MessageData{}

			err = msg.Deserialize(m.m)
			if err != nil {
				log.Debugf("Could not deserialize message data, err=%s", err)
				return
			}

			log.Debugf("output: round=%s, shared_key=%v, node=%d, cipher=%v", m.roundID, key[:], p.ID(), msg.ShareMap[uint32(p.ID())])
			share, verifier, verified := acss.Predicate(key[:], msg.ShareMap[uint32(p.ID())], msg.Commitments, k, m.curve)

			if verified {
				log.Debugf("acss_verified: share=%v", *share)
				s.S[int(dealer.Int64())] = *share
				s.TPrime = setBit(s.TPrime, int(dealer.Int64()))
				s.C[int(dealer.Int64())] = verifier.Commitments

				// Check proposals and emit
				for key, v := range s.TProposals {
					if keyset.Predicate(common.IntToByteValue(s.TPrime), common.IntToByteValue(v)) {
						roundID := common.CreateRound(commonadkgID, key, "keyset")
						keyset.OnKeysetVerified(roundID, m.curve, common.IntToByteValue(v), s, key, p)
						delete(s.TProposals, k)
					}
				}
			}
		}

		//Need to collect n-t honest parties and
		//Need to initiate keyset proposal only for old committee

		if countSetBits(s.TPrime) >= (n-k+1) && !m.newCommittee {
			var output [8]byte
			binary.BigEndian.PutUint64(output[:], uint64(s.TPrime))
			// Store output to use in predicate for other keyset proposals

			adkgid, err := common.ADKGIDFromRoundID(m.roundID)
			if err != nil {
				log.Debug("Could not get ADKGIDf from roundID")
				return
			}

			//Store adkgid and T
			//accessing the  commitment store and storing T set,adkgid
			commitmentStore := p.State().OutputCommitmentStore
			defaultCfc := common.CommitmentsForCommittees{
				Ti:    0,
				TjOld: 0,
				TjNew: 0,
			}
			commitstate, found := commitmentStore.GetOrSet(commonadkgID, &defaultCfc)
			commitstate.Lock()
			defer commitstate.Unlock()
			commitstate.AdkgID = adkgid
			commitstate.Ti = s.TPrime

			if (commitstate.TjOld&commitstate.Ti == commitstate.Ti) && found {
				msg := NewmultiAcssMessage(adkgid, commitstate.Ti, m.curve, p.ID())
				go p.ReceiveMessage(msg)
			}
			return

		}
		return

		//Only old committee perform the keyset proposal phase
	} else if m.handlerType == "commit" {

		// accessing the  commitment store to get  T set,adkgid
		commitmentStore := p.State().OutputCommitmentStore
		defaultCfc := common.CommitmentsForCommittees{
			Ti:    0,
			TjOld: 0,
			TjNew: 0,
		}
		commitstate, _ := commitmentStore.GetOrSet(commonadkgID, &defaultCfc)
		commitstate.Lock()
		defer commitstate.Unlock()

		//Keeping track of opposite acss and making sure they match with the acss in Ti
		dealer, err := m.roundID.Leader()
		if err != nil {
			log.Debugf("Could not get leader from roundID, err=%s", err)
			return
		}

		if m.newCommittee {
			commitstate.TjOld = setBit(commitstate.TjOld, int(dealer.Int64()))

		} else {
			//Storing commitment from old committee in new nodes
			if p.ID() > len(p.Nodes(false)) {
				// create default session to use below
				s, err := common.GetSessionStoreFromRoundID(m.roundID, p)
				if err != nil {
					log.Debugf("Could not get session store from roundID, err=%s", err)
					return
				}
				s.Lock()
				defer s.Unlock()
				dealer, err := m.roundID.Leader()
				if err != nil {
					log.Debugf("Could not get leader from roundID, err=%s", err)
					return
				}
				_, k, _ := p.Params(m.newCommittee)
				commitment, _ := acss.DecompressCommitments(k, m.m, m.curve)
				s.C[int(dealer.Int64())] = commitment

			}
			commitstate.TjNew = setBit(commitstate.TjNew, int(dealer.Int64()))

		}

		// the ACSSes that ended in the current committee must be a subset of the ACSSes
		//that ended in the opp committee
		if (commitstate.TjOld&commitstate.Ti == commitstate.Ti) && m.newCommittee && commitstate.AdkgID != common.ADKGID("") {
			adkgid := commitstate.AdkgID
			msg := NewmultiAcssMessage(adkgid, commitstate.Ti, m.curve, p.ID())
			go p.ReceiveMessage(msg)
		}

	}

}

func hasBit(n int, pos int) bool {
	val := n & (1 << pos)
	return (val > 0)
}
func setBit(n int, pos int) int {
	n |= (1 << pos)
	return n
}

func countSetBits(n int) int {
	count := 0
	for n > 0 {
		n &= (n - 1)
		count++
	}
	return count
}
