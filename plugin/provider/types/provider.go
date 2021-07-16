package types

import "github.com/mycontroller-org/2mqtt/pkg/model"

// Formatter interface, used to convert the data
type Plugin interface {
	ToSourceMessage(mqttMessage *model.Message) (*model.Message, error)
	ToMQTTMessage(sourceMessage *model.Message) (*model.Message, error)
}
