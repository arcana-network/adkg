package common

import (
	"fmt"

	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/avast/retry-go"
	log "github.com/sirupsen/logrus"
)

type IService interface {
	ID() string
	Start() error
	Stop() error
	IsRunning() bool
	Call(method string, args ...interface{}) (interface{}, error)
}

type ServiceRegistry struct {
	services map[string]IService
	bus      eventbus.Bus
}

func NewServiceRegistry(bus eventbus.Bus) *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]IService),
		bus:      bus,
	}
}

func (s *ServiceRegistry) RegisterService(service IService) error {
	if _, exists := s.services[service.ID()]; exists {
		return fmt.Errorf("service already exists: %v", service.ID())
	}
	s.services[service.ID()] = service
	return nil
}

func (s *ServiceRegistry) StartAll() error {
	for name, service := range s.services {
		go func(n string, s IService) {
			log.Info(fmt.Sprintf(`Starting service: %s`, n))
			err := s.Start()
			if err != nil {
				log.Info(fmt.Sprintf(`Error during starting service: %s %v`, n, err))
			}
		}(name, service)
	}
	return nil
}

func (s *ServiceRegistry) StopAll() (err error) {
	for name, service := range s.services {
		log.Info(fmt.Sprintf(`Stopping service: %s`, name))
		err = service.Stop()
		if err != nil {
			log.WithError(err).Errorf("Stopping service %q", name)
		}
	}
	return err
}

func (s *ServiceRegistry) SetupMethodRouting() {
	err := s.bus.SubscribeAsync("method", func(data interface{}) {
		methodRequest, ok := data.(MethodRequest)
		if !ok {
			log.Error("could not parse data for query")
			return
		}
		var baseService IService
		err := retry.Do(func() error {
			bs, ok := s.services[methodRequest.Service]
			if !ok {
				log.WithField("Service", methodRequest.Service).Error("could not find service")
				return fmt.Errorf("could not find service %v", methodRequest.Service)
			}
			if !bs.IsRunning() {
				log.WithFields(log.Fields{
					"Service": methodRequest.Service,
				}).Error("ServiceNotRunning")
				return fmt.Errorf("service %v is not running", methodRequest.Service)
			}
			baseService = bs
			return nil
		})
		if err != nil {
			s.bus.Publish(methodRequest.ID, MethodResponse{
				Error: err,
				Data:  nil,
			})
		} else {
			go func() {
				defer func() {
					if err := recover(); err != nil {
						log.Error(err)
						log.WithFields(log.Fields{
							"Caller": methodRequest.Caller,
							"Method": methodRequest.Method,
							"data":   data,
							"error":  err,
						}).Error("panicked during IService.Call")

						resp := MethodResponse{
							Error: fmt.Errorf("%v", err),
							Data:  nil,
						}
						s.bus.Publish(methodRequest.ID, resp)
					}
				}()
				data, err := baseService.Call(methodRequest.Method, methodRequest.Data...)
				resp := MethodResponse{
					Request: methodRequest,
					Error:   err,
					Data:    data,
				}
				s.bus.Publish(methodRequest.ID, resp)
			}()
		}
	}, false)
	if err != nil {
		log.WithError(err).Error("could not subscribe async")
	}
}
