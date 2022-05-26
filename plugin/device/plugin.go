package plugin

import (
	"fmt"

	"github.com/mycontroller-org/2mqtt/pkg/types"
	deviceType "github.com/mycontroller-org/2mqtt/plugin/device/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"go.uber.org/zap"
)

// CreatorFn func type
type CreatorFn func(ID string, config cmap.CustomMap, rxFunc func(msg *model.Message), statusFunc func(state *model.State)) (deviceType.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	if _, found := creators[name]; found {
		zap.L().Fatal("duplicate plugin found", zap.String("pluginName", name))
		return
	}
	creators[name] = fn
}

func Create(name, ID string, config cmap.CustomMap, rxFunc func(msg *model.Message), statusFunc func(state *model.State)) (p deviceType.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn("", config, rxFunc, statusFunc)
	} else {
		err = fmt.Errorf("device plugin [%s] is not registered", name)
	}
	return
}
