package raw

import (
	"fmt"

	"github.com/mycontroller-org/2mqtt/pkg/model"
	config "github.com/mycontroller-org/2mqtt/pkg/model/config"
)

func Get(adapterCfg *config.AdapterConfig) (model.Formatter, error) {
	sourceType := adapterCfg.Source.GetString(model.KeyType)

	switch sourceType {
	case model.DeviceSerial, model.DeviceEthernet:
		return SourceType(adapterCfg.Name), nil

	default:
		return nil, fmt.Errorf("unsupported source type:%s", sourceType)
	}
}
