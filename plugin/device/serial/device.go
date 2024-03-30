package serial

import (
	"context"
	"time"

	"github.com/mycontroller-org/2mqtt/pkg/types"
	contextTY "github.com/mycontroller-org/2mqtt/pkg/types/context"
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
	logger         *zap.Logger
	ID             string
	Config         *Config
	serCfg         *ser.Config
	Port           *ser.Port
	receiveMsgFunc func(rm *types.Message)
	statusFunc     func(state *types.State)
	safeClose      *concurrency.Channel
	txPreDelay     time.Duration
}

// New serial client
func NewDevice(ctx context.Context, ID string, config cmap.CustomMap, rxFunc func(msg *types.Message), statusFunc func(state *types.State)) (deviceType.Plugin, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = utils.MapToStruct(utils.TagNameYaml, config, &cfg)
	if err != nil {
		logger.Error("error on converting map to struct", zap.Error(err))
		return nil, err
	}

	if cfg.MessageSplitter == nil {
		splitter := byte(DefaultMessageSplitter)
		cfg.MessageSplitter = &splitter
	}

	logger.Debug("source device config", zap.String("id", ID), zap.Any("config", cfg))

	serCfg := &ser.Config{Name: cfg.Port, Baud: cfg.BaudRate}

	logger.Info("opening a serial port", zap.String("adapterName", ID), zap.String("port", cfg.Port))
	port, err := ser.OpenPort(serCfg)
	if err != nil {
		return nil, err
	}

	endpoint := &Endpoint{
		logger:         logger.Named("serial_client"),
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

func (ep *Endpoint) Write(message *types.Message) error {
	if message == nil || len(message.Data) == 0 {
		return nil
	}

	if ep.txPreDelay > 0 {
		time.Sleep(ep.txPreDelay) // transmit pre delay
	}
	_, err := ep.Port.Write(append(message.Data, *ep.Config.MessageSplitter))
	if err != nil {
		ep.statusFunc(&types.State{
			Status:  types.StatusError,
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
			ep.logger.Error("error on flushing into serial port", zap.String("adapterName", ep.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
		}
		err := ep.Port.Close()
		if err != nil {
			ep.logger.Error("error on closing the serial port", zap.String("adapterName", ep.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
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
			ep.logger.Info("received a close signal.", zap.String("id", ep.ID), zap.String("port", ep.serCfg.Name))
			return
		default:
			rxLength, err := ep.Port.Read(readBuf)
			if err != nil {
				ep.logger.Error("error on reading data from a serial port", zap.String("adapterName", ep.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
				// notify failed
				ep.statusFunc(&types.State{
					Status:  types.StatusError,
					Message: err.Error(),
					Since:   time.Now(),
				})
				return
			}
			//ep.logger.Debug("data", zap.Any("data", string(data)))
			for index := 0; index < rxLength; index++ {
				b := readBuf[index]
				if b == *ep.Config.MessageSplitter {
					// copy the received data
					dataCloned := make([]byte, len(data))
					copy(dataCloned, data)
					data = nil // reset local buffer
					rawMsg := types.NewMessage(dataCloned)
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
