package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	BloomFilterName = "Cuckoo:blocked_users"
	BatchSize       = 100
)

type BlockUserRepository struct {
	bf    *redis.Client
	block *redis.Client
}

func NewBlockUserRepository(ctx context.Context, cfg RedisConfig) *BlockUserRepository {
	bf := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	block := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       1,
	})

	bf.Do(ctx, "BF.RESERVE", BloomFilterName, "0.0001", "100000")

	return &BlockUserRepository{
		bf,
		block,
	}
}

func (s *BlockUserRepository) BlockUser(ctx context.Context, userID string) error {
	const op = "redis.BlockUser"

	added, err := s.bf.Do(ctx, "BF.ADD", BloomFilterName, userID).Bool()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !added {
		return fmt.Errorf("%s: user is already blocked", op)
	}

	if err := s.block.Do(ctx, "SET", userID, "1").Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *BlockUserRepository) UnblockUser(ctx context.Context, userID string) error {
	const op = "redis.UnblockUser"

	if err := s.block.Do(ctx, "DEL", userID).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.bf.Del(ctx, BloomFilterName).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	cursor := uint64(0)
	for cursor != 0 {
		var scanned []string
		var err error
		scanned, cursor, err = s.bf.Scan(ctx, cursor, "*", BatchSize).Result()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		for _, user := range scanned {
			if err := s.bf.Do(ctx, "BF.ADD", BloomFilterName, user).Err(); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}

	return nil
}

func (s *BlockUserRepository) IsUserBlocked(ctx context.Context, userID string) (bool, error) {
	const op = "redis.IsUserBlocked"

	exists, err := s.bf.Do(ctx, "BF.EXISTS", BloomFilterName, userID).Bool()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if !exists {
		return false, nil
	}

	blocked, err := s.block.Do(ctx, "GET", userID).Bool()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return blocked, nil
}
