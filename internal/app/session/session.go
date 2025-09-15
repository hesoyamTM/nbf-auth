// Package session represents the logic of token management
package session

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/internal/config"
	"github.com/hesoyamTM/nbf-auth/internal/domain/session"
	"github.com/hesoyamTM/nbf-auth/internal/domain/user"
	cfgtools "github.com/hesoyamTM/nbf-auth/pkg/config"
)

type SessionRepository interface {
	SaveSession(ctx context.Context, refreshToken string, userInfo *user.User, ttl time.Duration) error
	GetSession(ctx context.Context, refreshToken string) (*user.User, error)
	RefreshSession(ctx context.Context, oldRefreshToken, newRefreshToken string, ttl time.Duration) error
	DeleteSession(ctx context.Context, refreshToken string) error
}

type UserRepository interface {
	SaveUser(ctx context.Context, userInfo *user.User) error
	UserExist(ctx context.Context, userID string) (bool, error)
	GetUser(ctx context.Context, userID string) (uuid.UUID, error)
	DeleteUser(ctx context.Context, userID string) error
}

type StateRepository interface {
	SaveState(ctx context.Context, state string) error
	ValidateState(ctx context.Context, state string) error
}

type UserService interface {
	CreateUser(ctx context.Context, userInfo *user.User) error
}

type OAuthService interface {
	LoginURL(ctx context.Context, state string) string
	AuthorizeUser(ctx context.Context, code string) (*user.User, error)
}

type Service struct {
	sessionRepo SessionRepository
	userRepo    UserRepository
	stateRepo   StateRepository

	googleAuthService OAuthService

	userService UserService

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	privateKey      *ecdsa.PrivateKey
}

func NewService(
	ctx context.Context,
	cfg config.APP,
	sessionRepo SessionRepository,
	userRepo UserRepository,
	stateRepo StateRepository,
	userService UserService,
	googleAuthService OAuthService,
) (*Service, error) {
	const op = "session.NewService"

	privateKey, err := cfgtools.DecodePrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Service{
		sessionRepo,
		userRepo,
		stateRepo,
		googleAuthService,
		userService,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		privateKey,
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*session.Tokens, error) {
	const op = "session.RefreshToken"

	userInfo, err := s.sessionRepo.GetSession(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	newTokens, err := session.GenerateTokens(userInfo, s.privateKey, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return newTokens, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	const op = "session.Login"

	if err := s.sessionRepo.DeleteSession(ctx, refreshToken); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) GoogleLoginURL(ctx context.Context) (string, error) {
	return s.loginURL(ctx, s.googleAuthService)
}

func (s *Service) GoogleAuthorize(ctx context.Context, state, code string) (*session.Tokens, error) {
	return s.authorize(ctx, s.googleAuthService, state, code)
}

func (s *Service) authorize(ctx context.Context, authService OAuthService, state, code string) (*session.Tokens, error) {
	const op = "session.Authorize"

	if err := s.stateRepo.ValidateState(ctx, state); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// TODO: UserExist replace by GetUser with ErrNotExist
	userInfo, err := authService.AuthorizeUser(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ok, err := s.userRepo.UserExist(ctx, userInfo.AuthID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if !ok {
		userInfo.ID = uuid.New()

		if err := s.userService.CreateUser(ctx, userInfo); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if err := s.userRepo.SaveUser(ctx, userInfo); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	} else {
		userInfo.ID, err = s.userRepo.GetUser(ctx, userInfo.AuthID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	tokens, err := session.GenerateTokens(
		userInfo,
		s.privateKey,
		s.accessTokenTTL,
		s.refreshTokenTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (s *Service) loginURL(ctx context.Context, authService OAuthService) (string, error) {
	const op = "session.loginURL"

	state := uuid.NewString()

	if err := s.stateRepo.SaveState(ctx, state); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return authService.LoginURL(ctx, state), nil
}
