package http

import (
	"fmt"
	"net"
	"net/http"
	"time"

	model "github.com/mycontroller-org/2mqtt/pkg/types"
	deviceType "github.com/mycontroller-org/2mqtt/plugin/device/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

// Constants
const (
	PluginHTTP = "http"
)

var (
	defaultReadTimeout = time.Second * 60
)

// Config details
type Config struct {
	ListenAddress string `yaml:"listen_address"`
	IsAuthEnabled bool   `yaml:"is_auth_enabled"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password" json:"-"`
}

// Endpoint data
type Endpoint struct {
	ID             string
	Config         *Config
	receiveMsgFunc func(rm *model.Message)
	statusFunc     func(state *model.State)
	listener       net.Listener
}

// New http client
func NewDevice(ID string, config cmap.CustomMap, rxFunc func(msg *model.Message), statusFunc func(state *model.State)) (deviceType.Plugin, error) {
	var cfg Config
	err := utils.MapToStruct(utils.TagNameYaml, config, &cfg)
	if err != nil {
		return nil, err
	}

	zap.L().Debug("generated config", zap.Any("config", cfg))

	zap.L().Info("opening the listening address", zap.String("adapterName", ID), zap.String("listenAddress", cfg.ListenAddress))
	listener, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		return nil, err
	}

	endpoint := &Endpoint{
		ID:             ID,
		Config:         &cfg,
		listener:       listener,
		receiveMsgFunc: rxFunc,
		statusFunc:     statusFunc,
	}

	// start the server
	server := &http.Server{
		ReadTimeout: defaultReadTimeout,
		Handler:     deviceHandler{receiveMsgFunc: rxFunc, ID: ID},
		// 	ErrorLog:    , // implement error logger
	}

	go func() {
		err = server.Serve(endpoint.listener)
		if err != nil {
			zap.L().Error("error serve", zap.String("adapterName", ID), zap.String("listenAddress", cfg.ListenAddress), zap.Error(err))
			endpoint.statusFunc(&model.State{
				Status:  model.StatusError,
				Message: err.Error(),
				Since:   time.Now(),
			})
		}
	}()

	endpoint.statusFunc(&model.State{
		Status:  model.StatusUP,
		Message: "",
		Since:   time.Now(),
	})

	return endpoint, nil
}

func (ep *Endpoint) Name() string {
	return PluginHTTP
}

func (ep *Endpoint) Write(message *model.Message) error {
	return fmt.Errorf("write not supported by this device")
}

// Close the driver
func (ep *Endpoint) Close() error {
	if ep.listener != nil {
		err := ep.listener.Close()
		ep.listener = nil
		zap.L().Error("error on closing listener", zap.Error(err))
		return err
	}
	return nil
}
