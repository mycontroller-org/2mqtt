package mqtt

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	model "github.com/mycontroller-org/2mqtt/pkg/types"
	deviceType "github.com/mycontroller-org/2mqtt/plugin/device/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils"

	"go.uber.org/zap"
)

// Constants in serial device
const (
	PluginMQTT = "mqtt"

	transmitPreDelayDefault = time.Microsecond * 1 // 1 micro second
	reconnectDelayDefault   = time.Second * 10     // 10 seconds
)

// Config struct
type Config struct {
	Name               string `yaml:"name"`
	Broker             string `yaml:"broker"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	Username           string `yaml:"username"`
	Password           string `yaml:"password" json:"-"`
	Subscribe          string `yaml:"subscribe"`
	Publish            string `yaml:"publish"`
	QoS                int    `yaml:"qos"`
	TransmitPreDelay   string `yaml:"transmit_pre_delay"`
	ReconnectDelay     string `yaml:"reconnect_delay"`
}

// Endpoint data
type Endpoint struct {
	ID             string
	Config         *Config
	receiveMsgFunc func(msg *model.Message)
	statusFunc     func(state *model.State)
	Client         paho.Client
	txPreDelay     time.Duration
}

// NewDevice mqtt driver
func NewDevice(ID string, config cmap.CustomMap, rxFunc func(msg *model.Message), statusFunc func(state *model.State)) (deviceType.Plugin, error) {

	start := time.Now()

	var cfg Config
	err := utils.MapToStruct(utils.TagNameYaml, config, &cfg)
	if err != nil {
		return nil, err
	}
	zap.L().Debug("mqtt config", zap.Any("adapterName", ID), zap.Any("config", cfg))

	// endpoint
	endpoint := &Endpoint{
		ID:             ID,
		Config:         &cfg,
		receiveMsgFunc: rxFunc,
		statusFunc:     statusFunc,
		txPreDelay:     utils.ToDuration(cfg.TransmitPreDelay, transmitPreDelayDefault),
	}

	opts := paho.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetClientID(utils.RandID())
	opts.SetCleanSession(false)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetryInterval(utils.ToDuration(cfg.ReconnectDelay, reconnectDelayDefault))
	opts.SetOnConnectHandler(endpoint.onConnectionHandler)
	opts.SetConnectionLostHandler(endpoint.onConnectionLostHandler)

	// update tls config
	tlsConfig := &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}
	opts.SetTLSConfig(tlsConfig)

	c := paho.NewClient(opts)
	token := c.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		return nil, err
	}

	// adding client
	endpoint.Client = c

	zap.L().Debug("mqtt client connected successfully", zap.Any("adapterName", ID), zap.String("timeTaken", time.Since(start).String()), zap.Any("clientConfig", cfg))
	return endpoint, nil
}

func (ep *Endpoint) Name() string {
	return PluginMQTT
}

func (ep *Endpoint) onConnectionHandler(c paho.Client) {
	zap.L().Debug("mqtt connection success", zap.Any("adapterName", ep.ID))

	err := ep.Subscribe(ep.Config.Subscribe)
	if err != nil {
		zap.L().Error("error on subscribe topics", zap.Any("adapterName", ep.ID), zap.String("topics", ep.Config.Subscribe), zap.Error(err))
	}

	ep.statusFunc(&model.State{
		Status:  model.StatusUP,
		Message: "",
		Since:   time.Now(),
	})
}

func (ep *Endpoint) onConnectionLostHandler(c paho.Client, err error) {
	zap.L().Error("mqtt connection lost", zap.Any("adapterName", ep.ID), zap.Error(err))
	// Report connection lost
	if err != nil {
		ep.statusFunc(&model.State{
			Status:  model.StatusError,
			Message: err.Error(),
			Since:   time.Now(),
		})
	}
}

// Write publishes a payload
func (ep *Endpoint) Write(message *model.Message) error {
	if message == nil {
		return nil
	}
	zap.L().Debug("about to send a message", zap.Any("adapterName", ep.ID), zap.String("message", message.ToString()))
	topic := message.Others.GetString(model.KeyMqttTopic)
	qos := byte(ep.Config.QoS)

	time.Sleep(ep.txPreDelay) // transmit pre delay

	for _, rawtopic := range strings.Split(ep.Config.Publish, ",") {
		_topic := strings.TrimSpace(rawtopic)
		if topic != "" {
			_topic = fmt.Sprintf("%s/%s", _topic, topic)
		}
		token := ep.Client.Publish(_topic, qos, false, string(message.Data))
		if token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}

// Close the driver
func (ep *Endpoint) Close() error {
	if ep.Client.IsConnected() {
		ep.Client.Unsubscribe(ep.Config.Subscribe)
		ep.Client.Disconnect(0)
		zap.L().Debug("mqtt client connection closed", zap.String("adapterName", ep.ID))
	}
	return nil
}

func (ep *Endpoint) getCallBack() func(paho.Client, paho.Message) {
	return func(c paho.Client, message paho.Message) {
		rawMsg := model.NewMessage(message.Payload())
		rawMsg.Others.Set(model.KeyMqttTopic, message.Topic(), nil)
		rawMsg.Others.Set(model.KeyMqttQoS, int(message.Qos()), nil)
		ep.receiveMsgFunc(rawMsg)
	}
}

// Subscribe a topic
func (ep *Endpoint) Subscribe(topicsStr string) error {
	if topicsStr == "" {
		return nil
	}
	topics := strings.Split(topicsStr, ",")
	for _, topic := range topics {
		topic = strings.TrimSpace(topic)
		token := ep.Client.Subscribe(topic, 0, ep.getCallBack())
		token.WaitTimeout(3 * time.Second)
		if token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}
