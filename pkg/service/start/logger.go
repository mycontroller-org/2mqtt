package start

import (
	cfg "github.com/mycontroller-org/2mqtt/pkg/service/configuration"
	"github.com/mycontroller-org/2mqtt/pkg/version"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	"go.uber.org/zap"
)

// InitLogger func
func InitLogger() {
	logger := loggerUtils.GetLogger(cfg.CFG.Logger.Mode, cfg.CFG.Logger.Level, cfg.CFG.Logger.Encoding, false, 0, false)
	zap.ReplaceGlobals(logger)
	zap.L().Info("welcome to the 2mqtt adapter server :)")
	zap.L().Info("server detail", zap.Any("version", version.Get()), zap.Any("loggerConfig", cfg.CFG.Logger))
}
