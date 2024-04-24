package common

import (
	"errors"

	"github.com/arcana-network/dkgnode/secp256k1"
	"github.com/avast/retry-go"
	"github.com/libp2p/go-libp2p/core/protocol"
	log "github.com/sirupsen/logrus"
)

func SendForMessageType(broker *MessageBroker, nodeDetails NodeDetails, byt []byte, msgType string, protocolId protocol.ID) error {
	// get recipient details
	nodeReference := broker.ChainMethods().GetNodeDetailsByAddress(PointToEthAddress(Point(nodeDetails.PubKey)))

	p2pMsg := broker.P2PMethods().NewP2PMessage(secp256k1.HashToString(byt), false, byt, msgType)
	log.WithField("P2P connection string", nodeReference.P2PConnection).Debug()
	peerID, err := GetPeerIDFromP2pListenAddress(nodeReference.P2PConnection)
	if err != nil {
		return err
	}
	// sign the data
	signature, err := broker.P2PMethods().SignP2PMessage(&p2pMsg)
	if err != nil {
		return errors.New("failed to sign p2p Message" + err.Error())
	}
	p2pMsg.Sign = signature
	err = retry.Do(func() error {
		err := broker.P2PMethods().SendP2PMessage(*peerID, protocolId, &p2pMsg)
		if err != nil {
			log.WithFields(log.Fields{
				"peerID":     peerID,
				"protocolID": protocolId,
			}).WithError(err).Debug("error when sending p2p message")
			return err
		}
		return nil
	})
	if err != nil {
		log.Error("Could not send the p2p message, failed after retries " + err.Error())
		return err
	}

	return nil
}
