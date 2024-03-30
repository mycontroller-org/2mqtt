package adapter

import (
	"context"
	"sync"

	config "github.com/mycontroller-org/2mqtt/pkg/types/config"
	contextTY "github.com/mycontroller-org/2mqtt/pkg/types/context"
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
func Start(ctx context.Context, adapters []config.AdapterConfig) error {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return err
	}

	for index := range adapters {
		adapterCfg := adapters[index]
		if !adapterCfg.Enabled {
			continue
		}
		logger.Debug("starting an adapter", zap.String("name", adapterCfg.Name), zap.String("provider", adapterCfg.Provider))
		service, err := NewService(ctx, &adapterCfg)
		if err != nil {
			logger.Error("error on starting a service", zap.Error(err), zap.String("adapterName", adapterCfg.Name))
			continue
		}
		service.Start()
		servicesStore.Add(service)
	}

	return nil
}

// Close stops all the services
func Close() {
	servicesStore.StopAll()
}
