package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

const (
	prodEnv = "prod"
	devEnv  = "dev"
)

type Key string

var CtxKey Key = "CtxKey"

func SetupLogger(ctx context.Context, env string) (context.Context, error) {
	const op = "logger.SetupLogger"

	var zapLog *zap.Logger
	var err error

	switch env {
	case devEnv:
		zapLog, err = zap.NewDevelopment()
	case prodEnv:
		zapLog, err = zap.NewProduction()
	default:
		return ctx, ErrInvalidEnv
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ctx = context.WithValue(ctx, CtxKey, zapLog)

	return ctx, nil
}

func LoggerFromCtx(ctx context.Context) (*zap.Logger, error) {
	zapLog := ctx.Value(CtxKey)

	if zapLog == nil {
		return nil, ErrLoggerNil
	}

	resLog, ok := zapLog.(*zap.Logger)
	if !ok {
		return nil, ErrInvalidType
	}

	return resLog, nil
}
