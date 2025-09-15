package session

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/internal/domain/user"
)

// Tokens is presentaion of session tokens
type Tokens struct {
	AccessToken          string
	RefreshToken         string
	AccessTokenExpireAt  time.Time
	RefreshTokenExpireAt time.Time
}

// NewTokens reutrn JwtTokens struct
func NewTokens(accessToken, refreshToken string) (*Tokens, error) {
	if accessToken == "" || refreshToken == "" {
		return nil, ErrEmptyToken
	}

	return &Tokens{
		accessToken,
		refreshToken,
		time.Time{},
		time.Time{},
	}, nil
}

// GenerateTokens create new access and refresh tokens
func GenerateTokens(userInfo *user.User, privateKey *ecdsa.PrivateKey, accessTTL, refreshTTL time.Duration) (*Tokens, error) {
	const op = "session.GenerateTokens"

	token := jwt.New(jwt.SigningMethodES256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = userInfo.ID
	claims["name"] = userInfo.Name
	claims["surname"] = userInfo.Surname
	claims["exp"] = time.Now().Add(accessTTL).Unix()

	accessToken, err := token.SignedString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Tokens{
		accessToken,
		uuid.NewString(),
		time.Now().Add(accessTTL),
		time.Now().Add(refreshTTL),
	}, nil
}

// ValidateAccessToken validates jwt token by ES256 public key. Function returns user info or error
func (t *Tokens) ValidateAccessToken(publicKey *ecdsa.PublicKey) (*user.User, error) {
	const op = "session.ValidateAccessToken"

	token, err := jwt.Parse(t.AccessToken, func(t *jwt.Token) (interface{}, error) { return publicKey, nil })
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	claims := token.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["uid"].(string))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userInfo, err := user.NewUser(userID, claims["name"].(string), claims["surname"].(string), "")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return userInfo, nil
}
