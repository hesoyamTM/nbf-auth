package userclient

import (
	"context"
	"fmt"

	"github.com/hesoyamTM/nbf-auth/internal/domain/user"
	"github.com/hesoyamTM/nbf-auth/pkg/auth"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	userv1 "github.com/hesoyamTM/nbf-protos/gen/go/user"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClientConfig struct {
	Address string `yaml:"address" env:"USER_CLIENT_ADDRESS" env-required:"true"`
}

type UserClient struct {
	api userv1.UserClient
}

func NewUserClient(ctx context.Context, address string) (*UserClient, error) {
	const op = "userclient.NewUserClient"

	cc, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(auth.SettingMetadataInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &UserClient{
		api: userv1.NewUserClient(cc),
	}, nil
}

func (u *UserClient) CreateUser(ctx context.Context, userInfo *user.User) error {
	const op = "userclient.CreateUser"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = u.api.CreateUser(ctx, &userv1.CreateUserRequest{
		User: &userv1.UserInfo{
			Id:          userInfo.ID.String(),
			Name:        userInfo.Name,
			Surname:     userInfo.Surname,
			Description: "",
		},
	})
	if err != nil {
		log.Error("Failed to create user", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
