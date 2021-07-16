package adapter

import (
	"fmt"
	"sync"
	"time"

	"github.com/mycontroller-org/2mqtt/pkg/model"
	config "github.com/mycontroller-org/2mqtt/pkg/model/config"
	"github.com/mycontroller-org/2mqtt/pkg/queue"
	scheduler "github.com/mycontroller-org/2mqtt/pkg/service/scheduler"
	devicePlugin "github.com/mycontroller-org/2mqtt/plugin/device"
	providerPlugin "github.com/mycontroller-org/2mqtt/plugin/provider"
	"github.com/mycontroller-org/backend/v2/pkg/utils/concurrency"
	"go.uber.org/zap"
)

// default settings
const (
	SourceQueueLimit      = 1000
	MQTTQueueLimit        = 1000
	DefaultReconnectDelay = "30s"

	MqttDeviceName = "mqtt"
)

// Service component of the provider
type Service struct {
	adapterConfig       *config.AdapterConfig
	provider            model.Formatter
	sourceDevice        model.Device
	mqttDevice          model.Device
	sourceMessageQueue  *queue.MessageQueue
	mqttMessageQueue    *queue.MessageQueue
	statusSource        model.State
	statusMqtt          model.State
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
		zap.L().Info("error on parsing reconnect delay, running with default", zap.String("reconnectDelay", adapterCfg.ReconnectDelay), zap.String("default", DefaultReconnectDelay))
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
			zap.L().Error("error on colsing a source connection", zap.Error(err), zap.String("adapterName", s.adapterConfig.Name))
		}
	}
	if s.mqttDevice != nil {
		err := s.mqttDevice.Close()
		if err != nil {
			zap.L().Error("error on colsing a mqtt connection", zap.Error(err), zap.String("adapterName", s.adapterConfig.Name))
		}
	}
	// close message queues
	s.terminateMqttChan.SafeClose()
	s.terminateSourceChan.SafeClose()
}

func (s *Service) mqttMessageProcessor() {
	ticker := time.NewTicker(10 * time.Microsecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.terminateMqttChan.CH:
			return
		case <-ticker.C:
			if s.statusMqtt.Status == model.StatusUP {
				if message := s.mqttMessageQueue.Get(); message != nil {
					message.Others.Set(model.KeyMqttQoS, int(s.adapterConfig.MQTT.GetInt64(model.KeyMqttQoS)), nil)
					err := s.mqttDevice.Write(message)
					if err != nil {
						zap.L().Error("error on writing a message to mqtt", zap.Error(err), zap.String("adapterName", s.adapterConfig.Name))
					}
				}
			}
		}
	}
}

func (s *Service) sourceMessageProcessor() {
	ticker := time.NewTicker(10 * time.Microsecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.terminateSourceChan.CH:
			return
		case <-ticker.C:
			if s.statusSource.Status == model.StatusUP {
				if message := s.sourceMessageQueue.Get(); message != nil {
					zap.L().Debug("posting a message to source device", zap.String("message", message.ToString()))
					err := s.sourceDevice.Write(message)
					if err != nil {
						zap.L().Error("error on writing a message to source device", zap.Error(err), zap.String("adapterName", s.adapterConfig.Name))
					}
				}
			}
		}
	}
}

func (s *Service) onMqttMessage(message *model.Message) {
	zap.L().Debug("received a mqtt message", zap.String("message", message.ToString()))
	formattedMsg, err := s.provider.ToSourceMessage(message)
	if err != nil {
		zap.L().Error("error on formatting to source type", zap.Error(err), zap.String("adapterName", s.adapterConfig.Name))
		return
	}
	s.sourceMessageQueue.Add(formattedMsg)
}

func (s *Service) onSourceMessage(message *model.Message) {
	zap.L().Debug("received a message from source device", zap.String("message", message.ToString()))
	formattedMsg, err := s.provider.ToMQTTMessage(message)
	if err != nil {
		zap.L().Error("error on formatting to mqtt", zap.Error(err), zap.String("adapterName", s.adapterConfig.Name))
		return
	}
	s.mqttMessageQueue.Add(formattedMsg)
}

func (s *Service) onMqttStatus(state *model.State) {
	if state == nil {
		return
	}
	s.statusMqtt = *state

	if state.Status == model.StatusUP {
		return
	}

	// schedule a job for reconnect
	err := scheduler.Schedule(s.mqttID, s.reconnectDelay, s.reconnectMqttDevice)
	if err != nil {
		zap.L().Error("error on configuring a schedule", zap.Error(err), zap.String("id", s.mqttID))
	}

}

func (s *Service) onSourceStatus(state *model.State) {
	if state == nil {
		return
	}
	s.statusSource = *state

	if state.Status == model.StatusUP {
		return
	}

	// schedule a job for reconnect
	err := scheduler.Schedule(s.sourceID, s.reconnectDelay, s.reconnectSourceDevice)
	if err != nil {
		zap.L().Error("error on configuring a schedule", zap.Error(err), zap.String("id", s.sourceID))
	}

}

func (s *Service) reconnectMqttDevice() {
	scheduler.Unschedule(s.mqttID)
	if s.statusMqtt.Status == model.StatusUP {
		return
	}

	if s.mqttDevice != nil {
		err := s.mqttDevice.Close()
		if err != nil {
			zap.L().Error("error on colsing a mqtt connection", zap.Error(err), zap.String("adapterName", s.adapterConfig.Name))
		}
	}

	mqttDevice, err := devicePlugin.Create(MqttDeviceName, s.adapterConfig.Name, s.adapterConfig.MQTT, s.onMqttMessage, s.onMqttStatus)
	if err == nil {
		s.mqttDevice = mqttDevice
		// update status UP
		s.statusMqtt = model.State{
			Status: model.StatusUP,
			Since:  time.Now(),
		}
		zap.L().Info("connected to the mqtt broker", zap.String("adapterName", s.adapterConfig.Name))
		return
	}
	zap.L().Error("error on getting mqtt connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("reconnectDelay", s.reconnectDelay), zap.Error(err))
	// schedule a job for reconnect
	err = scheduler.Schedule(s.mqttID, s.reconnectDelay, s.reconnectMqttDevice)
	if err != nil {
		zap.L().Error("error on configuring a schedule", zap.Error(err), zap.String("id", s.mqttID))
	}
}

func (s *Service) reconnectSourceDevice() {
	scheduler.Unschedule(s.sourceID)
	if s.statusSource.Status == model.StatusUP {
		return
	}

	if s.sourceDevice != nil {
		err := s.sourceDevice.Close()
		if err != nil {
			zap.L().Error("error on colsing a source connection", zap.Error(err), zap.String("adapterName", s.adapterConfig.Name))
		}
	}

	sourceDevice, err := devicePlugin.Create(s.adapterConfig.Source.GetString(model.KeyType), s.adapterConfig.Name, s.adapterConfig.MQTT, s.onSourceMessage, s.onSourceStatus)
	if err == nil {
		s.sourceDevice = sourceDevice
		// update status UP
		s.statusSource = model.State{
			Status: model.StatusUP,
			Since:  time.Now(),
		}
		zap.L().Info("connected to the source device", zap.String("adapterName", s.adapterConfig.Name))
		return
	}
	zap.L().Error("error on getting source connection", zap.String("adapterName", s.adapterConfig.Name), zap.String("reconnectDelay", s.reconnectDelay), zap.Error(err))
	// schedule a job for reconnect
	err = scheduler.Schedule(s.sourceID, s.reconnectDelay, s.reconnectSourceDevice)
	if err != nil {
		zap.L().Error("error on configuring a schedule", zap.Error(err), zap.String("id", s.sourceID))
	}
}
