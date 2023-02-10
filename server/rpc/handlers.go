package rpc

import (
	"time"

	"github.com/arcana-network/dkgnode/config"

	"github.com/arcana-network/dkgnode/eventbus"
	eth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/osamingo/jsonrpc/v2"
)

const (
	ConnectionDetailsMethod    = "ConnectionDetails"
	KeyAssignMethod            = "KeyAssign"
	KeyCommitmentRequestMethod = "KeyCommitmentRequest"
	KeyShareRequestMethod      = "KeyShareRequest"
	PublicKeyLookupMethod      = "PublicKeyLookup"
	StoreKeyShareMethod        = "StoreKeyShare"
	RetrieveKeyShareMethod     = "RetrieveKeyShare"
	HealthMethod               = "HealthCheck"
)

type (
	KeyLookupHandler struct {
		bus eventbus.Bus
	}
	ShareRequestHandler struct {
		bus     eventbus.Bus
		TimeNow func() time.Time
	}
	StoreKeyRequestHandler struct {
		bus    eventbus.Bus
		client *ethclient.Client
	}
	RetrieveKeyRequestHandler struct {
		bus    eventbus.Bus
		client *ethclient.Client
	}

	ConnectionDetailsParams struct {
		PubKeyX                  string                   `json:"pubkeyx"`
		PubKeyY                  string                   `json:"pubkeyy"`
		ConnectionDetailsMessage ConnectionDetailsMessage `json:"connection_details_message"`
		Signature                []byte                   `json:"signature"`
	}
	ConnectionDetailsMessage struct {
		Timestamp   string      `json:"timestamp"`
		Message     string      `json:"message"`
		NodeAddress eth.Address `json:"node_address"`
	}
	ConnectionDetailsResult struct {
		TMP2PConnection string `json:"tm_p2p_connection"`
		P2PConnection   string `json:"p2p_connection"`
	}
	CommitmentRequestHandler struct {
		bus     eventbus.Bus
		TimeNow func() time.Time
	}
)

func SetUpJRPCHandler(eventBus eventbus.Bus) (*jsonrpc.MethodRepository, error) {
	mr := jsonrpc.NewMethodRepository()

	if err := mr.RegisterMethod(HealthMethod, HealthHandler{}, HealthParams{}, HealthResult{}); err != nil {
		return nil, err
	}
	if err := mr.RegisterMethod(KeyAssignMethod, KeyAssignHandler{eventBus}, KeyAssignParams{}, KeyAssignResult{}); err != nil {
		return nil, err
	}
	if err := mr.RegisterMethod(ConnectionDetailsMethod, ConnectionDetailsHandler{eventBus}, ConnectionDetailsParams{}, ConnectionDetailsResult{}); err != nil {
		return nil, err
	}

	if err := mr.RegisterMethod(PublicKeyLookupMethod, VerifierLookupHandler{eventBus}, VerifierLookupParams{}, VerifierLookupResult{}); err != nil {
		return nil, err
	}

	if err := mr.RegisterMethod(KeyCommitmentRequestMethod, CommitmentRequestHandler{bus: eventBus, TimeNow: time.Now}, CommitmentRequestParams{}, CommitmentRequestResult{}); err != nil {
		return nil, err
	}

	if err := mr.RegisterMethod(KeyShareRequestMethod, ShareRequestHandler{bus: eventBus, TimeNow: time.Now}, ShareRequestParams{}, ShareRequestResult{}); err != nil {
		return nil, err
	}

	client, err := ethclient.Dial(config.GlobalConfig.EthConnection)
	if err != nil {
		return nil, err
	}

	if err := mr.RegisterMethod(StoreKeyShareMethod, StoreKeyRequestHandler{bus: eventBus, client: client}, StoreKeyRequestParams{}, StoreResult{}); err != nil {
		return nil, err
	}
	if err := mr.RegisterMethod(RetrieveKeyShareMethod, RetrieveKeyRequestHandler{bus: eventBus, client: client}, RetrieveKeyRequestParams{}, RetrieveResult{}); err != nil {
		return nil, err
	}
	return mr, nil
}
