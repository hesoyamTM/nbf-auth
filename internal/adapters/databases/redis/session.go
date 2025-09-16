package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hesoyamTM/nbf-auth/internal/domain/user"
	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepository(ctx context.Context, cfg RedisConfig) *SessionRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &SessionRepository{
		rdb,
	}
}

func (s *SessionRepository) SaveSession(ctx context.Context, refreshToken string, userInfo *user.User, ttl time.Duration) error {
	const op = "redis.SaveSession"

	data, err := json.Marshal(*userInfo)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.rdb.Set(ctx, refreshToken, data, ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *SessionRepository) RefreshSession(ctx context.Context, oldRefreshToken, newRefreshToken string, ttl time.Duration) error {
	const op = "redis.RefreshSession"

	if err := s.rdb.Rename(ctx, oldRefreshToken, newRefreshToken).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.rdb.Expire(ctx, newRefreshToken, ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *SessionRepository) GetSession(ctx context.Context, refreshToken string) (*user.User, error) {
	const op = "redis.GetSession"

	data, err := s.rdb.Get(ctx, refreshToken).Bytes()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var userInfo user.User
	if err = json.Unmarshal(data, &userInfo); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &userInfo, nil
}

func (s *SessionRepository) DeleteSession(ctx context.Context, refreshToken string) error {
	const op = "redis.DeleteSession"

	if err := s.rdb.Del(ctx, refreshToken).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
