package p2p

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"

	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/torusresearch/bijson"

	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
)

type MessageType int32

const (
	RUNNING MessageType = 1
	PAUSED  MessageType = 0
	STOPPED MessageType = -1
)

type Key string

const CONTEXT_KEY = Key("id")

type P2PService struct {
	bus         eventbus.Bus
	p2pNode     host.Host
	hostAddress multiaddr.Multiaddr
	publicKey   common.Point
	status      MessageType
	signData    func(data []byte) (rawSig []byte)
	broker      *common.MessageBroker
	context     context.Context
	cancel      context.CancelFunc
}

func New(bus eventbus.Bus) *P2PService {
	p2pService := &P2PService{
		status: STOPPED,
		bus:    bus,
		broker: common.NewServiceBroker(bus, common.P2P_SERVICE_NAME),
	}
	return p2pService
}

func (*P2PService) ID() string {
	return common.P2P_SERVICE_NAME
}

func (service *P2PService) Start() error {
	context, cancel := context.WithCancel(context.Background())
	service.context = context
	service.cancel = cancel

	privKey := getPrivKey(service.broker)

	service.publicKey = service.broker.ChainMethods().GetSelfPublicKey()

	node, err := createLibp2pNode(privKey, service.context)
	if err != nil {
		log.WithError(err).Error("create_lib_p2p")
		return err
	}

	fullAddr := getHostAddress(node)

	service.hostAddress = fullAddr
	service.p2pNode = node
	service.status = RUNNING
	service.signData = service.broker.ChainMethods().SelfSignData

	log.Info("P2P service running.")
	return nil
}

func getPrivKey(broker *common.MessageBroker) libp2pcrypto.PrivKey {
	privKeyRaw := broker.ChainMethods().GetSelfPrivateKey()
	privKey, err := libp2pcrypto.UnmarshalSecp256k1PrivateKey(padPrivKeyBytes(privKeyRaw.Bytes()))
	if err != nil {
		log.WithError(err).Fatal("could not Unmarshal privateKey")
	}
	return privKey
}

func createLibp2pNode(privKey libp2pcrypto.PrivKey, ctx context.Context) (node host.Host, err error) {
	limiter := rcmgr.NewFixedLimiter(rcmgr.DefaultLimits.AutoScale())
	rcm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return
	}
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%s", "0.0.0.0", config.GlobalConfig.P2PPort)),
		libp2p.Identity(privKey),
		libp2p.DisableRelay(),
		libp2p.ResourceManager(rcm),
	}
	node, err = libp2p.New(opts...)
	return
}

func getHostAddress(node host.Host) multiaddr.Multiaddr {
	log.WithField("nodeid", node.ID().Pretty()).Info("P2P")
	hostAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", node.ID().Pretty()))
	if err != nil {
		log.WithError(err).Fatalln("Error while creating multiaddr")
	}
	addr := node.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.WithField("fullAddress", fullAddr).Info("P2P")
	return fullAddr
}

func (service *P2PService) IsRunning() bool {
	return service.status == RUNNING
}

func (service *P2PService) Stop() error {
	service.cancel()
	service.status = STOPPED
	return nil
}
func (service *P2PService) Pause() error {
	service.status = PAUSED
	return nil
}

func (service *P2PService) GetState() int32 {
	return int32(service.status)
}

func (service *P2PService) Connect(addr multiaddr.Multiaddr) {
	peer, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		log.Error(err)
	}
	if err := service.p2pNode.Connect(service.context, *peer); err != nil {
		log.Error(err)
		panic(err)
	}

}

func (p *P2PService) ForwardP2PToEventBus(proto string) {
	p.p2pNode.SetStreamHandler(protocol.ID(proto), func(s network.Stream) {
		buf, err := io.ReadAll(s)
		if err != nil {
			log.WithError(err).Error("stream_read")
			if e := s.Reset(); e != nil {
				log.WithError(e).Error("could not reset stream")
			}
			return
		}
		s.Close()

		var p2pMsg common.P2PBasicMessage
		if err = bijson.Unmarshal(buf, &p2pMsg); err != nil {
			log.WithError(err).Error("could not unmarshal p2pmsg")
			return
		}

		if err = p.authenticateMessage(p2pMsg); err != nil {
			log.WithField("Error", err).Error("failed to authenticate p2pMsg")
			return
		}
		p.bus.Publish("p2p:forward:"+proto, p2pMsg)
	})
}

func (service *P2PService) authenticateMessage(data common.P2PBasicMessage) error {
	var pk common.Point
	rawPk := data.GetNodePubKey()
	err := bijson.Unmarshal(rawPk, &pk)
	if err != nil {
		log.WithError(err).Error("could not unmarshal rawpk")
		return err
	}

	bin := data.GetSerializedBody()
	_, err = service.ChainMethods().VerifyDataWithNodelist(pk, data.GetSign(), bin)
	return err
}

func (service *P2PService) ChainMethods() *common.ChainMethods {
	return common.NewServiceBroker(service.bus, "p2p").ChainMethods()
}

func (service *P2PService) String() string {
	return fmt.Sprintf("P2PService: %v", service.context)
}

func (service *P2PService) signP2PMessage(message common.P2PMessage) ([]byte, error) {
	data, err := bijson.Marshal(message)
	if err != nil {
		return nil, err
	}
	return service.signData(data), nil
}

func (service *P2PService) Call(method string, args ...interface{}) (interface{}, error) {
	switch method {
	case "id":
		return service.p2pNode.ID(), nil
	case "get_host_address":

		if service.hostAddress == nil {
			return nil, errors.New("hostAddress not initialized")
		}
		return service.hostAddress.String(), nil
	case "set_stream_handler":
		var args0 string
		_ = common.CastOrUnmarshal(args[0], &args0)

		proto := args0
		service.ForwardP2PToEventBus(proto)
		return true, nil
	case "new_p2p_message":

		var args0, args3 string
		var args1 bool
		var args2 []byte
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)
		_ = common.CastOrUnmarshal(args[2], &args2)
		_ = common.CastOrUnmarshal(args[3], &args3)

		newMsg := *service.NewP2PMessage(args0, args1, args2, args3)
		return newMsg, nil
	case "sign_p2p_message":
		var args0 common.P2PBasicMessage
		_ = common.CastOrUnmarshal(args[0], &args0)

		sig, err := service.signP2PMessage(&args0)
		return sig, err
	case "send_p2p_message":

		var args0 peer.ID
		var args1 protocol.ID
		var args2 common.P2PBasicMessage
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)
		_ = common.CastOrUnmarshal(args[2], &args2)

		ctx := context.Background()
		err := service.sendP2PMessage(ctx, args0, args1, &args2)
		return nil, err
	case "authenticate_message":

		var args0 common.P2PBasicMessage
		_ = common.CastOrUnmarshal(args[0], &args0)

		p2pMsg := args0
		err := service.authenticateMessage(p2pMsg)
		return nil, err
	case "connect_to_p2p_node":

		var args0 string
		var args1 peer.ID
		_ = common.CastOrUnmarshal(args[0], &args0)
		_ = common.CastOrUnmarshal(args[1], &args1)

		err := service.ConnectToP2PNode(args0, args1)
		return nil, err
	}
	return false, nil
}

func (service *P2PService) sendP2PMessage(ctx context.Context, id peer.ID, p protocol.ID, msg common.P2PMessage) error {
	data, err := bijson.Marshal(msg)
	if err != nil {
		return err
	}

	s, err := service.p2pNode.NewStream(ctx, id, p)
	if err != nil {
		log.WithError(err).Error("failed to create stream")
		return err
	}

	err = writeMessage(s, data)
	if err != nil {
		log.WithError(err).Error("failed to write on stream")
		resetErr := s.Reset()
		if resetErr != nil {
			log.WithError(resetErr).Error("[write] failed to reset stream")
		}
	}
	return err
}

func writeMessage(s network.Stream, data []byte) error {
	_, err := s.Write(data)
	if err != nil {
		return err
	}

	err = s.Close()
	if err != nil {
		return err
	}

	return nil
}

func (service *P2PService) ConnectToP2PNode(nodeP2PConnection string, nodePeerID peer.ID) error {

	peerAdded := false
	for _, peer := range service.p2pNode.Peerstore().Peers() {
		if peer == nodePeerID {
			peerAdded = true
		}
	}

	if nodePeerID != service.p2pNode.ID() && !peerAdded {
		log.WithField("nodeP2PConnection", nodeP2PConnection).Info("adding nodeP2PConnection to addressbook")
		targetPeerAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peer.Encode(nodePeerID)))
		a, _ := multiaddr.NewMultiaddr(nodeP2PConnection)
		targetAddr := a.Decapsulate(targetPeerAddr)

		service.p2pNode.Peerstore().AddAddr(nodePeerID, targetAddr, peerstore.PermanentAddrTTL)
	}

	return nil
}

func (service *P2PService) NewP2PMessage(messageId string, gossip bool, payload []byte, msgType string) *common.P2PBasicMessage {
	rawPk, err := bijson.Marshal(service.publicKey)
	if err != nil {
		log.Error("could not marshal pk in newp2pmessage" + err.Error())
	}
	p2pBasicMsg := common.CreateP2PBasicMessage(common.P2PBasicMessageRaw{
		NodeId:     service.p2pNode.ID().String(),
		NodePubKey: rawPk,
		Timestamp:  *big.NewInt(time.Now().Unix()),
		Id:         messageId,
		Gossip:     gossip,
		Payload:    payload,
		MsgType:    msgType,
	})
	return &p2pBasicMsg
}

func padPrivKeyBytes(kBytes []byte) []byte {
	if len(kBytes) < 32 {
		tmp := make([]byte, 32)
		copy(tmp[32-len(kBytes):], kBytes)
		return tmp
	}
	return kBytes
}
