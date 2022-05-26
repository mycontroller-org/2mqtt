package adapter

import (
	"sync"

	config "github.com/mycontroller-org/2mqtt/pkg/types/config"
	"go.uber.org/zap"
)

type store struct {
	services map[string]*Service
	mutex    *sync.Mutex
}

var servicesStore = store{
	services: make(map[string]*Service),
	mutex:    &sync.Mutex{},
}

// Add a service
func (s *store) Add(service *Service) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.services[service.adapterConfig.Name] = service
}

// Remove all the services
func (s *store) StopAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for k := range servicesStore.services {
		service := servicesStore.services[k]
		if service != nil {
			service.Stop()
		}
	}
}

// Start all the services
func Start(adapters []config.AdapterConfig) {
	for index := range adapters {
		adapterCfg := adapters[index]
		if !adapterCfg.Enabled {
			continue
		}
		service, err := NewService(&adapterCfg)
		if err != nil {
			zap.L().Error("error on starting a service", zap.Error(err), zap.String("adapterName", adapterCfg.Name))
			continue
		}
		service.Start()
		servicesStore.Add(service)
	}
}

// Close stops all the services
func Close() {
	servicesStore.StopAll()
}
