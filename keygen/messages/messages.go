package messages

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type MessageData struct {
	Commitments []byte            `json:"commitments"`
	ShareMap    map[uint32][]byte `json:"share_map"`
}

func (m *MessageData) Serialize() ([]byte, error) {
	bytes, err := json.Marshal(*m)
	if err != nil {
		log.Infof("Could not marshal message data, err=%s", err)
		return nil, err
	}
	return bytes, nil
}

func (m *MessageData) Deserialize(input []byte) error {
	err := json.Unmarshal(input, m)
	return err
}
