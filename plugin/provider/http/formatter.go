package http

import (
	"errors"

	model "github.com/mycontroller-org/2mqtt/pkg/types"
)

type SourceType string

func (st SourceType) Name() string {
	return PluginHTTP
}

func (st SourceType) ToSourceMessage(mqttMessage *model.Message) (*model.Message, error) {
	return nil, errors.New("write not supported in http device")
}

func (st SourceType) ToMQTTMessage(sourceMessage *model.Message) (*model.Message, error) {
	if len(sourceMessage.Data) == 0 {
		return nil, nil
	}

	toMqttMsg := model.NewMessage(sourceMessage.Data)
	toMqttMsg.Timestamp = sourceMessage.Timestamp
	toMqttMsg.Others.Set(model.KeyMqttTopic, "", nil)
	return toMqttMsg, nil
}
