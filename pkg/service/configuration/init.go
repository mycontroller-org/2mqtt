package configuration

import (
	"flag"
	"io/ioutil"

	cfgML "github.com/mycontroller-org/2mqtt/pkg/model/config"
	loggerUtils "github.com/mycontroller-org/backend/v2/pkg/utils/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// configuration globally accessable
var (
	CFG *cfgML.Config
)

// InitConfig func
func InitConfig() {
	// init a temp logger
	logger := loggerUtils.GetLogger("development", "error", "console", false, 0)

	cf := flag.String("config", "./config.yaml", "Configuration file")
	flag.Parse()
	if cf == nil {
		logger.Fatal("configuration file not supplied")
		return
	}
	d, err := ioutil.ReadFile(*cf)
	if err != nil {
		logger.Fatal("error on reading configuration file", zap.Error(err))
	}

	err = yaml.Unmarshal(d, &CFG)
	if err != nil {
		logger.Fatal("failed to parse yaml data", zap.Error(err))
	}
}
