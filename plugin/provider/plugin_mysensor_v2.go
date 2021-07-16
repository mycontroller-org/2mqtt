package plugin

import (
	mysensorsV2 "github.com/mycontroller-org/2mqtt/plugin/provider/mysensors_v2"
)

func init() {
	Register(mysensorsV2.PluginMySensors, mysensorsV2.NewProvider)
}
