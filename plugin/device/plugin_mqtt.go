package plugin

import "github.com/mycontroller-org/2mqtt/plugin/device/mqtt"

func init() {
	Register(mqtt.PluginMQTT, mqtt.NewDevice)
}
