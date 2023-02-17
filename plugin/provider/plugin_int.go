package plugin

import (
	mysensorsV2 "github.com/mycontroller-org/2mqtt/plugin/provider/mysensors_v2"
	raw "github.com/mycontroller-org/2mqtt/plugin/provider/raw"
)

func init() {
	Register(raw.PluginRaw, raw.NewProvider)
	Register(mysensorsV2.PluginMySensors, mysensorsV2.NewProvider)
}
