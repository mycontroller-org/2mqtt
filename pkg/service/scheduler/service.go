package scheduler

import (
	"fmt"

	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"

	"go.uber.org/zap"
)

// Schedule adds a schedule
func Schedule(schedulerID, interval string, triggerFunc func()) error {
	Unschedule(schedulerID)
	zap.L().Info("schedule", zap.String("id", schedulerID), zap.String("interval", interval), zap.Any("triggerFunc", triggerFunc))
	cronSpec := fmt.Sprintf("@every %s", interval)
	err := coreScheduler.SVC.AddFunc(schedulerID, cronSpec, triggerFunc)
	if err != nil {
		zap.L().Error("error on adding schedule", zap.Error(err))
		return err
	}
	zap.L().Debug("added a schedule", zap.String("schedulerID", schedulerID), zap.String("interval", interval))
	return nil
}

// UnscheduleAll removes all with prefix
func UnscheduleAll(prefix string) {
	coreScheduler.SVC.RemoveWithPrefix(prefix)
}

// Unschedule removes a schedule
func Unschedule(scheduleID string) {
	coreScheduler.SVC.RemoveFunc(scheduleID)
}
