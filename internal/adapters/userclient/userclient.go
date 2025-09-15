package userclient

import (
	"context"
	"fmt"

	"github.com/hesoyamTM/nbf-auth/internal/domain/user"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
)

type UserClient struct{}

func NewUserClient(ctx context.Context) *UserClient {
	return &UserClient{}
}

func (u *UserClient) CreateUser(ctx context.Context, userInfo *user.User) error {
	const op = "userclient.CreateUser"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("User created")

	return nil
}
