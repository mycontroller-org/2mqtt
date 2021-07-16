package mysensors

import (
	"fmt"

	"github.com/mycontroller-org/2mqtt/pkg/model"
	providerType "github.com/mycontroller-org/2mqtt/plugin/provider/types"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

const PluginMySensors = "mysensors_v2"

func NewProvider(config cmap.CustomMap) (providerType.Plugin, error) {
	sourceType := config.GetString(model.KeyType)
	name := config.GetString(model.KeyName)

	switch sourceType {
	case model.DeviceSerial, model.DeviceEthernet:
		config.Set(model.KeyMessageSplitter, MessageSplitter, nil)
		return SourceType(name), nil

	default:
		return nil, fmt.Errorf("unsupported source type:%s", sourceType)
	}
}
