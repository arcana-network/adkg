package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type keysGeneratedCounter struct {
	keysGenerated prometheus.Counter
}
type keysAssignedCounter struct {
	keysAssigned prometheus.Counter
}
type shareReqCounter struct {
	success prometheus.Counter
	failure prometheus.Counter
}

func NewKeysGeneratedCounter() *keysGeneratedCounter {
	m := &keysGeneratedCounter{
		keysGenerated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "key_generation_participant",
			Help: "Participation in key generation",
		}),
	}
	_ = prometheus.Register(m.keysGenerated)
	return m
}

func NewKeysAssignedCounter() *keysAssignedCounter {
	m := &keysAssignedCounter{
		keysAssigned: promauto.NewCounter(prometheus.CounterOpts{
			Name: "keys_assigned",
			Help: "Keys assignment requests received",
		}),
	}
	_ = prometheus.Register(m.keysAssigned)
	return m
}
func NewSuccessShareReqCounter() *shareReqCounter {
	m := &shareReqCounter{
		success: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "key_share_success_request",
			Help: "Keys share successful requests",
		}),
		failure: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "key_share_failed_request",
			Help: "Keys share failed requests",
		}),
	}
	_ = prometheus.Register(m.success)
	_ = prometheus.Register(m.failure)
	return m
}
