# 2mqtt
![lint workflow](https://github.com/mycontroller-org/2mqtt/actions/workflows/lint.yaml/badge.svg)
![publish container images](https://github.com/mycontroller-org/2mqtt/actions/workflows/publish_container_images.yaml/badge.svg)
![publish executables](https://github.com/mycontroller-org/2mqtt/actions/workflows/publish_executables.yaml/badge.svg)

2mqtt is MQTT bridge. You can convert the serial, ethernet to MQTT

### Supported Providers
* [MySensors](https://www.mysensors.org/)
  * `serial` to `MQTT`
  * `ethernet` to `MQTT`
* `raw` - sends the serial,ethernet messages to mqtt as is
  * `serial` to `MQTT`
  * `ethernet` to `MQTT`

## Download
### Container images
`master` branch images are tagged as `:master`<br>
Both released and master branch container images are published in to the following registries,
  * [Docker Hub](https://hub.docker.com/r/mycontroller/2mqtt)
  * [Quay.io](https://quay.io/repository/mycontroller/2mqtt)
#### Docker Run
```bash
docker run --detach --name 2mqtt \
    --volume $PWD/config.yaml:/app/config.yaml \
    --device /dev/ttyUSB0:/dev/ttyUSB0 \
    --env  TZ="Asia/Kolkata" \
    --restart unless-stopped \
    docker.io/mycontroller/2mqtt:master
```

### Executables
* [Released versions](https://github.com/mycontroller-org/2mqtt/releases)
* [Pre Release](https://download.mycontroller.org/2mqtt/master/) - `master` branch executables


## Configuration
You can have more than one adapter configurations

Provider options: `mysensors_v2` and `raw`
```yaml
logger:
  mode: development   # logger mode: development, production
  encoding: console   # encoding options: console, json
  level: info         # log levels: debug, info, warn, error, fatal

adapters:   # you can have more than one adapter
  - name: adapter1          # name of the adapter
    enabled: false          # enable or disable the adapter, default disabled
    reconnect_delay: 20s    # reconnect automatically, if there is a failure on the connection
    provider: mysensors_v2  # provider type, options: mysensors_v2, raw
    source: # source is the device, to be converted to MQTT, based on the type, configurations will be different
      type: serial              # source device type: serial
      port: /dev/ttyUSB0        # serial port
      baud_rate: 115200         # serial baud rate
      transmit_pre_delay: 10ms  # waits and sends a message, to avoid collision on the source network
    mqtt: # mqtt broker details
      broker: tcp://192.168.10.21:1883  # broker url: supports tcp, mqtt, tls, mqtts
      insecure_skip_verify: false       # enable/disable insecure on tls connection
      username:                         # username of the broker
      password:                         # password of the broker
      subscribe: in_rfm69/#             # subscribe a topic, should include `#` at the end, your controller to serial port(source)
      publish: out_rfm69                # publish on this topic, can add many topics with comma, serial to your controller
      qos: 0                            # qos number: 0, 1, 2
      transmit_pre_delay: 0s
      reconnect_delay: 5s
```
### Source device configuration
Based on the source type the configurations will be different.
#### Serial
```yaml
source:
  type: serial              # source device type
  port: /dev/ttyUSB0        # serial port
  baud_rate: 115200         # serial baud rate
  transmit_pre_delay: 10ms  # waits and sends a message, to avoid collision on the source network
  message_splitter: # message splitter byte, default '10'
```
#### Ethernet
```yaml
source:
  type: ethernet                    # source device type
  server: tcp://192.168.10.21:5003  # ethernet server address with port
  transmit_pre_delay: 10ms          # waits and sends a message, to avoid collision on the source network
  message_splitter: # message splitter byte, default '10'
```

### Special note on message_splitter
* `message_splitter` is a reference char to understand the end of message on serial and ethernet device read<br>
* This special char will be included while writing to the device.
* supports only one char, should be supplied in byte, ie:`0` to `255`, extended ASCII chars

#### Quick references
For complete details refer the extended ASCII table 
* `0` - Null char
* `3` - End of Text
* `4` - End of Transmission
* `8` - Back Space
* `10` - Line Feed
* `13` - Carriage Return