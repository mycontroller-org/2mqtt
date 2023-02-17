package mysensors

import (
	"errors"
	"strings"

	"github.com/mycontroller-org/2mqtt/pkg/types"
	"go.uber.org/zap"
)

const (
	MessageSplitter = '\n'
)

type SourceType string

func (st SourceType) Name() string {
	return PluginMySensors
}

func (st SourceType) ToSourceMessage(mqttMessage *types.Message) (*types.Message, error) {
	// node-id;child-sensor-id;command;ack;type;payload\n
	topic := mqttMessage.Others.GetString(types.KeyMqttTopic)
	topicSlice := strings.Split(topic, "/")
	if len(topicSlice) < 5 {
		zap.L().Warn("invalid topic", zap.Any("message", mqttMessage))
		return nil, errors.New("invalid topic")
	}
	topicSlice = topicSlice[len(topicSlice)-5:]

	payload := ""
	if len(mqttMessage.Data) > 0 {
		payload = string(mqttMessage.Data)
	}

	topicSlice = append(topicSlice, payload)

	finalData := strings.Join(topicSlice[:], ";")

	formattedMessage := &types.Message{
		Data:      []byte(finalData),
		Others:    mqttMessage.Others,
		Timestamp: mqttMessage.Timestamp,
	}
	return formattedMessage, nil
}

func (st SourceType) ToMQTTMessage(sourceMessage *types.Message) (*types.Message, error) {
	// structure: node-id/child-sensor-id/command/ack/type payload
	data := ""
	if len(sourceMessage.Data) > 0 {
		data = string(sourceMessage.Data)
	}
	dataSlice := strings.Split(data, ";")
	if len(dataSlice) != 6 {
		zap.L().Warn("invalid message format", zap.String("message", data))
		return nil, errors.New("invalid message format")
	}
	topic := strings.Join(dataSlice[:5], "/")
	payload := dataSlice[5]

	formattedMessage := types.NewMessage([]byte(payload))
	formattedMessage.Timestamp = sourceMessage.Timestamp
	formattedMessage.Others.Set(types.KeyMqttTopic, topic, nil)

	return formattedMessage, nil
}
