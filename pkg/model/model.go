package model

import (
	"time"
)

const (
	// device types
	DeviceEthernet = "ethernet"
	DeviceSerial   = "serial"
	DeviceMQTT     = "mqtt"

	// keys used across
	KeyType            = "type"
	KeyMqttTopic       = "mqtt_topic"
	KeyMqttQoS         = "mqtt_qos"
	KeyMessageSplitter = "message_splitter"

	// Status
	StatusUP    = "up"
	StatusError = "error"
)

// State struct
type State struct {
	Status  string
	Message string
	Since   time.Time
}
