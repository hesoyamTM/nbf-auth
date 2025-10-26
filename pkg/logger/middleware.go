package logger

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type Middleware func(next http.Handler) http.Handler

func NewLoggingMiddleware(lctx context.Context) (Middleware, error) {
	log, err := LoggerFromCtx(lctx)
	if err != nil {
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log = log.With(
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
			)
			ctx := context.WithValue(r.Context(), CtxKey, log)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}, nil
}
