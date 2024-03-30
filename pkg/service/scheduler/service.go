package scheduler

import (
	"context"
	"errors"
	"fmt"

	contextTY "github.com/mycontroller-org/2mqtt/pkg/types/context"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"

	"go.uber.org/zap"
)

type Scheduler struct {
	logger        *zap.Logger
	coreScheduler schedulerTY.CoreScheduler
}

func FromContext(ctx context.Context) (*Scheduler, error) {
	_scheduler, ok := ctx.Value(contextTY.CUSTOM_SCHEDULER).(*Scheduler)
	if !ok {
		return nil, errors.New("invalid scheduler instance received in context")
	}
	if _scheduler == nil {
		return nil, errors.New("scheduler instance not provided in context")
	}
	return _scheduler, nil
}

func WithContext(ctx context.Context, scheduler *Scheduler) context.Context {
	return context.WithValue(ctx, contextTY.CUSTOM_SCHEDULER, scheduler)
}

func New(ctx context.Context) (*Scheduler, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	coreScheduler, err := schedulerTY.FromContext(ctx)
	if err != nil {
		logger.Error("error on getting a core scheduler", zap.Error(err))
		return nil, err
	}

	_scheduler := &Scheduler{
		logger:        logger.Named("scheduler"),
		coreScheduler: coreScheduler,
	}

	return _scheduler, nil
}

// Schedule adds a schedule
func (sh *Scheduler) Schedule(schedulerID, interval string, triggerFunc func()) error {
	sh.Unschedule(schedulerID)
	sh.logger.Info("schedule", zap.String("id", schedulerID), zap.String("interval", interval), zap.Any("triggerFunc", triggerFunc))
	cronSpec := fmt.Sprintf("@every %s", interval)
	err := sh.coreScheduler.AddFunc(schedulerID, cronSpec, triggerFunc)
	if err != nil {
		sh.logger.Error("error on adding schedule", zap.Error(err))
		return err
	}
	sh.logger.Debug("added a schedule", zap.String("schedulerID", schedulerID), zap.String("interval", interval))
	return nil
}

// UnscheduleAll removes all with prefix
func (sh *Scheduler) UnscheduleAll(prefix string) {
	sh.coreScheduler.RemoveWithPrefix(prefix)
}

// Unschedule removes a schedule
func (sh *Scheduler) Unschedule(scheduleID string) {
	sh.coreScheduler.RemoveFunc(scheduleID)
}
