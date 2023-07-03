package common

type DPSSID string
type DPSSRoundID string

type DPSSRoundDetails struct {
	DPSSID DPSSID
	Dealer int
	Kind   string
}

type DPSSMessageType string

type DPSSMessageVersion string

type DPSSMessageRaw struct {
	RoundID DPSSRoundID
	Method  DPSSMessageType
	Data    []byte
}

type DPSSMessage struct {
	Version DPSSMessageVersion `json:"version,omitempty"`
	PSSID   DPSSRoundID        `json:"pssid"`
	Method  DPSSMessageType    `json:"type"`
	Data    []byte             `json:"data"`
}

func createDPSSMessage(r DPSSMessageRaw) DPSSMessage {
	return DPSSMessage{
		Version: DPSSMessageVersion(VERSION),
		PSSID:   r.RoundID,
		Method:  r.Method,
		Data:    r.Data,
	}
}

func CreateDPSSMessage(id DPSSRoundID, kind DPSSMessageType, data []byte) DPSSMessage {
	return createDPSSMessage(DPSSMessageRaw{
		RoundID: id,
		Method:  kind,
		Data:    data,
	})
}
