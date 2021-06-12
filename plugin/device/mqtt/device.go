package mqtt

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/mycontroller-org/2mqtt/pkg/model"
	config "github.com/mycontroller-org/2mqtt/pkg/model/config"
	"github.com/mycontroller-org/backend/v2/pkg/utils"

	"go.uber.org/zap"
)

// Constants in serial device
const (
	transmitPreDelayDefault = time.Microsecond * 1 // 1 micro second
	reconnectDelayDefault   = time.Second * 10     // 10 seconds
)

// Endpoint data
type Endpoint struct {
	ID             string
	Config         *config.MQTTConfig
	receiveMsgFunc func(msg *model.Message)
	statusFunc     func(state *model.State)
	Client         paho.Client
	txPreDelay     time.Duration
}

// New mqtt driver
func New(ID string, cfg *config.MQTTConfig, rxFunc func(msg *model.Message), statusFunc func(state *model.State)) (*Endpoint, error) {
	zap.L().Debug("mqtt config", zap.String("id", ID), zap.Any("config", cfg))

	start := time.Now()

	// endpoint
	endpoint := &Endpoint{
		ID:             ID,
		Config:         cfg,
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

	err := endpoint.Subscribe(cfg.Subscribe)
	if err != nil {
		zap.L().Error("error on subscribe a topic", zap.String("topic", cfg.Subscribe), zap.Error(err))
	}
	zap.L().Debug("mqtt client connected successfully", zap.String("timeTaken", time.Since(start).String()), zap.Any("clientConfig", cfg))
	return endpoint, nil
}

func (ep *Endpoint) onConnectionHandler(c paho.Client) {
	zap.L().Debug("mqtt connection success", zap.Any("adapterName", ep.ID))
	ep.statusFunc(&model.State{
		Status:  model.StatusUP,
		Message: "",
		Since:   time.Now(),
	})
}

func (ep *Endpoint) onConnectionLostHandler(c paho.Client, err error) {
	zap.L().Error("mqtt connection lost", zap.Any("id", ep.ID), zap.Error(err))
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
	zap.L().Debug("about to send a message", zap.String("message", message.ToString()))
	topic := message.Others.GetString(model.KeyMqttTopic)
	qos := byte(ep.Config.QoS)

	time.Sleep(ep.txPreDelay) // transmit pre delay

	for _, rawtopic := range strings.Split(ep.Config.Publish, ",") {
		_topic := fmt.Sprintf("%s/%s", strings.TrimSpace(rawtopic), topic)
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
func (ep *Endpoint) Subscribe(topic string) error {
	token := ep.Client.Subscribe(topic, 0, ep.getCallBack())
	token.WaitTimeout(3 * time.Second)
	if token.Error() != nil {
		ep.statusFunc(&model.State{
			Status:  model.StatusError,
			Message: token.Error().Error(),
			Since:   time.Now(),
		})
	}
	return token.Error()
}
