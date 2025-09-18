package yandex

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/internal/domain/user"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

type YandexAuthConfig struct {
	ClientID     string `env:"YANDEX_AUTH_CLIENT_ID" yaml:"client-id" env-required:"true"`
	ClientSecret string `env:"YANDEX_AUTH_CLIENT_SECRET" yaml:"client-secret" env-required:"true"`
	Redirect     string `env:"YANDEX_AUTH_REDIRECT" yaml:"redirect" env-required:"true"`
}

type YandexAuth struct {
	oauthConfig *oauth2.Config
}

type YandexUserInfo struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Sex       string `json:"sex"`
}

func NewYandexAuth(ctx context.Context, cfg YandexAuthConfig) *YandexAuth {
	return &YandexAuth{
		&oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.Redirect,
			Scopes: []string{
				"login:info",
			},
			Endpoint: yandex.Endpoint,
		},
	}
}

func (g *YandexAuth) LoginURL(ctx context.Context, state string) string {
	return g.oauthConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)
}

func (g *YandexAuth) AuthorizeUser(ctx context.Context, code string) (*user.User, error) {
	const op = "yandex.AuthorizeUser"

	tokens, err := g.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	client := g.oauthConfig.Client(ctx, tokens)
	resp, err := client.Get("https://login.yandex.ru/info?format=json")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer resp.Body.Close()

	var yandexUserInfo YandexUserInfo
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&yandexUserInfo); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userInfo, err := user.NewUser(
		uuid.Nil,
		yandexUserInfo.ID,
		yandexUserInfo.FirstName,
		yandexUserInfo.LastName,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return userInfo, nil
}
