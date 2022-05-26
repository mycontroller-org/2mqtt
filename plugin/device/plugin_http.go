package plugin

import (
	httpDevice "github.com/mycontroller-org/2mqtt/plugin/device/http"
)

func init() {
	Register(httpDevice.PluginHTTP, httpDevice.NewDevice)
}
