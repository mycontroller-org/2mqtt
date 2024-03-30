package helper

import (
	"context"
	"time"

	adapterSVC "github.com/mycontroller-org/2mqtt/pkg/service/adapter"
	"github.com/mycontroller-org/2mqtt/pkg/service/scheduler"
	"github.com/mycontroller-org/2mqtt/pkg/types/config"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	"go.uber.org/zap"
)

type ToMqtt struct {
	ctx    context.Context
	config *config.Config
	logger *zap.Logger

	// services
	coreSchedulerSVC schedulerTY.CoreScheduler // core scheduler, used to execute all the cron jobs
}

func (g *ToMqtt) Start(ctx context.Context, cfg *config.Config) error {
	startTime := time.Now()

	g.ctx = ctx

	// load logger
	ctx, logger := loadLogger(ctx, cfg.Logger)

	// get core scheduler and inject into context
	ctx, coreScheduler := loadCoreScheduler(ctx)
	err := coreScheduler.Start()
	if err != nil {
		logger.Error("error on starting core scheduler", zap.Error(err))
		return err
	}

	// inject custom scheduler into context
	customScheduler, err := scheduler.New(ctx)
	if err != nil {
		logger.Error("error on loading custom scheduler", zap.Error(err))
		return err
	}
	ctx = scheduler.WithContext(ctx, customScheduler)

	// add into struct
	g.ctx = ctx
	g.config = cfg
	g.logger = logger
	g.coreSchedulerSVC = coreScheduler

	// start adapter services
	err = adapterSVC.Start(ctx, cfg.Adapters)
	if err != nil {
		logger.Error("error on starting adapter services", zap.Error(err))
		return err
	}

	logger.Info("services are started", zap.String("timeTaken", time.Since(startTime).String()))

	// call shutdown hook
	shutdownHook := NewShutdownHook(g.logger, g.stop)
	shutdownHook.Start()

	return nil
}

func (g *ToMqtt) stop() {
	// stop services

	// stop adapter services
	g.logger.Debug("closing adapter services")
	adapterSVC.Close()

	g.logger.Debug("closing core scheduler")
	// stop core scheduler
	if err := g.coreSchedulerSVC.Close(); err != nil {
		g.logger.Error("error on closing core scheduler", zap.Error(err))
	}
}
