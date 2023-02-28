package common

import (
	"crypto/ecdsa"
	"math/big"
	"os"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"

	"github.com/arcana-network/dkgnode/secp256k1"
)

type Key string

const CONTEXT_KEY = Key("id")

func PointToEthAddress(p Point) ethCommon.Address {
	return ethCrypto.PubkeyToAddress(ecdsa.PublicKey{Curve: secp256k1.Curve, X: &p.X, Y: &p.Y})
}

type Point struct {
	X big.Int
	Y big.Int
}

type HexPoint struct {
	X string
	Y string
}

func (p HexPoint) ToPoint() Point {
	return Point{
		X: *secp256k1.HexToBigInt(p.X),
		Y: *secp256k1.HexToBigInt(p.Y),
	}
}

func (p Point) ToHex() HexPoint {
	return HexPoint{
		X: p.X.Text(16),
		Y: p.Y.Text(16),
	}
}

type p2pMessageVersion string

const NodeVersion = "0.1"

func CreateP2PBasicMessage(r P2PBasicMessageRaw) P2PBasicMessage {
	return P2PBasicMessage{
		Version:    p2pMessageVersion(NodeVersion),
		Timestamp:  r.Timestamp,
		Id:         r.Id,
		Gossip:     r.Gossip,
		NodeId:     r.NodeId,
		NodePubKey: r.NodePubKey,
		Sign:       r.Sign,
		MsgType:    r.MsgType,
		Payload:    r.Payload,
	}
}

type P2PBasicMessageRaw struct {
	// shared between all requests
	Timestamp  big.Int
	Id         string
	Gossip     bool
	NodeId     string
	NodePubKey []byte
	Sign       []byte
	MsgType    string
	Payload    []byte
}

type P2PBasicMsg struct {
	// shared between all requests
	Version    p2pMessageVersion `json:"version,omitempty"`
	Timestamp  big.Int           `json:"timestamp,omitempty"`  // unix time
	Id         string            `json:"id,omitempty"`         // allows requesters to use request data when processing a response
	Gossip     bool              `json:"gossip,omitempty"`     // true to have receiver peer gossip the message to neighbors
	NodeId     string            `json:"nodeId,omitempty"`     // id of node that created the message (not the peer that may have sent it). =base58(multihash(nodePubKey))
	NodePubKey []byte            `json:"nodePubKey,omitempty"` // Authoring node Secp256k1 public key (32bytes)
	Sign       []byte            `json:"sign,omitempty"`       // signature of message data + method specific data by message authoring node.
	MsgType    string            `json:"msgtype,omitempty"`    // identifyng message type
	Payload    []byte            `json:"payload"`              // payload data to be unmarshalled
}

type P2PMessage interface {
	GetTimestamp() big.Int
	GetId() string
	GetGossip() bool
	GetNodeId() string
	GetNodePubKey() []byte
	GetSign() []byte
	SetSign(sig []byte)
	GetMsgType() string
	GetPayload() []byte
	GetSerializedBody() []byte
}

func (msg *P2PBasicMessage) GetTimestamp() big.Int {
	return msg.Timestamp
}
func (msg *P2PBasicMessage) GetId() string {
	return msg.Id
}
func (msg *P2PBasicMessage) GetGossip() bool {
	return msg.Gossip
}
func (msg *P2PBasicMessage) GetNodeId() string {
	return msg.NodeId
}

func (msg *P2PBasicMessage) GetNodePubKey() []byte {
	return msg.NodePubKey
}
func (msg *P2PBasicMessage) GetSign() []byte {
	return msg.Sign
}
func (msg *P2PBasicMessage) GetMsgType() string {
	return msg.MsgType
}
func (msg *P2PBasicMessage) GetPayload() []byte {
	return msg.Payload
}
func (msg *P2PBasicMessage) SetSign(sig []byte) {
	msg.Sign = sig
}
func (msg P2PBasicMessage) GetSerializedBody() []byte {
	msg.SetSign(nil)
	// marshall msg without the signature to bytes format
	bin, err := bijson.Marshal(msg)
	if err != nil {
		log.Errorf("failed to marshal pb message %v", err)
		return nil
	}
	return bin
}

type P2PBasicMessage struct {
	// shared between all requests
	Version    p2pMessageVersion `json:"version,omitempty"`
	Timestamp  big.Int           `json:"timestamp,omitempty"`  // unix time
	Id         string            `json:"id,omitempty"`         // allows requesters to use request data when processing a response
	Gossip     bool              `json:"gossip,omitempty"`     // true to have receiver peer gossip the message to neighbors
	NodeId     string            `json:"nodeId,omitempty"`     // id of node that created the message (not the peer that may have sent it). =base58(multihash(nodePubKey))
	NodePubKey []byte            `json:"nodePubKey,omitempty"` // Authoring node Secp256k1 public key (32bytes)
	Sign       []byte            `json:"sign,omitempty"`       // signature of message data + method specific data by message authoring node.
	MsgType    string            `json:"msgtype,omitempty"`    // identifyng message type
	Payload    []byte            `json:"payload"`              // payload data to be unmarshalled
}

type StreamMessage struct {
	Protocol string
	Message  P2PBasicMessage
}

type EventBusBytes []byte

type KeyStorage struct {
	KeyIndex       big.Int
	Si             big.Int
	Siprime        big.Int
	CommitmentPoly []Point
}
type NodeReference struct {
	Address         *ethCommon.Address
	Index           *big.Int
	PeerID          peer.ID
	PublicKey       *ecdsa.PublicKey
	P2PConnection   string
	TMP2PConnection string
}

type SerializedNodeReference struct {
	Address         [20]byte
	Index           big.Int
	PeerID          string
	PublicKey       Point
	P2PConnection   string
	TMP2PConnection string
}

func (nodeRef NodeReference) Serialize() SerializedNodeReference {
	var nodeRefAddress [20]byte
	var nodeRefIndex big.Int
	var nodeRefPublicKey Point
	if nodeRef.Address != nil {
		nodeRefAddress = *nodeRef.Address
	}
	if nodeRef.Index != nil {
		nodeRefIndex = *nodeRef.Index
	}
	if nodeRef.PublicKey != nil {
		nodeRefPublicKey = Point{
			X: *nodeRef.PublicKey.X,
			Y: *nodeRef.PublicKey.Y,
		}
	}
	return SerializedNodeReference{
		Address:         nodeRefAddress,
		Index:           nodeRefIndex,
		PeerID:          string(nodeRef.PeerID),
		PublicKey:       nodeRefPublicKey,
		P2PConnection:   nodeRef.P2PConnection,
		TMP2PConnection: nodeRef.TMP2PConnection,
	}
}

func (nodeRef NodeReference) Deserialize(serializedNodeRef SerializedNodeReference) NodeReference {
	addr := ethCommon.Address(serializedNodeRef.Address)
	nodeRef.Address = &addr
	nodeRef.Index = &serializedNodeRef.Index
	nodeRef.PeerID = peer.ID(serializedNodeRef.PeerID)
	nodeRef.PublicKey = &ecdsa.PublicKey{Curve: secp256k1.Curve, X: &serializedNodeRef.PublicKey.X, Y: &serializedNodeRef.PublicKey.Y}
	nodeRef.P2PConnection = serializedNodeRef.P2PConnection
	nodeRef.TMP2PConnection = serializedNodeRef.TMP2PConnection
	return nodeRef
}

func GetPeerIDFromP2pListenAddress(p2pListenAddress string) (*peer.ID, error) {
	ipfsaddr, err := ma.NewMultiaddr(p2pListenAddress)
	if err != nil {
		log.WithError(err).Error("could not get ipfsaddr")
		return nil, err
	}

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.WithError(err).Error("could not get pid")
		return nil, err
	}

	peerid, err := peer.Decode(pid)
	if err != nil {
		log.WithError(err).Error("could not get peerid")
		return nil, err
	}

	return &peerid, nil
}

func PadPrivKeyBytes(kBytes []byte) []byte {
	if len(kBytes) < 32 {
		tmp := make([]byte, 32)
		copy(tmp[32-len(kBytes):], kBytes)
		return tmp
	}
	return kBytes
}

type ConnectionDetailsResult struct {
	TMP2PConnection string `json:"tm_p2p_connection"`
	P2PConnection   string `json:"p2p_connection"`
}

func GetSocketAddress() string {
	// return "unix://" + filepath.Join(config.GlobalConfig.BasePath, "dkg.sock")
	return "unix://dkg.sock"
}

func DoesFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else if os.IsPermission(err) {
			return false
		} else {
			return true
		}
	}
	return true
}
