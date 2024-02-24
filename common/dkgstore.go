package common

type NodeNetwork struct {
	Nodes map[NodeDetailsID]NodeDetails
	N     int
	T     int
	K     int
	ID    int
}

const (
	Delimiter1 = "\x1c"
	Delimiter2 = "\x1d"
	Delimiter3 = "\x1e"
	Delimiter4 = "\x1f"
	Delimiter5 = "\x1a"
)

type KeygenMessageVersion string

type DKGMessage struct {
	Version KeygenMessageVersion `json:"version,omitempty"`
	RoundID RoundID              `json:"round_id"`
	Method  string               `json:"type"`
	Data    []byte               `json:"data"`
}
