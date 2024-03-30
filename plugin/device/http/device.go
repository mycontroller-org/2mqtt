package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	model "github.com/mycontroller-org/2mqtt/pkg/types"
	contextTY "github.com/mycontroller-org/2mqtt/pkg/types/context"
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
	logger         *zap.Logger
	ID             string
	Config         *Config
	receiveMsgFunc func(rm *model.Message)
	statusFunc     func(state *model.State)
	listener       net.Listener
}

// New http client
func NewDevice(ctx context.Context, ID string, config cmap.CustomMap, rxFunc func(msg *model.Message), statusFunc func(state *model.State)) (deviceType.Plugin, error) {
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

	logger.Debug("source device config", zap.String("id", ID), zap.Any("config", cfg))

	logger.Info("opening the listening address", zap.String("adapterName", ID), zap.String("listenAddress", cfg.ListenAddress))
	listener, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		return nil, err
	}

	endpoint := &Endpoint{
		logger:         logger.Named("http_client"),
		ID:             ID,
		Config:         &cfg,
		listener:       listener,
		receiveMsgFunc: rxFunc,
		statusFunc:     statusFunc,
	}

	// handler
	var handler http.Handler = deviceHandler{logger: endpoint.logger, receiveMsgFunc: rxFunc, ID: ID}
	// include basic auth if enabled
	if cfg.IsAuthEnabled {
		handler = MiddlewareBasicAuthentication(cfg.Username, cfg.Password, handler)
	}

	// start the server
	server := &http.Server{
		ReadTimeout: defaultReadTimeout,
		Handler:     handler,
		// 	ErrorLog:    , // implement error logger
	}

	go func() {
		err = server.Serve(endpoint.listener)
		if err != nil {
			endpoint.logger.Error("error serve", zap.String("adapterName", ID), zap.String("listenAddress", cfg.ListenAddress), zap.Error(err))
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
	return fmt.Errorf("write not support not implemented, adapterName:%s", ep.ID)
}

// Close the driver
func (ep *Endpoint) Close() error {
	if ep.listener != nil {
		err := ep.listener.Close()
		ep.listener = nil
		ep.logger.Error("error on closing listener", zap.Error(err))
		return err
	}
	return nil
}
