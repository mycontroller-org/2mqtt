package context

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

const (
	CONFIG           ContextKey = "mc_config"
	LOGGER           ContextKey = "mc_logger"
	CUSTOM_SCHEDULER ContextKey = "mc_custom_scheduler"
)

type ContextKey string

func LoggerFromContext(ctx context.Context) (*zap.Logger, error) {
	logger, ok := ctx.Value(LOGGER).(*zap.Logger)
	if !ok {
		return nil, errors.New("invalid logger instance received in context")
	}
	if logger == nil {
		return nil, errors.New("logger instance not provided in context")
	}
	return logger, nil
}

func LoggerWithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, LOGGER, logger)
}
