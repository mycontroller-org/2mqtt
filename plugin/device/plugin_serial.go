package plugin

import "github.com/mycontroller-org/2mqtt/plugin/device/serial"

func init() {
	Register(serial.PluginSerial, serial.NewDevice)
}
