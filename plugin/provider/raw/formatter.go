package raw

import (
	"fmt"

	"github.com/mycontroller-org/2mqtt/pkg/types"
	cfgTY "github.com/mycontroller-org/2mqtt/pkg/types/config"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	js "github.com/mycontroller-org/server/v2/pkg/utils/javascript"
)

const (
	KeyRawData = "raw_data"
	KeyData    = "data"
	KeyOthers  = "others"
)

type RawProvider struct {
	name      string
	formatter cfgTY.FormatterScript
}

func (rp *RawProvider) Name() string {
	return PluginRaw
}

func (rp *RawProvider) ToSourceMessage(mqttMessage *types.Message) (*types.Message, error) {
	toSourceMsg := &types.Message{
		Timestamp: mqttMessage.Timestamp,
	}

	// if script available execute it
	if rp.formatter.ToSource != "" {
		outMsg, err := rp.executeScript(rp.formatter.ToSource, mqttMessage)
		if err != nil {
			return nil, err
		}
		toSourceMsg = outMsg
	} else if len(mqttMessage.Data) == 0 {
		// if script not available and no data received and return nil and no error
		return nil, nil
	} else {
		toSourceMsg.Data = mqttMessage.Data
		toSourceMsg.Others = mqttMessage.Others
	}

	return toSourceMsg, nil
}

func (rp *RawProvider) ToMQTTMessage(sourceMessage *types.Message) (*types.Message, error) {
	if len(sourceMessage.Data) == 0 {
		return nil, nil
	}
	toMqttMsg := types.NewMessage(sourceMessage.Data)
	toMqttMsg.Timestamp = sourceMessage.Timestamp
	toMqttMsg.Others.Set(types.KeyMqttTopic, "", nil)

	// if script available execute it
	if rp.formatter.ToMQTT != "" {
		outMsg, err := rp.executeScript(rp.formatter.ToMQTT, sourceMessage)
		if err != nil {
			return nil, err
		}
		toMqttMsg = outMsg
	}

	return toMqttMsg, nil
}

func (rp *RawProvider) executeScript(script string, msg *types.Message) (*types.Message, error) {
	outMsg := types.Message{
		Timestamp: msg.Timestamp,
		Others:    make(cmap.CustomMap),
	}
	// form input map
	input := map[string]interface{}{
		KeyRawData: "",
	}
	// load raw data
	if len(string(msg.Data)) > 0 {
		input[KeyRawData] = string(msg.Data)
	}
	// load data from others
	for key, value := range msg.Others {
		input[key] = value
	}

	// executes script
	response, err := js.Execute(script, input)
	if err != nil {
		return nil, err
	}
	// gat response from script output
	if stringData, ok := response.(string); ok {
		outMsg.Data = []byte(stringData)
	} else if mapData, ok := response.(map[string]interface{}); ok {
		foundData := false
		// update received data
		for key, value := range mapData {
			if key == KeyData {
				foundData = true
				outMsg.Data = []byte(fmt.Sprintf("%v", value))
				continue
			}
			outMsg.Others[key] = value
		}

		if !foundData {
			return nil, fmt.Errorf("adapterName: %s, key '%s' is not found on the script response[%+v], rawData:[%+v]", rp.name, KeyData, mapData, msg)
		}
	}

	return &outMsg, nil
}
