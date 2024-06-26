package raw

import (
	"context"
	"fmt"

	"github.com/mycontroller-org/2mqtt/pkg/types"
	cfgTY "github.com/mycontroller-org/2mqtt/pkg/types/config"
	providerType "github.com/mycontroller-org/2mqtt/plugin/provider/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

const PluginRaw = "raw"

func NewProvider(ctx context.Context, cfg cmap.CustomMap, formatter cfgTY.FormatterScript) (providerType.Plugin, error) {
	sourceType := cfg.GetString(types.KeyType)
	name := cfg.GetString(types.KeyName)

	switch sourceType {
	case types.DeviceSerial, types.DeviceEthernet, types.DeviceHTTP:
		return New(ctx, name, formatter)

	default:
		return nil, fmt.Errorf("unsupported source type:%s", sourceType)
	}
}
