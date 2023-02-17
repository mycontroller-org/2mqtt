package raw

import (
	"fmt"
	"testing"
	"time"

	"github.com/mycontroller-org/2mqtt/pkg/types"
	"github.com/mycontroller-org/2mqtt/pkg/types/config"
	"github.com/stretchr/testify/assert"
)

func TestFormatterScript(t *testing.T) {
	timeStamp := time.Now()
	tests := []struct {
		testName       string
		toSourceInput  *types.Message
		toSourceOutput *types.Message
		toMqttInput    *types.Message
		toMqttOutput   *types.Message
		formatter      config.FormatterScript
	}{
		{
			testName:       "TestScriptMapReturn",
			toSourceInput:  &types.Message{Data: []byte("hello"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toSourceOutput: &types.Message{Data: []byte("hello_modified"), Others: map[string]interface{}{"mqtt_topic": "hello_serial"}, Timestamp: timeStamp},
			toMqttInput:    &types.Message{Data: []byte("hello"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toMqttOutput:   &types.Message{Data: []byte("hello_modified"), Others: map[string]interface{}{"mqtt_topic": "hello_mqtt"}, Timestamp: timeStamp},
			formatter: config.FormatterScript{
				ToSource: `
				result={
					data: raw_data + "_modified",
					mqtt_topic: "hello_serial",
				}
			`,
				ToMQTT: `
				result={
					data: raw_data + "_modified",
					mqtt_topic: "hello_mqtt",
				}
			`,
			},
		},
		{
			testName:       "TestScriptStringReturn",
			toSourceInput:  &types.Message{Data: []byte("hello"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toSourceOutput: &types.Message{Data: []byte("hello_modified_string"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toMqttInput:    &types.Message{Data: []byte("hello"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toMqttOutput:   &types.Message{Data: []byte("hello_modified_string"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			formatter: config.FormatterScript{
				ToSource: `
				result=raw_data + "_modified_string"
			`,
				ToMQTT: `
				result=raw_data + "_modified_string"
			`,
			},
		},
		{
			testName:       "TestNoScript",
			toSourceInput:  &types.Message{Data: []byte("hello"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toSourceOutput: &types.Message{Data: []byte("hello"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toMqttInput:    &types.Message{Data: []byte("hello"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toMqttOutput:   &types.Message{Data: []byte("hello"), Others: map[string]interface{}{"mqtt_topic": ""}, Timestamp: timeStamp},
			formatter:      config.FormatterScript{},
		},
		{
			testName:       "TestScriptOnlyToSource",
			toSourceInput:  &types.Message{Data: []byte("hello"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toSourceOutput: &types.Message{Data: []byte("hello_modified_string"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toMqttInput:    &types.Message{Data: []byte("hello"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toMqttOutput:   &types.Message{Data: []byte("hello"), Others: map[string]interface{}{"mqtt_topic": ""}, Timestamp: timeStamp},
			formatter: config.FormatterScript{
				ToSource: `
				result=raw_data + "_modified_string"
				`,
			},
		},
		{
			testName:       "TestScriptOnlyToMqtt",
			toSourceInput:  &types.Message{Data: []byte("hello"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toSourceOutput: &types.Message{Data: []byte("hello"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toMqttInput:    &types.Message{Data: []byte("hello"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toMqttOutput:   &types.Message{Data: []byte("hello_modified_string"), Others: map[string]interface{}{}, Timestamp: timeStamp},
			formatter: config.FormatterScript{
				ToMQTT: `
				result=raw_data + "_modified_string"
				`,
			},
		},
		{
			testName: "TestScriptJson",
			toSourceInput: &types.Message{
				Data:      []byte("3.141592653"),
				Others:    map[string]interface{}{"mqtt_topic": "myTopicPi"},
				Timestamp: timeStamp,
			},
			toSourceOutput: &types.Message{Data: []byte(`[1,"myTopicPi","3.141592653"]`), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toMqttInput:    &types.Message{Data: []byte(`[1,"myTopicPi","3.141592653","1","0", "0xa0"]\r`), Others: map[string]interface{}{}, Timestamp: timeStamp},
			toMqttOutput: &types.Message{
				Data:      []byte("3.141592653"),
				Others:    map[string]interface{}{"command": int64(1), "crc": "0xa0", "mqtt_qos": "1", "mqtt_retain": "0", "mqtt_topic": "myTopicPi"},
				Timestamp: timeStamp,
			},
			formatter: config.FormatterScript{
				ToSource: `
				// [<COMMAND>,<TOPIC>,<MESSAGE>]
				result={
					data: JSON.stringify([1, mqtt_topic, raw_data])
				}
				`,
				ToMQTT: `
				let update_raw_data = raw_data
				if (update_raw_data.endsWith("\\r")){
					update_raw_data = update_raw_data.slice(0, -2)
				}
				console.log(update_raw_data)
				let serialData = JSON.parse(update_raw_data)
				// [<COMMAND>,<TOPIC>,<MESSAGE>,<QOS>,<RETAIN>,<CRC>]
				// [1,"myTopicPi","3.141592653"]

				let keyMap = ["command", "mqtt_topic", "data", "mqtt_qos", "mqtt_retain", "crc"]
				let serialDataMap = {}
				for (let index=0; index<serialData.length;index++) {
					serialDataMap[keyMap[index]] = serialData[index]
				}

				result={
					...serialDataMap
				}
				`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			rawProvider := RawProvider{
				name:      fmt.Sprint("my_adapter_", test.testName),
				formatter: test.formatter,
			}

			// verify ToSource
			toSourceOut, err := rawProvider.ToSourceMessage(test.toSourceInput)
			assert.NoError(t, err)
			assert.Equal(t, test.toSourceOutput, toSourceOut)

			// verify ToMqtt
			toMqttOut, err := rawProvider.ToMQTTMessage(test.toMqttInput)
			assert.NoError(t, err)
			assert.Equal(t, test.toMqttOutput, toMqttOut)
		})
	}

}
