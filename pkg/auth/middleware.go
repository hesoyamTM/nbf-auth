package auth

import (
	"context"
	"crypto/ecdsa"
	"net/http"
	"time"

	"github.com/hesoyamTM/nbf-auth/internal/domain/session"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"go.uber.org/zap"
)

type (
	ctxKey         string
	Middleware     func(next http.Handler) http.Handler
	RequestContext struct {
		Path   string
		Method string
	}

	AuthClient interface {
		IsUserBlocked(ctx context.Context, uid string) (bool, error)
	}
)

const (
	UID     ctxKey = "uid"
	NAME    ctxKey = "name"
	SURNAME ctxKey = "surname"
)

func NewAuthMiddleware(cookieAccessTokenName string, AuthClient AuthClient, publicKey *ecdsa.PublicKey) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l, err := logger.LoggerFromCtx(r.Context())
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}

			l.Info("Validating access token", zap.String("path", r.URL.Path), zap.String("method", r.Method))

			cookieToken := r.CookiesNamed(cookieAccessTokenName)

			if len(cookieToken) == 0 {
				l.Error("cookies is empty", zap.String("path", r.URL.Path), zap.String("method", r.Method))
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			stringToken := cookieToken[0].Value
			token, err := session.NewTokens(stringToken, "")
			if err != nil {
				l.Error("failed to parse token", zap.Error(err), zap.String("path", r.URL.Path), zap.String("method", r.Method))
				http.Error(w, "Failed to parse token", http.StatusUnauthorized)
				return
			}

			user, err := token.ValidateAccessToken(publicKey)
			if err != nil {
				l.Error("failed to validate token", zap.Error(err), zap.String("path", r.URL.Path), zap.String("method", r.Method))
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			isBlocked, err := AuthClient.IsUserBlocked(r.Context(), user.ID.String())
			if err != nil {
				l.Error("failed to check if user is blocked", zap.Error(err), zap.String("path", r.URL.Path), zap.String("method", r.Method))
				http.Error(w, "Failed to check if user is blocked", http.StatusInternalServerError)
				return
			}
			if isBlocked {
				l.Error("user is blocked", zap.String("path", r.URL.Path), zap.String("method", r.Method))

				http.SetCookie(w, &http.Cookie{
					Name:     cookieAccessTokenName,
					Value:    "",
					Path:     "/",
					Expires:  time.Now().Add(-1 * time.Hour),
					HttpOnly: true,
					Secure:   true,
				})

				http.Error(w, "user is blocked", http.StatusForbidden)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UID, user.ID.String())
			ctx = context.WithValue(ctx, NAME, user.Name)
			ctx = context.WithValue(ctx, SURNAME, user.Surname)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
