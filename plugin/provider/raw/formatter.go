package raw

import (
	"github.com/mycontroller-org/2mqtt/pkg/model"
)

type SourceType string

func (st SourceType) ToSourceMessage(mqttMessage *model.Message) (*model.Message, error) {
	if len(mqttMessage.Data) == 0 {
		return nil, nil
	}

	toSourceMsg := &model.Message{
		Data:      mqttMessage.Data,
		Others:    mqttMessage.Others,
		Timestamp: mqttMessage.Timestamp,
	}
	return toSourceMsg, nil
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
