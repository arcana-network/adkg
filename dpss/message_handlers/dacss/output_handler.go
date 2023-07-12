package dacss

import (
	"encoding/binary"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/arcana-network/dkgnode/common"
	dpsscommon "github.com/arcana-network/dkgnode/dpss/common"
	"github.com/arcana-network/dkgnode/dpss/message_handlers/keyset"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/messages"

	"github.com/coinbase/kryptology/pkg/core/curves"
	log "github.com/sirupsen/logrus"
)

var OutputMessageType common.DPSSMessageType = "dacss_output"

type OutputMessage struct {
	RoundID      common.DPSSRoundID
	kind         common.DPSSMessageType
	curve        *curves.Curve
	m            []byte
	newCommittee bool
	handlerType  string
}

func NewOutputMessage(id common.DPSSRoundID, data []byte, curve *curves.Curve, handlerType string, newCommittee bool) (*common.DPSSMessage, error) {
	m := OutputMessage{
		RoundID:      id,
		newCommittee: newCommittee,
		kind:         OutputMessageType,
		curve:        curve,
		m:            data,
		handlerType:  handlerType,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateDPSSMessage(m.RoundID, m.kind, bytes)
	return &msg, nil
}

func (m OutputMessage) Process(sender common.KeygenNodeDetails, p dpsscommon.DPSSParticipant) {
	log.Debugf("Received output message on %d, OVER!!!!", sender.Index)

	//extracting common adkgid for old and new committee
	commonDPSSID, err := common.DPSSIDFromRoundID(m.RoundID)
	if err != nil {
		log.Debugf("Could not get adkgid from roundID, err=%s", err)
		return
	}
	substrs := strings.Split(string(commonDPSSID), common.Delimiter3)
	id := new(big.Int)
	id.SetString(substrs[1], 16)

	commonDPSSID = common.GenerateDPSSID(*id)

	if m.handlerType == "ready" {
		n, k, _ := p.Params(m.newCommittee)

		// create default session to use below
		s, err := dpsscommon.GetSessionStoreFromRoundID(m.RoundID, p)
		if err != nil {
			log.Debugf("Could not get session store from roundID, err=%s", err)
			return
		}
		s.Lock()
		defer s.Unlock()
		dealer, err := m.RoundID.Leader()
		if err != nil {
			log.Debugf("Could not get leader from roundID, err=%s", err)
			return
		}

		// Check if this round completeness has been accounted in current adkg id
		if !hasBit(s.TPrime, int(dealer.Int64())) {
			log.Debugf("Inside !hasBit, roundID=%s", m.RoundID)
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

			log.Debugf("output: round=%s, shared_key=%v, node=%d, cipher=%v", m.RoundID, key[:], p.ID(), msg.ShareMap[uint32(p.ID())])
			share, verifier, verified := acss.Predicate(key[:], msg.ShareMap[uint32(p.ID())], msg.Commitments, k, m.curve)

			if verified {
				log.Debugf("acss_verified: share=%v", *share)
				s.S[int(dealer.Int64())] = *share
				s.TPrime = setBit(s.TPrime, int(dealer.Int64()))
				s.C[int(dealer.Int64())] = verifier.Commitments

				// Check proposals and emit
				for key, v := range s.TProposals {
					if keyset.Predicate(dpsscommon.IntToByteValue(s.TPrime), dpsscommon.IntToByteValue(v)) {
						roundID := common.CreateDPSSRound(commonDPSSID, key, "keyset")
						keyset.OnKeysetVerified(roundID, *m.curve, dpsscommon.IntToByteValue(v), s, key, p)
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

			dpssID, err := common.DPSSIDFromRoundID(m.RoundID)
			if err != nil {
				log.Debug("Could not get ADKGIDf from roundID")
				return
			}

			//Store adkgid and T
			//accessing the  commitment store and storing T set,adkgid
			commitmentStore := p.State().OutputCommitmentStore
			defaultCfc := dpsscommon.CommitmentsForCommittees{
				Ti:    0,
				TjOld: 0,
				TjNew: 0,
			}
			commitstate, found := commitmentStore.GetOrSet(m.RoundID, &defaultCfc)
			commitstate.Lock()
			defer commitstate.Unlock()
			commitstate.DPSSID = dpssID
			commitstate.Ti = s.TPrime

			if (commitstate.TjOld&commitstate.Ti == commitstate.Ti) && found {
				msg, err := NewmultiAcssMessage(m.RoundID, commitstate.Ti, m.curve)
				if err != nil {
					return
				}
				go p.ReceiveMessage(*msg)
			}
			return

		}
		return

		//Only old committee perform the keyset proposal phase
	} else if m.handlerType == "commit" {

		// accessing the  commitment store to get  T set,adkgid
		commitmentStore := p.State().OutputCommitmentStore
		defaultCfc := dpsscommon.CommitmentsForCommittees{
			Ti:    0,
			TjOld: 0,
			TjNew: 0,
		}
		commitstate, _ := commitmentStore.GetOrSet(m.RoundID, &defaultCfc)
		commitstate.Lock()
		defer commitstate.Unlock()

		//Keeping track of opposite acss and making sure they match with the acss in Ti
		dealer, err := m.RoundID.Leader()
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
				s, err := dpsscommon.GetSessionStoreFromRoundID(m.RoundID, p)
				if err != nil {
					log.Debugf("Could not get session store from roundID, err=%s", err)
					return
				}
				s.Lock()
				defer s.Unlock()
				dealer, err := m.RoundID.Leader()
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
		if (commitstate.TjOld&commitstate.Ti == commitstate.Ti) && m.newCommittee && commitstate.DPSSID != common.DPSSID("") {
			msg, err := NewmultiAcssMessage(m.RoundID, commitstate.Ti, m.curve)
			if err != nil {
				return
			}
			go p.ReceiveMessage(*msg)
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
