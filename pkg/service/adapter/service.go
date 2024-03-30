package adapter

import (
	"context"
	"fmt"
	"sync"
	"time"

	scheduler "github.com/mycontroller-org/2mqtt/pkg/service/scheduler"
	"github.com/mycontroller-org/2mqtt/pkg/types"
	config "github.com/mycontroller-org/2mqtt/pkg/types/config"
	contextTY "github.com/mycontroller-org/2mqtt/pkg/types/context"
	devicePlugin "github.com/mycontroller-org/2mqtt/plugin/device"
	providerPlugin "github.com/mycontroller-org/2mqtt/plugin/provider"
	queue "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

// default settings
const (
	SourceQueueSize       = 1000
	MQTTQueueSize         = 1000
	DefaultReconnectDelay = "30s"

	MqttDeviceName = "mqtt"
)

// Service component of the provider
type Service struct {
	ctx                context.Context
	logger             *zap.Logger
	scheduler          *scheduler.Scheduler
	adapterConfig      *config.AdapterConfig
	provider           types.Formatter
	sourceDevice       types.Device
	mqttDevice         types.Device
	sourceMessageQueue *queue.Queue
	mqttMessageQueue   *queue.Queue
	statusSource       types.State
	statusMqtt         types.State
	mutex              *sync.RWMutex
	reconnectDelay     string
	sourceID           string
	mqttID             string
}

// NewService creates brand new Service
func NewService(ctx context.Context, adapterCfg *config.AdapterConfig) (*Service, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	scheduler, err := scheduler.FromContext(ctx)
	if err != nil {
		logger.Error("error on getting scheduler", zap.Error(err))
		return nil, err
	}

	provider, err := providerPlugin.Create(ctx, adapterCfg.Provider, adapterCfg.Source, adapterCfg.FormatterScript)
	if err != nil {
		logger.Error("error on get a provider", zap.String("name", adapterCfg.Name), zap.String("provider", adapterCfg.Provider), zap.Error(err))
		return nil, err
	}

	s := &Service{
		ctx:           ctx,
		logger:        logger.Named("adapter"),
		scheduler:     scheduler,
		adapterConfig: adapterCfg,
		provider:      provider,
		mutex:         &sync.RWMutex{},
		sourceID:      fmt.Sprintf("%s_adapter_source", adapterCfg.Name),
		mqttID:        fmt.Sprintf("%s_adapter_mqtt", adapterCfg.Name),
	}
	s.sourceMessageQueue = queue.New(logger, s.sourceID, SourceQueueSize, func(item interface{}) {}, 1)
	s.mqttMessageQueue = queue.New(logger, s.mqttID, MQTTQueueSize, func(item interface{}) {}, 1)

	// update reconnectDelay
	_, err = time.ParseDuration(adapterCfg.ReconnectDelay)
	if err != nil {
		logger.Info("error on parsing reconnect delay, running with default", zap.String("reconnectDelay", adapterCfg.ReconnectDelay), zap.String("default", DefaultReconnectDelay), zap.Error(err))
		s.reconnectDelay = DefaultReconnectDelay
	} else {
		s.reconnectDelay = adapterCfg.ReconnectDelay
	}

	return s, nil
}

// Start starts a adapter service
func (s *Service) Start() {
	s.reconnectMqttDevice()
	s.reconnectSourceDevice()

	s.sourceMessageQueue.Queue.StartConsumers(1, s.sourceMessageProcessor)
	s.mqttMessageQueue.Queue.StartConsumers(1, s.mqttMessageProcessor)
}

// Start stops a adapter service
func (s *Service) Stop() {
	if s.sourceDevice != nil {
		err := s.sourceDevice.Close()
		if err != nil {
			s.logger.Error("error on closing a source connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		}
	}
	if s.mqttDevice != nil {
		err := s.mqttDevice.Close()
		if err != nil {
			s.logger.Error("error on closing a mqtt connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		}
	}

	// close message queues
	if s.mqttMessageQueue != nil {
		s.mqttMessageQueue.Queue.Stop()
	}
	if s.sourceMessageQueue != nil {
		s.sourceMessageQueue.Queue.Stop()
	}
}

func (s *Service) mqttMessageProcessor(item interface{}) {
	if item == nil {
		return
	}
	message, ok := item.(*types.Message)
	if !ok {
		s.logger.Error("error on cast a message", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Any("message", item))
		return
	}
	if message == nil {
		s.logger.Error("message can not be nil", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider))
		return
	}

	if s.statusMqtt.Status == types.StatusUP {
		message.Others.Set(types.KeyMqttQoS, int(s.adapterConfig.MQTT.GetInt64(types.KeyMqttQoS)), nil)
		err := s.mqttDevice.Write(message)
		if err != nil {
			s.logger.Error("error on writing a message to mqtt", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		}
	} else {
		// TODO: this message will be dropped, needs to handle this message
	}
}

func (s *Service) sourceMessageProcessor(item interface{}) {
	if item == nil {
		return
	}
	message, ok := item.(*types.Message)
	if !ok {
		s.logger.Error("error on cast a message", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Any("message", item))
		return
	}
	if message == nil {
		s.logger.Error("message can not be nil", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider))
		return
	}
	if s.statusSource.Status == types.StatusUP {
		s.logger.Debug("posting a message to source device", zap.String("message", message.ToString()), zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider))
		err := s.sourceDevice.Write(message)
		if err != nil {
			s.logger.Error("error on writing a message to source device", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		}
	} else {
		// TODO: this message will be dropped, needs to handle this message
	}
}

func (s *Service) onMqttMessage(message *types.Message) {
	s.logger.Debug("received a mqtt message", zap.String("message", message.ToString()), zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider))
	formattedMsg, err := s.provider.ToSourceMessage(message)
	if err != nil {
		s.logger.Error("error on formatting to source type", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		return
	}
	s.sourceMessageQueue.Produce(formattedMsg)
}

func (s *Service) onSourceMessage(message *types.Message) {
	s.logger.Debug("received a message from source device", zap.String("message", message.ToString()), zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider))
	formattedMsg, err := s.provider.ToMQTTMessage(message)
	if err != nil {
		s.logger.Error("error on formatting to mqtt", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		return
	}
	s.mqttMessageQueue.Produce(formattedMsg)
}

func (s *Service) onMqttStatus(state *types.State) {
	if state == nil {
		return
	}
	s.statusMqtt = *state

	if state.Status == types.StatusUP {
		return
	}

	// schedule a job for reconnect
	err := s.scheduler.Schedule(s.mqttID, s.reconnectDelay, s.reconnectMqttDevice)
	if err != nil {
		s.logger.Error("error on configuring a schedule", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.String("id", s.mqttID), zap.Error(err))
	}

}

func (s *Service) onSourceStatus(state *types.State) {
	if state == nil {
		return
	}
	s.statusSource = *state

	if state.Status == types.StatusUP {
		return
	}

	// schedule a job for reconnect
	err := s.scheduler.Schedule(s.sourceID, s.reconnectDelay, s.reconnectSourceDevice)
	if err != nil {
		s.logger.Error("error on configuring a schedule", zap.String("id", s.sourceID), zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
	}

}

func (s *Service) reconnectMqttDevice() {
	s.scheduler.Unschedule(s.mqttID)
	if s.statusMqtt.Status == types.StatusUP {
		return
	}

	if s.mqttDevice != nil {
		err := s.mqttDevice.Close()
		if err != nil {
			s.logger.Error("error on colsing a mqtt connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		}
	}

	mqttDevice, err := devicePlugin.Create(s.ctx, MqttDeviceName, s.adapterConfig.Name, s.adapterConfig.MQTT, s.onMqttMessage, s.onMqttStatus)
	if err == nil {
		s.mqttDevice = mqttDevice
		// update status UP
		s.statusMqtt = types.State{
			Status: types.StatusUP,
			Since:  time.Now(),
		}
		s.logger.Info("connected to the mqtt broker", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider))
		return
	}
	s.logger.Error("error on getting mqtt connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.String("reconnectDelay", s.reconnectDelay), zap.Error(err))
	// schedule a job for reconnect
	err = s.scheduler.Schedule(s.mqttID, s.reconnectDelay, s.reconnectMqttDevice)
	if err != nil {
		s.logger.Error("error on configuring a schedule", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.String("id", s.mqttID), zap.Error(err))
	}
}

func (s *Service) reconnectSourceDevice() {
	s.scheduler.Unschedule(s.sourceID)
	if s.statusSource.Status == types.StatusUP {
		return
	}

	if s.sourceDevice != nil {
		err := s.sourceDevice.Close()
		if err != nil {
			s.logger.Error("error on closing a source connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		}
	}

	sourceDevice, err := devicePlugin.Create(s.ctx, s.adapterConfig.Source.GetString(types.KeyType), s.adapterConfig.Name, s.adapterConfig.Source, s.onSourceMessage, s.onSourceStatus)
	if err == nil {
		s.sourceDevice = sourceDevice
		// update status UP
		s.statusSource = types.State{
			Status: types.StatusUP,
			Since:  time.Now(),
		}
		s.logger.Info("connected to the source device", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider))
		return
	}
	s.logger.Error("error on getting source connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.String("reconnectDelay", s.reconnectDelay), zap.Error(err))
	// schedule a job for reconnect
	err = s.scheduler.Schedule(s.sourceID, s.reconnectDelay, s.reconnectSourceDevice)
	if err != nil {
		s.logger.Error("error on configuring a schedule", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.String("id", s.sourceID), zap.Error(err))
	}
}
