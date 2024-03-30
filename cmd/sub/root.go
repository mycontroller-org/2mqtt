package sub

import (
	"context"
	"fmt"
	"os"

	"github.com/mycontroller-org/2mqtt/cmd/helper"
	cfgTY "github.com/mycontroller-org/2mqtt/pkg/types/config"
	"github.com/mycontroller-org/2mqtt/pkg/version"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	cfgFilePath string
	cfg         *cfgTY.Config
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFilePath, "config", "./config.yaml", "configuration file")
	rootCmd.AddCommand(versionCmd)
}

var rootCmd = &cobra.Command{
	// https://github.com/spf13/cobra/blob/main/shell_completions.md#adapting-the-default-completion-command
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Use:   "2mqtt",
	Short: "2mqtt",
	Long:  `2mqtt is MQTT bridge. You can convert the serial, ethernet to MQTT`,
	Run: func(cmd *cobra.Command, args []string) {
		// init a temp logger
		logger := loggerUtils.GetLogger("record_all", "error", "console", false, 0, false)

		if cfgFilePath == "" {
			zap.L().Fatal("config can not be empty")
		}

		// load config file
		d, err := os.ReadFile(cfgFilePath)
		if err != nil {
			logger.Fatal("error on reading config file", zap.Error(err))
		}

		err = yaml.Unmarshal(d, &cfg)
		if err != nil {
			logger.Fatal("failed to parse config file", zap.Error(err))
		}

		// start service
		ctx := context.Background()
		toMqtt := helper.ToMqtt{}
		toMqtt.Start(ctx, cfg)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		zap.L().Error("error on calling execution", zap.Error(err))
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:  "version",
	Long: "Prints version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Get().String())
	},
}
