package model

const (
	ProviderMySensorsV2 = "mysensors_v2"
	ProviderRaw         = "raw"
)

// Formatter interface, used to convert the data
type Formatter interface {
	ToSourceMessage(mqttMessage *Message) (*Message, error)
	ToMQTTMessage(sourceMessage *Message) (*Message, error)
}

// Device can be a serial, mqtt, ethernet, etc.,
type Device interface {
	Close() error
	Write(message *Message) error
}
