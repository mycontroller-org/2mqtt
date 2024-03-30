package sub

import (
	"fmt"
	"log"

	cfgTY "github.com/mycontroller-org/2mqtt/pkg/types/config"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateConfigCmd)
}

var generateConfigCmd = &cobra.Command{
	Use:  "generate-config",
	Long: "Generates sample config file",
	Run: func(cmd *cobra.Command, args []string) {
		sampleConfig := cfgTY.Config{
			Logger: cfgTY.LoggerConfig{
				Mode:             "development",
				Encoding:         "console",
				Level:            "info",
				EnableStacktrace: false,
			},
			Adapters: []cfgTY.AdapterConfig{
				{
					Name:           "my_first_adapter",
					Enabled:        true,
					ReconnectDelay: "30s",
					Provider:       "raw",
					Source: cmap.CustomMap{
						"type":               "serial",
						"port":               "/dev/ttyUSB0",
						"baud_rate":          "115200",
						"transmit_pre_delay": "10ms",
					},
					MQTT: cmap.CustomMap{
						"broker":             "tcp://192.168.10.21:1883",
						"insecure":           "false",
						"username":           "",
						"password":           "",
						"subscribe":          "receive_data/#",
						"publish":            "publish_data",
						"qos":                "0",
						"transmit_pre_delay": "0s",
						"reconnect_delay":    "5s",
					},
					FormatterScript: cfgTY.FormatterScript{},
				},
			},
		}

		data, err := yaml.Marshal(&sampleConfig)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(data))
	},
}
