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

type BlockUserRepository interface {
	BlockUser(ctx context.Context, userID string) error
	UnblockUser(ctx context.Context, userID string) error
	IsUserBlocked(ctx context.Context, userID string) (bool, error)
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
	sessionRepo   SessionRepository
	userRepo      UserRepository
	stateRepo     StateRepository
	blockUserRepo BlockUserRepository

	googleAuthService OAuthService
	yandexAuthService OAuthService

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
	blockUserRepo BlockUserRepository,

	userService UserService,
	googleAuthService OAuthService,
	yandexAuthService OAuthService,
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
		blockUserRepo,

		googleAuthService,
		yandexAuthService,
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

	blocked, err := s.blockUserRepo.IsUserBlocked(ctx, userInfo.ID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if blocked {
		s.sessionRepo.DeleteSession(ctx, refreshToken)

		return nil, fmt.Errorf("%s: user is blocked", op)
	}

	newTokens, err := session.GenerateTokens(userInfo, s.privateKey, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = s.sessionRepo.RefreshSession(ctx, refreshToken, newTokens.RefreshToken, s.refreshTokenTTL)
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

func (s *Service) YandexLoginURL(ctx context.Context) (string, error) {
	return s.loginURL(ctx, s.yandexAuthService)
}

func (s *Service) YandexAuthorize(ctx context.Context, state, code string) (*session.Tokens, error) {
	return s.authorize(ctx, s.yandexAuthService, state, code)
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

		blocked, err := s.blockUserRepo.IsUserBlocked(ctx, userInfo.ID.String())
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if blocked {
			return nil, fmt.Errorf("%s: user is blocked", op)
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

	err = s.sessionRepo.SaveSession(ctx, tokens.RefreshToken, userInfo, s.refreshTokenTTL)
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

func (s *Service) IsUserBlocked(ctx context.Context, userID string) (bool, error) {
	const op = "session.IsUserBlocked"

	ok, err := s.blockUserRepo.IsUserBlocked(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return ok, nil
}

func (s *Service) BlockUser(ctx context.Context, userID string) error {
	const op = "session.BlockUser"

	if err := s.blockUserRepo.BlockUser(ctx, userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) UnblockUser(ctx context.Context, userID string) error {
	const op = "session.UnblockUser"

	if err := s.blockUserRepo.UnblockUser(ctx, userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
