package mysensors

import (
	"fmt"

	"github.com/mycontroller-org/2mqtt/pkg/model"
	config "github.com/mycontroller-org/2mqtt/pkg/model/config"
	serialDevice "github.com/mycontroller-org/2mqtt/plugin/device/serial"
	ethernetFormatter "github.com/mycontroller-org/2mqtt/plugin/provider/mysensors_v2/ethernet"
	serialFormatter "github.com/mycontroller-org/2mqtt/plugin/provider/mysensors_v2/serial"
)

func Get(adapterCfg *config.AdapterConfig) (model.Formatter, error) {
	sourceType := adapterCfg.Source.GetString(model.KeyType)

	switch sourceType {
	case model.DeviceSerial:
		adapterCfg.Source.Set(serialDevice.KeyMessageSplitter, serialFormatter.MessageSplitter, nil)
		return serialFormatter.SourceType(adapterCfg.Name), nil

	case model.DeviceEthernet:
		return ethernetFormatter.SourceType(adapterCfg.Name), nil

	default:
		return nil, fmt.Errorf("unsupported source type:%s", sourceType)
	}
}
