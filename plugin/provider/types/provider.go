package types

import "github.com/mycontroller-org/2mqtt/pkg/types"

// Formatter interface, used to convert the data
type Plugin interface {
	ToSourceMessage(mqttMessage *types.Message) (*types.Message, error)
	ToMQTTMessage(sourceMessage *types.Message) (*types.Message, error)
}
