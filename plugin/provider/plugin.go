package plugin

import (
	"fmt"

	providerType "github.com/mycontroller-org/2mqtt/plugin/provider/types"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	"go.uber.org/zap"
)

// CreatorFn func type
type CreatorFn func(config cmap.CustomMap) (providerType.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	if _, found := creators[name]; found {
		zap.L().Fatal("duplicate plugin found", zap.String("pluginName", name))
		return
	}
	creators[name] = fn
}

func Create(name string, config cmap.CustomMap) (p providerType.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(config)
	} else {
		err = fmt.Errorf("device plugin [%s] is not registered", name)
	}
	return
}
