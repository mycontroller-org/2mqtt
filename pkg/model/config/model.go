package model

import "github.com/mycontroller-org/backend/v2/pkg/model/cmap"

// Config
type Config struct {
	Logger   LoggerConfig    `yaml:"logger"`
	Adapters []AdapterConfig `yaml:"adapters"`
}

// LoggerConfig struct
type LoggerConfig struct {
	Mode     string `yaml:"mode"`
	Encoding string `yaml:"encoding"`
	Level    string `yaml:"level"`
}

// AdapterConfig struct
type AdapterConfig struct {
	Name           string         `yaml:"name"`
	Enabled        bool           `yaml:"enabled"`
	ReconnectDelay string         `yaml:"reconnect_delay"`
	Provider       string         `yaml:"provider"`
	Source         cmap.CustomMap `yaml:"source"`
	MQTT           MQTTConfig     `yaml:"mqtt"`
}

// MQTTConfig struct
type MQTTConfig struct {
	Broker             string `yaml:"broker"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	Username           string `yaml:"username"`
	Password           string `yaml:"password" json:"-"`
	Subscribe          string `yaml:"subscribe"`
	Publish            string `yaml:"publish"`
	QoS                int    `yaml:"qos"`
	TransmitPreDelay   string `yaml:"transmit_pre_delay"`
	ReconnectDelay     string `yaml:"reconnect_delay"`
}
