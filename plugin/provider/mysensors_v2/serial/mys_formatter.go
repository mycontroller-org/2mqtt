package serial

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mycontroller-org/2mqtt/pkg/model"
	"go.uber.org/zap"
)

const (
	MessageSplitter = '\n'
)

type SourceType string

func (st SourceType) ToSourceMessage(mqttMessage *model.Message) (*model.Message, error) {
	// node-id;child-sensor-id;command;ack;type;payload/n
	topic := mqttMessage.Others.GetString(model.KeyMqttTopic)
	topicSlice := strings.Split(topic, "/")
	if len(topicSlice) < 5 {
		zap.L().Warn("invalid topic", zap.Any("message", mqttMessage))
		return nil, errors.New("invalid topic")
	}
	topicSlice = topicSlice[len(topicSlice)-5:]

	payload := ""
	if payloadBytes, ok := mqttMessage.Data.([]byte); ok {
		payload = string(payloadBytes)
	}

	topicSlice = append(topicSlice, payload)

	finalData := strings.Join(topicSlice[:], ";")

	formattedMessage := &model.Message{
		Data:      fmt.Sprintf("%s%c", finalData, MessageSplitter),
		Others:    mqttMessage.Others,
		Timestamp: mqttMessage.Timestamp,
	}
	return formattedMessage, nil
}

func (st SourceType) ToMQTTMessage(sourceMessage *model.Message) (*model.Message, error) {
	// structure: node-id/child-sensor-id/command/ack/type payload
	data := ""
	if sourceMessage.Data != nil {
		if byteData, ok := sourceMessage.Data.([]byte); ok {
			data = string(byteData)
		}
	}
	dataSlice := strings.Split(data, ";")
	if len(dataSlice) != 6 {
		zap.L().Warn("invalid message format", zap.String("message", data))
		return nil, errors.New("invalid message format")
	}
	topic := strings.Join(dataSlice[:5], "/")
	payload := dataSlice[5]

	formattedMessage := model.NewMessage([]byte(payload))
	formattedMessage.Timestamp = sourceMessage.Timestamp
	formattedMessage.Others.Set(model.KeyMqttTopic, topic, nil)

	return formattedMessage, nil
}
