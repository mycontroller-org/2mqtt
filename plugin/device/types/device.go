package device

import (
	model "github.com/mycontroller-org/2mqtt/pkg/model"
)

// Device can be a serial, mqtt, ethernet, etc.,
type Plugin interface {
	Name() string
	Close() error
	Write(message *model.Message) error
}
