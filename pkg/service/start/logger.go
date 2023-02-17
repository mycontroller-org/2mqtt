package start

import (
	cfgTY "github.com/mycontroller-org/2mqtt/pkg/types/config"
	"github.com/mycontroller-org/2mqtt/pkg/version"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	"go.uber.org/zap"
)

// InitLogger func
func InitLogger(loggerCfg cfgTY.LoggerConfig) {
	logger := loggerUtils.GetLogger(loggerCfg.Mode, loggerCfg.Level, loggerCfg.Encoding, false, 0, false)
	zap.ReplaceGlobals(logger)
	zap.L().Info("welcome to the 2mqtt adapter server :)")
	zap.L().Info("server information", zap.Any("version", version.Get()), zap.Any("logger", loggerCfg))
}
