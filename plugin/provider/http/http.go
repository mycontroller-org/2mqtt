package http

import (
	"fmt"

	types "github.com/mycontroller-org/2mqtt/pkg/types"
	providerType "github.com/mycontroller-org/2mqtt/plugin/provider/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

const PluginHTTP = "http"

func NewProvider(cfg cmap.CustomMap) (providerType.Plugin, error) {
	sourceType := cfg.GetString(types.KeyType)
	name := cfg.GetString(types.KeyName)

	switch sourceType {
	case types.DeviceHTTP:
		return SourceType(name), nil

	default:
		return nil, fmt.Errorf("unsupported source type:%s", sourceType)
	}
}
