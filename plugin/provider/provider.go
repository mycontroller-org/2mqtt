package provider

import (
	"fmt"

	"github.com/mycontroller-org/2mqtt/pkg/model"
	config "github.com/mycontroller-org/2mqtt/pkg/model/config"
	mysensors "github.com/mycontroller-org/2mqtt/plugin/provider/mysensors_v2"
	raw "github.com/mycontroller-org/2mqtt/plugin/provider/raw"
)

func GetFormatter(adapterCfg *config.AdapterConfig) (model.Formatter, error) {
	switch adapterCfg.Provider {
	case model.ProviderMySensorsV2:
		return mysensors.Get(adapterCfg)

	case model.ProviderRaw:
		return raw.Get(adapterCfg)

	default:
		return nil, fmt.Errorf("unsupported provider type:%s", adapterCfg.Provider)
	}
}
