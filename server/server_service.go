package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"
	"github.com/arcana-network/dkgnode/server/rpc"

	"github.com/arcana-network/dkgnode/eventbus"
	eth "github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

const contentType = "application/json; charset=utf-8"

type ConnectionDetailsParams struct {
	PubKeyX                  string                   `json:"pubkeyx"`
	PubKeyY                  string                   `json:"pubkeyy"`
	ConnectionDetailsMessage ConnectionDetailsMessage `json:"connection_details_message"`
	Signature                []byte                   `json:"signature"`
}
type ConnectionDetailsJRPCResponse struct {
	RpcVersion string                         `json:"jsonrpc"`
	Id         int                            `json:"id"`
	Result     common.ConnectionDetailsResult `json:"result"`
}

type ConnectionDetailsMessage struct {
	Timestamp   string      `json:"timestamp"`
	Message     string      `json:"message"`
	NodeAddress eth.Address `json:"node_address"`
}

type ConnectionDetailsRequestBody struct {
	RpcVersion string                  `json:"jsonrpc"`
	Method     string                  `json:"method"`
	Id         int                     `json:"id"`
	Params     ConnectionDetailsParams `json:"params"`
}

func (c *ConnectionDetailsMessage) String() string {
	return strings.Join([]string{c.Timestamp, c.Message, c.NodeAddress.String()}, common.Delimiter1)
}

func (c *ConnectionDetailsMessage) Validate(pubKeyX, pubKeyY big.Int, sig []byte) (bool, error) {
	message := c.Message
	if message != "ConnectionDetails" {
		log.WithField("cMessage", c.Message).Error("message not ConnectionDetails")
		return false, errors.New("message is not ConnectionDetails")
	}
	timeSigned := c.Timestamp
	unixTime, err := strconv.ParseInt(timeSigned, 10, 64)
	if err != nil {
		log.WithError(err).Error("could not parse time signed ")
		return false, err
	}
	if time.Unix(unixTime, 0).Add(10 * time.Minute).Before(time.Now()) {
		log.WithError(err).Error("signature expired")
		return false, err
	}
	return common.ECDSAVerify(c.String(), &common.Point{X: pubKeyX, Y: pubKeyY}, sig), nil
}

type ServerService struct {
	bus    eventbus.Bus
	client *http.Client
	server *http.Server
	broker *common.MessageBroker
}

func New(bus eventbus.Bus) *ServerService {
	return &ServerService{
		bus:    bus,
		broker: common.NewServiceBroker(bus, common.SERVER_SERVICE_NAME),
	}
}

func (s *ServerService) ID() string {
	return common.SERVER_SERVICE_NAME
}

func (s *ServerService) Start() error {
	addr := fmt.Sprintf(":%s", config.GlobalConfig.HttpServerPort)
	s.client = &http.Client{
		Timeout: 30 * time.Second,
	}
	s.server = createServer(s.bus, addr)
	go startServer(s.server)

	return nil
}

func startServer(server *http.Server) {
	err := server.ListenAndServe()
	if err != nil {
		log.WithError(err).Fatal()
	}
}

func createServer(bus eventbus.Bus, addr string) *http.Server {
	router := setUpRouter(bus)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	return server
}

func (s *ServerService) Stop() error {
	s.client.CloseIdleConnections()
	err := s.server.Shutdown(context.Background())
	return err
}
func (service *ServerService) IsRunning() bool {
	return true
}
func (service *ServerService) Call(method string, args ...interface{}) (interface{}, error) {
	switch method {
	case "request_connection_details":

		var args0 string
		_ = common.CastOrUnmarshal(args[0], &args0)

		return service.RequestConnectionDetails(args0)
	}
	return nil, fmt.Errorf("server service method %v not found", method)
}

func (s *ServerService) RequestConnectionDetails(endpoint string) (connectionDetails common.ConnectionDetails, err error) {
	pubKey := s.broker.ChainMethods().GetSelfPublicKey()
	privKey := s.broker.ChainMethods().GetSelfPrivateKey()
	addr := s.broker.ChainMethods().GetSelfAddress()
	connectionDetailsMessage := ConnectionDetailsMessage{
		Message:     "ConnectionDetails",
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
		NodeAddress: addr,
	}
	sig := common.ECDSASign(connectionDetailsMessage.String(), &privKey)
	connectionDetailsParams := ConnectionDetailsParams{
		PubKeyX:                  pubKey.X.Text(16),
		PubKeyY:                  pubKey.Y.Text(16),
		ConnectionDetailsMessage: connectionDetailsMessage,
		Signature:                sig,
	}
	connectionDetailsRequestBody := ConnectionDetailsRequestBody{
		RpcVersion: "2.0",
		Method:     "ConnectionDetails",
		Id:         10,
		Params:     connectionDetailsParams,
	}

	body, err := bijson.Marshal(connectionDetailsRequestBody)
	if err != nil {
		return connectionDetails, err
	}

	substrs := strings.Split(endpoint, ":")
	var uri, port string
	if len(substrs) < 1 {
		return connectionDetails, errors.New("could not get uri from endpoint")
	} else if len(substrs) < 2 {
		uri = substrs[0]
	} else {
		uri = substrs[0]
		port = substrs[1]
	}

	var respErr error
	var resp *http.Response

	if port != "" {
		resp, respErr = s.client.Post(fmt.Sprintf("http://%s:%s/rpc", uri, port), contentType, bytes.NewBuffer(body))
		log.WithFields(log.Fields{
			"status":  resp.StatusCode,
			"respErr": respErr,
		}).Info("RequestConnectionDetails()")
		if respErr != nil || resp.StatusCode >= 400 {
			log.WithError(respErr).Error("could not get connection details http uri port")
		}
	}
	if respErr != nil || resp == nil || resp.StatusCode >= 400 {
		log.WithError(respErr).Error("could not get connection details http uri port")
		resp, respErr = s.client.Post(fmt.Sprintf("https://%s/rpc", uri), contentType, bytes.NewBuffer(body))
	}
	if respErr != nil || resp.StatusCode >= 400 {
		log.WithError(respErr).Error("could not get connection details https uri only")
		resp, respErr = s.client.Post(fmt.Sprintf("http://%s/rpc", uri), contentType, bytes.NewBuffer(body))
	}
	if respErr != nil || resp.StatusCode >= 400 {
		log.WithError(respErr).Error("could not get connection details http uri only")
		return connectionDetails, errors.New("could not get connection Details")
	}

	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return connectionDetails, err
	}
	log.WithField("responseBody", string(responseBody)).Debug("responseBody")
	var connectionDetailsJRPCResponse ConnectionDetailsJRPCResponse
	err = bijson.Unmarshal(responseBody, &connectionDetailsJRPCResponse)
	if err != nil {
		return connectionDetails, err
	}
	res := common.ConnectionDetails{
		TMP2PConnection: connectionDetailsJRPCResponse.Result.TMP2PConnection,
		P2PConnection:   connectionDetailsJRPCResponse.Result.P2PConnection,
	}
	log.WithField("res", res).Debug("got connection details")
	return res, nil
}

func setUpRouter(eventBus eventbus.Bus) http.Handler {
	mr, err := rpc.SetUpJRPCHandler(eventBus)
	if err != nil {
		log.WithError(err).Fatal()
	}

	router := mux.NewRouter().StrictSlash(true)

	router.Handle("/rpc", mr)
	// AttachProfiler(router)

	router.Use(parseBodyMiddleware)
	router.Use(augmentRequestMiddleware)
	router.Use(loggingMiddleware)

	handler := cors.Default().Handler(router)
	return handler
}

// func AttachProfiler(router *mux.Router) {
// 	router.HandleFunc("/debug/pprof/", pprof.Index)
// 	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
// 	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
// 	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

// 	// Manually add support for paths linked to by index page at /debug/pprof/
// 	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
// 	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
// 	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
// 	router.Handle("/debug/pprof/block", pprof.Handler("block"))
// }
