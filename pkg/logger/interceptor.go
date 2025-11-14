// Package logger provides tools for logging. It is a wrapper around the
// standard log packages.
// Implements REST API middlewares and interceptors for gRPC.
package logger

import (
	"context"

	"google.golang.org/grpc"
)

// NewLoggingInterceptor returns a new logging unary interceptor for gRPC.
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

// NewLoggingStreamServerInterceptor returns a new logging stream interceptor for gRPC.
func NewLoggingStreamServerInterceptor(logCtx context.Context) (grpc.StreamServerInterceptor, error) {
	log, err := LoggerFromCtx(logCtx)
	if err != nil {
		return nil, err
	}

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		ctx = context.WithValue(ctx, CtxKey, log)

		wrappedSS := &wrappedStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrappedSS)
	}, nil
}

// wrappedStream is a wrapper around grpc.ServerStream that implements
// Context method with its own context.
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the context of the wrapped stream.
func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
