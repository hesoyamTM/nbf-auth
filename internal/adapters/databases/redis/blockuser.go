package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const BloomFilterName = "Cuckoo:blocked_users"

type BlockUserRepository struct {
	rdb *redis.Client
}

func NewBlockUserRepository(ctx context.Context, cfg RedisConfig) *BlockUserRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &BlockUserRepository{
		rdb,
	}
}

func (s *BlockUserRepository) BlockUser(ctx context.Context, userID string) error {
	const op = "redis.BlockUser"

	added, err := s.rdb.Do(ctx, "CF.ADD", BloomFilterName, userID).Bool()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !added {
		return fmt.Errorf("%s: user is already blocked", op)
	}

	return nil
}

func (s *BlockUserRepository) UnblockUser(ctx context.Context, userID string) error {
	const op = "redis.UnblockUser"

	removed, err := s.rdb.Do(ctx, "CF.DEL", BloomFilterName, userID).Bool()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !removed {
		return fmt.Errorf("%s: user is not blocked", op)
	}

	return nil
}

func (s *BlockUserRepository) IsUserBlocked(ctx context.Context, userID string) (bool, error) {
	const op = "redis.IsUserBlocked"

	exists, err := s.rdb.Do(ctx, "CF.EXISTS", BloomFilterName, userID).Bool()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}
