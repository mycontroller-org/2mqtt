package adapter

import (
	"fmt"
	"sync"
	"time"

	"github.com/mycontroller-org/2mqtt/pkg/queue"
	scheduler "github.com/mycontroller-org/2mqtt/pkg/service/scheduler"
	"github.com/mycontroller-org/2mqtt/pkg/types"
	config "github.com/mycontroller-org/2mqtt/pkg/types/config"
	devicePlugin "github.com/mycontroller-org/2mqtt/plugin/device"
	providerPlugin "github.com/mycontroller-org/2mqtt/plugin/provider"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"go.uber.org/zap"
)

// default settings
const (
	SourceQueueLimit      = 1000
	MQTTQueueLimit        = 1000
	DefaultReconnectDelay = "30s"

	MqttDeviceName = "mqtt"

	defaultMessageProcessTick = 1 * time.Second
)

// Service component of the provider
type Service struct {
	adapterConfig       *config.AdapterConfig
	provider            types.Formatter
	sourceDevice        types.Device
	mqttDevice          types.Device
	sourceMessageQueue  *queue.MessageQueue
	mqttMessageQueue    *queue.MessageQueue
	statusSource        types.State
	statusMqtt          types.State
	mutex               *sync.RWMutex
	reconnectDelay      string
	sourceID            string
	mqttID              string
	terminateMqttChan   *concurrency.Channel
	terminateSourceChan *concurrency.Channel
}

// NewService creates brand new Service
func NewService(adapterCfg *config.AdapterConfig) (*Service, error) {
	provider, err := providerPlugin.Create(adapterCfg.Provider, adapterCfg.Source)
	if err != nil {
		return nil, err
	}

	s := &Service{
		adapterConfig:       adapterCfg,
		provider:            provider,
		mutex:               &sync.RWMutex{},
		sourceID:            fmt.Sprintf("%s_adapter_source", adapterCfg.Name),
		mqttID:              fmt.Sprintf("%s_adapter_mqtt", adapterCfg.Name),
		terminateMqttChan:   concurrency.NewChannel(1),
		terminateSourceChan: concurrency.NewChannel(1),
	}
	s.sourceMessageQueue = queue.New(s.sourceID, SourceQueueLimit)
	s.mqttMessageQueue = queue.New(s.mqttID, MQTTQueueLimit)

	// update reconnectDelay
	_, err = time.ParseDuration(adapterCfg.ReconnectDelay)
	if err != nil {
		zap.L().Info("error on parsing reconnect delay, running with default", zap.String("reconnectDelay", adapterCfg.ReconnectDelay), zap.String("default", DefaultReconnectDelay), zap.Error(err))
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
	go s.sourceMessageProcessor()
	go s.mqttMessageProcessor()
}

// Start stops a adapter service
func (s *Service) Stop() {
	if s.sourceDevice != nil {
		err := s.sourceDevice.Close()
		if err != nil {
			zap.L().Error("error on closing a source connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		}
	}
	if s.mqttDevice != nil {
		err := s.mqttDevice.Close()
		if err != nil {
			zap.L().Error("error on closing a mqtt connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		}
	}
	// close message queues
	s.terminateMqttChan.SafeClose()
	s.terminateSourceChan.SafeClose()
}

func (s *Service) mqttMessageProcessor() {
	ticker := time.NewTicker(defaultMessageProcessTick)
	defer ticker.Stop()

	for {
		select {
		case <-s.terminateMqttChan.CH:
			return
		case <-ticker.C:
			for {
				if s.statusMqtt.Status == types.StatusUP {
					message := s.mqttMessageQueue.Get()
					if message == nil {
						break
					}
					message.Others.Set(types.KeyMqttQoS, int(s.adapterConfig.MQTT.GetInt64(types.KeyMqttQoS)), nil)
					err := s.mqttDevice.Write(message)
					if err != nil {
						zap.L().Error("error on writing a message to mqtt", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
					}
				} else {
					break
				}
			}
		}
	}
}

func (s *Service) sourceMessageProcessor() {
	ticker := time.NewTicker(defaultMessageProcessTick)
	defer ticker.Stop()

	for {
		select {
		case <-s.terminateSourceChan.CH:
			return
		case <-ticker.C:
			for {
				if s.statusSource.Status == types.StatusUP {
					message := s.sourceMessageQueue.Get()
					if message == nil {
						break
					}
					zap.L().Debug("posting a message to source device", zap.String("message", message.ToString()), zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider))
					err := s.sourceDevice.Write(message)
					if err != nil {
						zap.L().Error("error on writing a message to source device", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
					}
				} else {
					break
				}
			}
		}
	}
}

func (s *Service) onMqttMessage(message *types.Message) {
	zap.L().Debug("received a mqtt message", zap.String("message", message.ToString()), zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider))
	formattedMsg, err := s.provider.ToSourceMessage(message)
	if err != nil {
		zap.L().Error("error on formatting to source type", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		return
	}
	s.sourceMessageQueue.Add(formattedMsg)
}

func (s *Service) onSourceMessage(message *types.Message) {
	zap.L().Debug("received a message from source device", zap.String("message", message.ToString()), zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider))
	formattedMsg, err := s.provider.ToMQTTMessage(message)
	if err != nil {
		zap.L().Error("error on formatting to mqtt", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		return
	}
	s.mqttMessageQueue.Add(formattedMsg)
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
	err := scheduler.Schedule(s.mqttID, s.reconnectDelay, s.reconnectMqttDevice)
	if err != nil {
		zap.L().Error("error on configuring a schedule", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.String("id", s.mqttID), zap.Error(err))
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
	err := scheduler.Schedule(s.sourceID, s.reconnectDelay, s.reconnectSourceDevice)
	if err != nil {
		zap.L().Error("error on configuring a schedule", zap.String("id", s.sourceID), zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
	}

}

func (s *Service) reconnectMqttDevice() {
	scheduler.Unschedule(s.mqttID)
	if s.statusMqtt.Status == types.StatusUP {
		return
	}

	if s.mqttDevice != nil {
		err := s.mqttDevice.Close()
		if err != nil {
			zap.L().Error("error on colsing a mqtt connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		}
	}

	mqttDevice, err := devicePlugin.Create(MqttDeviceName, s.adapterConfig.Name, s.adapterConfig.MQTT, s.onMqttMessage, s.onMqttStatus)
	if err == nil {
		s.mqttDevice = mqttDevice
		// update status UP
		s.statusMqtt = types.State{
			Status: types.StatusUP,
			Since:  time.Now(),
		}
		zap.L().Info("connected to the mqtt broker", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider))
		return
	}
	zap.L().Error("error on getting mqtt connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.String("reconnectDelay", s.reconnectDelay), zap.Error(err))
	// schedule a job for reconnect
	err = scheduler.Schedule(s.mqttID, s.reconnectDelay, s.reconnectMqttDevice)
	if err != nil {
		zap.L().Error("error on configuring a schedule", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.String("id", s.mqttID), zap.Error(err))
	}
}

func (s *Service) reconnectSourceDevice() {
	scheduler.Unschedule(s.sourceID)
	if s.statusSource.Status == types.StatusUP {
		return
	}

	if s.sourceDevice != nil {
		err := s.sourceDevice.Close()
		if err != nil {
			zap.L().Error("error on closing a source connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.Error(err))
		}
	}

	sourceDevice, err := devicePlugin.Create(s.adapterConfig.Source.GetString(types.KeyType), s.adapterConfig.Name, s.adapterConfig.Source, s.onSourceMessage, s.onSourceStatus)
	if err == nil {
		s.sourceDevice = sourceDevice
		// update status UP
		s.statusSource = types.State{
			Status: types.StatusUP,
			Since:  time.Now(),
		}
		zap.L().Info("connected to the source device", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider))
		return
	}
	zap.L().Error("error on getting source connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.String("reconnectDelay", s.reconnectDelay), zap.Error(err))
	// schedule a job for reconnect
	err = scheduler.Schedule(s.sourceID, s.reconnectDelay, s.reconnectSourceDevice)
	if err != nil {
		zap.L().Error("error on configuring a schedule", zap.String("adapterName", s.adapterConfig.Name), zap.String("provider", s.adapterConfig.Provider), zap.String("id", s.sourceID), zap.Error(err))
	}
}
