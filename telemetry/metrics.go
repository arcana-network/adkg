package telemetry

import "github.com/prometheus/client_golang/prometheus"

type validatorKeysGeneratedCounter struct {
	keysGenerated *prometheus.CounterVec
}

func NewValidatorKeysGeneratedCounter(reg prometheus.Registerer) *validatorKeysGeneratedCounter {
	m := &validatorKeysGeneratedCounter{
		keysGenerated: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "KeysGenerated",
			Help: "Keys Generated by validator",
		}, []string{validatorIDLabel}),
	}
	reg.MustRegister(m.keysGenerated)
	return m
}
