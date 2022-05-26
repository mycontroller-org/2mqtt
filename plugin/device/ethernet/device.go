package ethernet

import (
	"net"
	"net/url"
	"time"

	"github.com/mycontroller-org/2mqtt/pkg/types"
	deviceType "github.com/mycontroller-org/2mqtt/plugin/device/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"go.uber.org/zap"
)

// Constants in ethernet protocol
const (
	PluginEthernet = "ethernet"

	MaxDataLength           = 1000
	transmitPreDelayDefault = time.Microsecond * 1 // 1 microsecond

	DefaultMessageSplitter = '\n'
)

// Config details
type Config struct {
	Server           string `yaml:"server"`
	MessageSplitter  *byte  `yaml:"message_splitter"`
	TransmitPreDelay string `yaml:"transmit_pre_delay"`
}

// Endpoint data
type Endpoint struct {
	ID             string
	Config         Config
	connUrl        *url.URL
	conn           net.Conn
	receiveMsgFunc func(rm *model.Message)
	statusFunc     func(state *model.State)
	safeClose      *concurrency.Channel
	txPreDelay     time.Duration
}

// NewDevice ethernet driver
func NewDevice(ID string, config cmap.CustomMap, rxFunc func(msg *model.Message), statusFunc func(state *model.State)) (deviceType.Plugin, error) {
	var cfg Config
	err := utils.MapToStruct(utils.TagNameYaml, config, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.MessageSplitter == nil {
		splitter := byte(DefaultMessageSplitter)
		cfg.MessageSplitter = &splitter
	}

	zap.L().Debug("generated config", zap.Any("config", cfg))

	serverURL, err := url.Parse(cfg.Server)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial(serverURL.Scheme, serverURL.Host)
	if err != nil {
		return nil, err
	}

	endpoint := &Endpoint{
		ID:             ID,
		Config:         cfg,
		connUrl:        serverURL,
		conn:           conn,
		receiveMsgFunc: rxFunc,
		statusFunc:     statusFunc,
		safeClose:      concurrency.NewChannel(0),
		txPreDelay:     utils.ToDuration(cfg.TransmitPreDelay, transmitPreDelayDefault),
	}

	// start serail read listener
	go endpoint.dataListener()
	return endpoint, nil
}

func (ep *Endpoint) Name() string {
	return PluginEthernet
}

func (ep *Endpoint) Write(message *model.Message) error {
	if message == nil || len(message.Data) == 0 {
		return nil
	}
	time.Sleep(ep.txPreDelay) // transmit pre delay
	_, err := ep.conn.Write(append(message.Data, *ep.Config.MessageSplitter))
	return err
}

// Close the connection
func (ep *Endpoint) Close() error {
	go func() { ep.safeClose.SafeSend(true) }() // terminate the data listener

	if ep.conn != nil {
		err := ep.conn.Close()
		if err != nil {
			zap.L().Error("error on closing a connection", zap.String("adapterID", ep.ID), zap.String("server", ep.Config.Server), zap.Error(err))
		}
		ep.conn = nil
	}
	return nil
}

// DataListener func
func (ep *Endpoint) dataListener() {
	readBuf := make([]byte, 128)
	data := make([]byte, 0)
	for {
		select {
		case <-ep.safeClose.CH:
			zap.L().Info("received close signal", zap.String("adapterID", ep.ID), zap.String("server", ep.Config.Server))
			return
		default:
			rxLength, err := ep.conn.Read(readBuf)
			if err != nil {
				zap.L().Error("error on reading data from a ethernet connection", zap.String("adapterID", ep.ID), zap.String("server", ep.Config.Server), zap.Error(err))
				state := &model.State{
					Status:  model.StatusError,
					Message: err.Error(),
					Since:   time.Now(),
				}
				ep.statusFunc(state)
				return
			}

			for index := 0; index < rxLength; index++ {
				b := readBuf[index]
				if b == *ep.Config.MessageSplitter {
					// copy the received data
					dataCloned := make([]byte, len(data))
					copy(dataCloned, data)
					data = nil // reset local buffer
					message := model.NewMessage(dataCloned)
					ep.receiveMsgFunc(message)
				} else {
					data = append(data, b)
				}
				if len(data) >= MaxDataLength {
					data = nil
				}
			}
		}
	}
}
