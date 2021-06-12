package ethernet

import "github.com/mycontroller-org/2mqtt/pkg/model"

type SourceType string

func (st SourceType) ToSourceMessage(mqttMessage *model.Message) (*model.Message, error) {
	return nil, nil
}

func (st SourceType) ToMQTTMessage(sourceMessage *model.Message) (*model.Message, error) {
	return nil, nil
}
