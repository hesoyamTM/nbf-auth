package v1

import (
	"context"
	"fmt"
	"net"

	"github.com/hesoyamTM/nbf-auth/internal/adapters/databases/redis"
	"github.com/hesoyamTM/nbf-auth/internal/adapters/oauth2/google"
	"github.com/hesoyamTM/nbf-auth/internal/adapters/userclient"
	"github.com/hesoyamTM/nbf-auth/internal/app/session"
	"github.com/hesoyamTM/nbf-auth/internal/config"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"google.golang.org/grpc"
)

type GrpcApp struct {
	server *grpc.Server
	host   string
	port   int
}

func NewGrpcApp(ctx context.Context, cfg *config.Config) (*GrpcApp, error) {
	const op = "grpcv1.NewGrpcApp"

	logInterceptor, err := logger.NewLoggingInterceptor(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(logInterceptor),
	)

	userRepo := redis.NewUserRepository(ctx, cfg.Redis)
	sessionRepo := redis.NewSessionRepository(ctx, cfg.Redis)
	stateRepo := redis.NewStateRepository(ctx, cfg.Redis)

	userClient := userclient.NewUserClient(ctx)
	googleAuthService := google.NewGoogleAuth(ctx, cfg.Google)

	authService, err := session.NewService(
		ctx,
		cfg.App,
		sessionRepo,
		userRepo,
		stateRepo,
		userClient,
		googleAuthService,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	Register(server, authService)

	return &GrpcApp{
		server: server,
		host:   cfg.Grpc.Host,
		port:   cfg.Grpc.Port,
	}, nil
}

func (a *GrpcApp) MustStart(ctx context.Context) {
	const op = "grpcv1.MustStart"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}

	log.Info("gRPC server is staring")

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.host, a.port))
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}

	if err := a.server.Serve(l); err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}
}

func (a *GrpcApp) MustStop(ctx context.Context) {
	const op = "grpcv1.MustStop"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}

	log.Info("gRPC server is stopping")

	a.server.GracefulStop()

	log.Info("gRPC server stopped")
}
