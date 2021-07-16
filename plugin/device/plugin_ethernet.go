package plugin

import "github.com/mycontroller-org/2mqtt/plugin/device/ethernet"

func init() {
	Register(ethernet.PluginEthernet, ethernet.NewDevice)
}
