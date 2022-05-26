package model

import (
	"time"
)

const (
	// device types
	DeviceEthernet = "ethernet"
	DeviceSerial   = "serial"
	DeviceMQTT     = "mqtt"
	DeviceHTTP     = "http"

	// keys used across
	KeyType            = "type"
	KeyName            = "name"
	KeyMqttTopic       = "mqtt_topic"
	KeyMqttQoS         = "mqtt_qos"
	KeyMessageSplitter = "message_splitter"
	KeyHeaders         = "headers"
	KeyURL             = "url"

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
