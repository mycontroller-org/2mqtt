package serial

import (
	"time"

	"github.com/mycontroller-org/2mqtt/pkg/types"
	deviceType "github.com/mycontroller-org/2mqtt/plugin/device/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	ser "github.com/tarm/serial"
	"go.uber.org/zap"
)

// Constants in serial device
const (
	PluginSerial = "serial"

	MaxDataLength           = 1000
	transmitPreDelayDefault = time.Millisecond * 1 // 1ms

	DefaultMessageSplitter = '\n'
)

// Config details
type Config struct {
	Port             string `yaml:"port"`
	BaudRate         int    `yaml:"baud_rate"`
	MessageSplitter  *byte  `yaml:"message_splitter"`
	TransmitPreDelay string `yaml:"transmit_pre_delay"`
}

// Endpoint data
type Endpoint struct {
	ID             string
	Config         *Config
	serCfg         *ser.Config
	Port           *ser.Port
	receiveMsgFunc func(rm *model.Message)
	statusFunc     func(state *model.State)
	safeClose      *concurrency.Channel
	txPreDelay     time.Duration
}

// New serial client
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

	serCfg := &ser.Config{Name: cfg.Port, Baud: cfg.BaudRate}

	zap.L().Info("opening a serial port", zap.String("adapterName", ID), zap.String("port", cfg.Port))
	port, err := ser.OpenPort(serCfg)
	if err != nil {
		return nil, err
	}

	endpoint := &Endpoint{
		ID:             ID,
		Config:         &cfg,
		serCfg:         serCfg,
		receiveMsgFunc: rxFunc,
		statusFunc:     statusFunc,
		Port:           port,
		safeClose:      concurrency.NewChannel(0),
		txPreDelay:     utils.ToDuration(cfg.TransmitPreDelay, transmitPreDelayDefault),
	}

	// start serail read listener
	go endpoint.dataListener()

	return endpoint, nil
}

func (ep *Endpoint) Name() string {
	return PluginSerial
}

func (ep *Endpoint) Write(message *model.Message) error {
	if message == nil && len(message.Data) > 0 {
		return nil
	}
	time.Sleep(ep.txPreDelay) // transmit pre delay
	_, err := ep.Port.Write(append(message.Data, *ep.Config.MessageSplitter))
	if err != nil {
		ep.statusFunc(&model.State{
			Status:  model.StatusError,
			Message: err.Error(),
			Since:   time.Now(),
		})
	}
	return err
}

// Close the driver
func (ep *Endpoint) Close() error {
	go func() { ep.safeClose.SafeSend(true) }() // terminate the data listener

	if ep.Port != nil {
		if err := ep.Port.Flush(); err != nil {
			zap.L().Error("error on flushing into serial port", zap.String("adapterName", ep.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
		}
		err := ep.Port.Close()
		if err != nil {
			zap.L().Error("error on closing the serial port", zap.String("adapterName", ep.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
		}
		return err
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
			zap.L().Info("received a close signal.", zap.String("id", ep.ID), zap.String("port", ep.serCfg.Name))
			return
		default:
			rxLength, err := ep.Port.Read(readBuf)
			if err != nil {
				zap.L().Error("error on reading data from a serial port", zap.String("adapterName", ep.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
				// notify failed
				if err != nil {
					ep.statusFunc(&model.State{
						Status:  model.StatusError,
						Message: err.Error(),
						Since:   time.Now(),
					})
				}
				return
			}
			//zap.L().Debug("data", zap.Any("data", string(data)))
			for index := 0; index < rxLength; index++ {
				b := readBuf[index]
				if b == *ep.Config.MessageSplitter {
					// copy the received data
					dataCloned := make([]byte, len(data))
					copy(dataCloned, data)
					data = nil // reset local buffer
					rawMsg := model.NewMessage(dataCloned)
					ep.receiveMsgFunc(rawMsg)
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
