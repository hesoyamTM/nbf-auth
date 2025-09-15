package logger

import (
	"context"

	"google.golang.org/grpc"
)

func NewLoggingInterceptor(logCtx context.Context) (grpc.UnaryServerInterceptor, error) {
	log, err := LoggerFromCtx(logCtx)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = context.WithValue(ctx, CtxKey, log)

		return handler(ctx, req)
	}, nil
}
