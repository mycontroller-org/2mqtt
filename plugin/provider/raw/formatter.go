package raw

import (
	"github.com/mycontroller-org/2mqtt/pkg/types"
)

type SourceType string

func (st SourceType) Name() string {
	return PluginRaw
}

func (st SourceType) ToSourceMessage(mqttMessage *types.Message) (*types.Message, error) {
	if len(mqttMessage.Data) == 0 {
		return nil, nil
	}

	toSourceMsg := &types.Message{
		Data:      mqttMessage.Data,
		Others:    mqttMessage.Others,
		Timestamp: mqttMessage.Timestamp,
	}
	return toSourceMsg, nil
}

func (st SourceType) ToMQTTMessage(sourceMessage *types.Message) (*types.Message, error) {
	if len(sourceMessage.Data) == 0 {
		return nil, nil
	}

	toMqttMsg := types.NewMessage(sourceMessage.Data)
	toMqttMsg.Timestamp = sourceMessage.Timestamp
	toMqttMsg.Others.Set(types.KeyMqttTopic, "", nil)
	return toMqttMsg, nil
}
