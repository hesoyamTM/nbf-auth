package session

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

type OAuth2Tokens struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

func NewOAuth2Tokens(accessToken, refreshToken string, expiry time.Time) *OAuth2Tokens {
	return &OAuth2Tokens{
		accessToken,
		refreshToken,
		expiry,
	}
}

func OAuth2TokenFromCode(ctx context.Context, cfg *oauth2.Config, code string) (*OAuth2Tokens, error) {
	const op = "session.OAuth2TokenFromCode"

	oauthToken, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens := NewOAuth2Tokens(
		oauthToken.AccessToken,
		oauthToken.RefreshToken,
		oauthToken.Expiry)

	return tokens, nil
}

func (t *OAuth2Tokens) ConvertToken(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	const op = "session.ConvertToken"

	tokens := &oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Expiry:       t.Expiry,
	}

	oauth2Tokens, err := cfg.TokenSource(ctx, tokens).Token()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return oauth2Tokens, nil
}
