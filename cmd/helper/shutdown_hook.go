package helper

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

var (
	graceTerminationPeriod = time.Second * 30 // shutdown grace termination period
)

type ShutdownHook struct {
	logger       *zap.Logger
	callbackFunc func()
}

func NewShutdownHook(logger *zap.Logger, callbackFunc func()) *ShutdownHook {
	return &ShutdownHook{
		logger:       logger.Named("shutdown_hook"),
		callbackFunc: callbackFunc,
	}
}

func (sh *ShutdownHook) Start() {
	sh.handleShutdownSignal()
}

// handel process shutdown signal
func (sh *ShutdownHook) handleShutdownSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// waiting for signal
	sig := <-sigs
	close(sigs)

	sh.logger.Info("shutdown initiated..", zap.Any("signal", sig))
	sh.triggerShutdown()
}

func (sh *ShutdownHook) triggerShutdown() {
	start := time.Now()

	// force termination block
	ticker := time.NewTicker(graceTerminationPeriod)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				sh.logger.Warn("some services are not terminating on graceful period. Performing force termination", zap.String("gracePeriod", graceTerminationPeriod.String()))
				os.Exit(-1)
			}
		}
	}()

	// trigger callback function
	if sh.callbackFunc != nil {
		sh.callbackFunc()
	}

	// stop force termination ticker
	ticker.Stop()
	done <- true

	sh.logger.Info("closing services completed", zap.String("timeTaken", time.Since(start).String()))
	sh.logger.Debug("bye, see you soon :)")

	// stop web/api service
	os.Exit(0)
}
