package common

import (
	"strconv"
	"strings"
)

type NodeNetwork struct {
	Nodes map[NodeDetailsID]KeygenNodeDetails
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
)

type KeygenMessageVersion string

type DKGMessage struct {
	Version KeygenMessageVersion `json:"version,omitempty"`
	RoundID RoundID              `json:"round_id"`
	Method  string               `json:"type"`
	Data    []byte               `json:"data"`
}

type NodeDetailsID string

const NullNodeDetails = NodeDetailsID("")

func (n *KeygenNodeDetails) ToNodeDetailsID() NodeDetailsID {
	return NodeDetailsID(strings.Join([]string{
		strconv.Itoa(n.Index),
		n.PubKey.X.Text(16),
		n.PubKey.Y.Text(16),
	}, Delimiter1))
}
