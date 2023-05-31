package acss

import (
	"encoding/json"

	"github.com/arcana-network/dkgnode/common"
	kcommon "github.com/arcana-network/dkgnode/keygen/common"
	"github.com/arcana-network/dkgnode/keygen/common/acss"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/keyset"
	"github.com/arcana-network/dkgnode/keygen/messages"

	log "github.com/sirupsen/logrus"
)

var OutputMessageType string = "acss_output"

type OutputMessage struct {
	RoundID common.RoundID
	Kind    string
	Curve   common.CurveName
	M       []byte
}

func NewOutputMessage(id common.RoundID, data []byte, curve common.CurveName) (*common.DKGMessage, error) {
	m := OutputMessage{
		id,
		OutputMessageType,
		curve,
		data,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m OutputMessage) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {
	log.Debugf("Received output message on %d", self.ID())
	// Ignore if not received by self
	if sender.Index != self.ID() {
		return
	}

	adkgid, err := common.ADKGIDFromRoundID(m.RoundID)
	if err != nil {
		log.Errorf("Error parsing round id, err=%s", err)
		return
	}
	// create default session to use below
	sessionStore, complete := self.State().SessionStore.GetOrSetIfNotComplete(adkgid, common.DefaultADKGSession())
	if complete {
		log.Infof("Keygen already complete: %s", adkgid)
		return
	}

	sessionStore.Lock()
	defer sessionStore.Unlock()

	log.Debugf("acss_output: round=%v, self=%v", m.RoundID, self.ID())

	dealer, err := m.RoundID.Leader()
	if err != nil {
		log.Errorf("Could not get leader from roundID, err=%s", err)
		return
	}

	if !kcommon.HasBit(sessionStore.TPrime, int(dealer.Int64())) {
		dealerPublicKey := self.PublicKey(int(dealer.Int64()))

		// Recover shared symmetric key
		priv := self.PrivateKey()
		key := acss.SharedKey(priv, dealerPublicKey)

		// Predicate from recovered output data
		msg := messages.MessageData{}

		err = msg.Deserialize(m.M)
		if err != nil {
			log.Errorf("Could not deserialize message data, err=%s", err)
			return
		}

		_, k, _ := self.Params()

		curve := common.CurveFromName(m.Curve)
		share, verifier, verified := acss.Predicate(key[:], msg.ShareMap[uint32(self.ID())], msg.Commitments, k, curve)

		if verified {
			log.Debugf("acss_verified: share=%v", *share)
			sessionStore.S[int(dealer.Int64())] = *share
			sessionStore.TPrime = kcommon.SetBit(sessionStore.TPrime, int(dealer.Int64()))
			sessionStore.C[int(dealer.Int64())] = verifier.Commitments

			// Check proposals and emit
			for key, v := range sessionStore.TProposals {
				if keyset.Predicate(kcommon.IntToByteValue(sessionStore.TPrime), kcommon.IntToByteValue(v)) {
					roundID := common.CreateRound(adkgid, key, "keyset")
					keyset.OnKeysetVerified(roundID, m.Curve, kcommon.IntToByteValue(v), sessionStore, key, self)
					delete(sessionStore.TProposals, key)
				}
			}
		} else {
			log.Errorf("didnt pass acss_predicate")
		}

		if kcommon.CountBit(sessionStore.TPrime) == k {
			adkgid, err := common.ADKGIDFromRoundID(m.RoundID)
			if err != nil {
				log.Error("Could not get ADKGID from roundID")
				return
			}
			round := common.CreateRound(adkgid, self.ID(), "keyset")
			log.Debugf("acss_output(s.T==k): dealer=%d, round=%s", self.ID(), round)

			sessionStore.T[self.ID()] = sessionStore.TPrime
			output := kcommon.IntToByteValue(sessionStore.T[self.ID()])
			msg, err := keyset.NewInitMessage(round, output, m.Curve)
			if err != nil {
				log.Errorf("Could not create keyset_init")
				return
			}
			go self.ReceiveMessage(self.Details(), *msg)
		}
	}
}

func Contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
