package plugin

import (
	"context"
	"fmt"

	cfgTY "github.com/mycontroller-org/2mqtt/pkg/types/config"
	providerType "github.com/mycontroller-org/2mqtt/plugin/provider/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"go.uber.org/zap"
)

// CreatorFn func type
type CreatorFn func(ctx context.Context, config cmap.CustomMap, formatter cfgTY.FormatterScript) (providerType.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	if _, found := creators[name]; found {
		zap.L().Fatal("duplicate plugin found", zap.String("pluginName", name))
		return
	}
	creators[name] = fn
}

func Create(ctx context.Context, name string, config cmap.CustomMap, formatter cfgTY.FormatterScript) (p providerType.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(ctx, config, formatter)
	} else {
		err = fmt.Errorf("device plugin [%s] is not registered", name)
	}
	return
}
