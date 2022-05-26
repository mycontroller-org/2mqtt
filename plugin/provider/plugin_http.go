package plugin

import (
	httpProvider "github.com/mycontroller-org/2mqtt/plugin/provider/http"
)

func init() {
	Register(httpProvider.PluginHTTP, httpProvider.NewProvider)
}
