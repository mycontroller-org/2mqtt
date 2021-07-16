package plugin

import (
	raw "github.com/mycontroller-org/2mqtt/plugin/provider/raw"
)

func init() {
	Register(raw.PluginRaw, raw.NewProvider)
}
