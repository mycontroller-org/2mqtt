package device

import (
	"fmt"

	"github.com/mycontroller-org/2mqtt/pkg/model"
	config "github.com/mycontroller-org/2mqtt/pkg/model/config"
	ethernet "github.com/mycontroller-org/2mqtt/plugin/device/ethernet"
	mqtt "github.com/mycontroller-org/2mqtt/plugin/device/mqtt"
	serial "github.com/mycontroller-org/2mqtt/plugin/device/serial"
)

func GetSourceDevice(adapterCfg *config.AdapterConfig, rxFunc func(msg *model.Message), statusFunc func(state *model.State)) (model.Device, error) {
	deviceType := adapterCfg.Source.GetString(model.KeyType)

	switch deviceType {
	case model.DeviceSerial:
		return serial.New(adapterCfg.Name, adapterCfg.Source, rxFunc, statusFunc)

	case model.DeviceEthernet:
		return ethernet.New(adapterCfg.Name, adapterCfg.Source, rxFunc, statusFunc)

	}
	return nil, fmt.Errorf("unsupported source device:%s", deviceType)
}

func GetMQTTDevice(adapterCfg *config.AdapterConfig, rxFunc func(msg *model.Message), statusFunc func(state *model.State)) (model.Device, error) {
	return mqtt.New(adapterCfg.Name, &adapterCfg.MQTT, rxFunc, statusFunc)
}
