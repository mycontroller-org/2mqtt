package helper

import (
	"context"

	"github.com/mycontroller-org/2mqtt/pkg/version"
	"go.uber.org/zap"

	contextTY "github.com/mycontroller-org/2mqtt/pkg/types/context"

	cfgTY "github.com/mycontroller-org/2mqtt/pkg/types/config"
	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
)

// loads logger
func loadLogger(ctx context.Context, cfg cfgTY.LoggerConfig) (context.Context, *zap.Logger) {
	logger := loggerUtils.GetLogger(cfg.Mode, cfg.Level, cfg.Encoding, false, 0, cfg.EnableStacktrace)
	logger.Info("welcome to the 2mqtt adapter server :)")
	ver := version.Get()
	logger.Info("server information", zap.Any("version", ver), zap.Any("logger", cfg))

	// in some places still using "z.L()...", which needs global logger should be enabled
	// enabling global logger.
	// to fix this, do `grep -rl "zap\.L()"` and fix those manually.
	zap.ReplaceGlobals(logger)

	return contextTY.LoggerWithContext(ctx, logger), logger
}

// load core scheduler
func loadCoreScheduler(ctx context.Context) (context.Context, schedulerTY.CoreScheduler) {
	coreScheduler := coreScheduler.New()
	ctx = schedulerTY.WithContext(ctx, coreScheduler)
	return ctx, coreScheduler
}
