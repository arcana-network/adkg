package telemetry

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func IncrementKeysGenerated(validatorID string) {
	keysGenerated.keysGenerated.With(prometheus.Labels{
		validatorIDLabel: validatorID}).Inc()
}

func StartClient() {
	prometheusRegistry = prometheus.NewRegistry()

	keysGenerated = NewValidatorKeysGeneratedCounter(prometheusRegistry)
	http.Handle(
		"/metrics", promhttp.HandlerFor(
			prometheusRegistry,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			}),
	)
	log.Fatalln(http.ListenAndServe(":9090", nil))
}

var prometheusRegistry *prometheus.Registry

const validatorIDLabel = "validatorID"

var keysGenerated *validatorKeysGeneratedCounter
