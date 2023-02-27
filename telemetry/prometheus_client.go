package telemetry

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var keysGenerated *keysGeneratedCounter
var keysAssigned *keysAssignedCounter
var keyShareCalls *shareReqCounter

func IncrementKeysGenerated() {
	keysGenerated.keysGenerated.Inc()
}

func IncrementKeyAssigned() {
	keysAssigned.keysAssigned.Inc()
}

func IncrementShareReqSuccess() {
	keyShareCalls.success.Inc()
}

func IncrementShareReqFail() {
	keyShareCalls.failure.Inc()
}

func StartClient() {

	keysGenerated = NewKeysGeneratedCounter()
	keysAssigned = NewKeysAssignedCounter()
	keyShareCalls = NewSuccessShareReqCounter()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatalln(http.ListenAndServe(":9090", nil))
}
