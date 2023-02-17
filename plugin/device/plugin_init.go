package plugin

import (
	"github.com/mycontroller-org/2mqtt/plugin/device/ethernet"
	httpDevice "github.com/mycontroller-org/2mqtt/plugin/device/http"
	"github.com/mycontroller-org/2mqtt/plugin/device/serial"
)

func init() {
	Register(serial.PluginSerial, serial.NewDevice)
	Register(httpDevice.PluginHTTP, httpDevice.NewDevice)
	Register(ethernet.PluginEthernet, ethernet.NewDevice)
}
