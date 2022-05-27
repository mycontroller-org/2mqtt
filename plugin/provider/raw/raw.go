package raw

import (
	"fmt"

	model "github.com/mycontroller-org/2mqtt/pkg/types"
	providerType "github.com/mycontroller-org/2mqtt/plugin/provider/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

const PluginRaw = "raw"

func NewProvider(cfg cmap.CustomMap) (providerType.Plugin, error) {
	sourceType := cfg.GetString(model.KeyType)
	name := cfg.GetString(model.KeyName)

	switch sourceType {
	case model.DeviceSerial, model.DeviceEthernet, model.DeviceHTTP:
		return SourceType(name), nil

	default:
		return nil, fmt.Errorf("unsupported source type:%s", sourceType)
	}
}
