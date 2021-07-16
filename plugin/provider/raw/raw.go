package raw

import (
	"fmt"

	"github.com/mycontroller-org/2mqtt/pkg/model"
	providerType "github.com/mycontroller-org/2mqtt/plugin/provider/types"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

const PluginRaw = "raw"

func NewProvider(cfg cmap.CustomMap) (providerType.Plugin, error) {
	sourceType := cfg.GetString(model.KeyType)
	name := cfg.GetString(model.KeyName)

	switch sourceType {
	case model.DeviceSerial, model.DeviceEthernet:
		return SourceType(name), nil

	default:
		return nil, fmt.Errorf("unsupported source type:%s", sourceType)
	}
}
