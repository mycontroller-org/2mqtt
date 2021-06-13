package mysensors

import (
	"fmt"

	"github.com/mycontroller-org/2mqtt/pkg/model"
	config "github.com/mycontroller-org/2mqtt/pkg/model/config"
	serialDevice "github.com/mycontroller-org/2mqtt/plugin/device/serial"
)

func Get(adapterCfg *config.AdapterConfig) (model.Formatter, error) {
	sourceType := adapterCfg.Source.GetString(model.KeyType)

	switch sourceType {
	case model.DeviceSerial, model.DeviceEthernet:
		adapterCfg.Source.Set(serialDevice.KeyMessageSplitter, MessageSplitter, nil)
		return SourceType(adapterCfg.Name), nil

	default:
		return nil, fmt.Errorf("unsupported source type:%s", sourceType)
	}
}
