package redis

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/internal/domain/user"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type UserRepository struct {
	rdb *redis.Client
}

func NewUserRepository(ctx context.Context, cfg RedisConfig) *UserRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &UserRepository{
		rdb,
	}
}

func (u *UserRepository) SaveUser(ctx context.Context, user *user.User) error {
	const op = "redis.SaveUser"

	if err := u.rdb.Set(ctx, user.AuthID, user.ID, 0).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (u *UserRepository) UserExist(ctx context.Context, userID string) (bool, error) {
	const op = "redis.UserExist"

	exist, err := u.rdb.Exists(ctx, userID).Result()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exist > 0, nil
}

func (u *UserRepository) GetUser(ctx context.Context, userID string) (uuid.UUID, error) {
	const op = "redis.GetUser"

	res, err := u.rdb.Get(ctx, userID).Bytes()
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("uuid: ", zap.ByteString("uuid", res))

	uid, err := uuid.ParseBytes(res)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return uid, nil
}

func (u *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	const op = "redis.DeleteUser"

	if err := u.rdb.Del(ctx, userID).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
