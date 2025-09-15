package redis

import (
	"context"
	"fmt"

	"github.com/hesoyamTM/nbf-auth/internal/adapters/databases"
	"github.com/redis/go-redis/v9"
)

type StateRepository struct {
	rdb *redis.Client
}

func NewStateRepository(ctx context.Context, cfg RedisConfig) *StateRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &StateRepository{
		rdb,
	}
}

func (s *StateRepository) SaveState(ctx context.Context, state string) error {
	const op = "redis.SaveState"

	if err := s.rdb.Set(ctx, state, nil, 0).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *StateRepository) ValidateState(ctx context.Context, state string) error {
	const op = "redis.ValidateState"

	exists, err := s.rdb.Exists(ctx, state).Result()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if exists <= 0 {
		return databases.ErrStateNotExists
	}

	return nil
}
