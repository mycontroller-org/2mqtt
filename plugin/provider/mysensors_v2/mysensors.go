package mysensors

import (
	"context"
	"fmt"

	"github.com/mycontroller-org/2mqtt/pkg/types"
	cfgTY "github.com/mycontroller-org/2mqtt/pkg/types/config"
	providerType "github.com/mycontroller-org/2mqtt/plugin/provider/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

const PluginMySensors = "mysensors_v2"

func NewProvider(ctx context.Context, config cmap.CustomMap, formatter cfgTY.FormatterScript) (providerType.Plugin, error) {
	sourceType := config.GetString(types.KeyType)
	name := config.GetString(types.KeyName)

	switch sourceType {
	case types.DeviceSerial, types.DeviceEthernet:
		config.Set(types.KeyMessageSplitter, MessageSplitter, nil)
		return New(ctx, name)

	default:
		return nil, fmt.Errorf("unsupported source type:%s", sourceType)
	}
}
