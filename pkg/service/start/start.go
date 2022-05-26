package start

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	adapterSVC "github.com/mycontroller-org/2mqtt/pkg/service/adapter"
	cfg "github.com/mycontroller-org/2mqtt/pkg/service/configuration"
	sch "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	"go.uber.org/zap"
)

func StartServices() {
	start := time.Now()

	cfg.InitConfig()
	InitLogger()

	sch.Init() // scheduler

	// start adapter services
	adapterSVC.Start(cfg.CFG.Adapters)

	zap.L().Info("services started", zap.String("timeTaken", time.Since(start).String()))

	handleShutdownSignal()
}

// handleShutdownSignal func
func handleShutdownSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// waiting for signal
	sig := <-sigs
	close(sigs)

	zap.L().Info("shutdown initiated..", zap.Any("signal", sig))
	triggerShutdown()
}

func triggerShutdown() {
	start := time.Now()

	// close adapter services
	adapterSVC.Close()

	if sch.SVC != nil {
		sch.SVC.Close()
	}

	zap.L().Debug("closing services are done", zap.String("timeTaken", time.Since(start).String()))
	zap.L().Debug("bye, see you soon :)")

	os.Exit(0)
}
