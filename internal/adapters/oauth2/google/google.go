package google

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/internal/domain/user"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	googleclient "google.golang.org/api/oauth2/v2"
)

type GoogleAuthConfig struct {
	ClientID     string `env:"GOOGLE_AUTH_CLIENT_ID" yaml:"client-id" env-required:"true"`
	ClientSecret string `env:"GOOGLE_AUTH_CLIENT_SECRET" yaml:"client-secret" env-required:"true"`
	Redirect     string `env:"GOOGLE_AUTH_REDIRECT" yaml:"redirect" env-required:"true"`
}

type GoogleAuth struct {
	oauthConfig *oauth2.Config
}

func NewGoogleAuth(ctx context.Context, cfg GoogleAuthConfig) *GoogleAuth {
	return &GoogleAuth{
		&oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.Redirect,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (g *GoogleAuth) LoginURL(ctx context.Context, state string) string {
	return g.oauthConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)
}

func (g *GoogleAuth) AuthorizeUser(ctx context.Context, code string) (*user.User, error) {
	const op = "google.AuthorizeUser"

	tokens, err := g.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	client := g.oauthConfig.Client(ctx, tokens)
	service, err := googleclient.New(client)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	googleUserInfo, err := service.Userinfo.Get().Do()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userInfo, err := user.NewUser(
		uuid.Nil,
		googleUserInfo.Id,
		googleUserInfo.GivenName,
		googleUserInfo.FamilyName,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return userInfo, nil
}
