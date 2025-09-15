package config

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

const (
	KeyQueueConfig = "queue"
)

// Config
type Config struct {
	Logger   LoggerConfig    `yaml:"logger"`
	Adapters []AdapterConfig `yaml:"adapters"`
}

// LoggerConfig struct
type LoggerConfig struct {
	Mode             string `yaml:"mode"`
	Encoding         string `yaml:"encoding"`
	Level            string `yaml:"level"`
	EnableStacktrace bool   `yaml:"enable_stacktrace"`
}

// AdapterConfig struct
type AdapterConfig struct {
	Name            string          `yaml:"name"`
	Enabled         bool            `yaml:"enabled"`
	ReconnectDelay  string          `yaml:"reconnect_delay"`
	Provider        string          `yaml:"provider"`
	Source          cmap.CustomMap  `yaml:"source"`
	MQTT            cmap.CustomMap  `yaml:"mqtt"`
	FormatterScript FormatterScript `yaml:"formatter_script"`
}

// enter formatter script details, will be used along with raw provider
type FormatterScript struct {
	ToSource string `yaml:"to_source"`
	ToMQTT   string `yaml:"to_mqtt"`
}

// internal queue config for provider and source
type QueueConfig struct {
	QueueSize     int           `yaml:"size"`
	RetryEnabled  bool          `yaml:"retry_enabled"`
	RetryMaxCount uint32        `yaml:"retry_max_count"`
	RetryInterval time.Duration `yaml:"retry_interval"`
}
