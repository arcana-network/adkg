package dpss

import (
	"github.com/arcana-network/dkgnode/common"
)


type PssNode struct {
	broker  					*common.MessageBroker
	// TODO how to decouple this? A name change will incur many changes in keygen
	details 					common.KeygenNodeDetails
	Transport         *common.NodeTransport

}