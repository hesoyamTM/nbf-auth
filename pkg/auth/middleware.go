package auth

import (
	"context"
	"crypto/ecdsa"
	"net/http"

	"github.com/hesoyamTM/nbf-auth/internal/domain/session"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
)

type (
	ctxKey     string
	Middleware func(next http.Handler) http.Handler
)

const (
	UID     ctxKey = "uid"
	NAME    ctxKey = "name"
	SURNAME ctxKey = "surname"
)

func NewAuthMiddleware(cookieAccessTokenName string, authMethods map[string]bool, publicKey *ecdsa.PublicKey) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l, err := logger.LoggerFromCtx(r.Context())
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}

			if !authMethods[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			cookieToken := r.CookiesNamed(cookieAccessTokenName)

			if len(cookieToken) == 0 {
				l.Error("cookies is empty")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			stringToken := cookieToken[0].Value
			token, err := session.NewTokens(stringToken, "")
			if err != nil {
				l.Error("failed to parse token")
				http.Error(w, "Failed to parse token", http.StatusUnauthorized)
				return
			}

			user, err := token.ValidateAccessToken(publicKey)
			if err != nil {
				l.Error(err.Error())
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), UID, user.ID))
			r = r.WithContext(context.WithValue(r.Context(), NAME, user.Name))
			r = r.WithContext(context.WithValue(r.Context(), SURNAME, user.Surname))

			next.ServeHTTP(w, r)
		})
	}
}
