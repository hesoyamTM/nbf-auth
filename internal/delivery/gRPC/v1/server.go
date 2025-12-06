package v1

import (
	"context"

	"github.com/hesoyamTM/nbf-auth/internal/domain/session"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	authv1 "github.com/hesoyamTM/nbf-protos/gen/go/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthService interface {
	GoogleLoginURL(ctx context.Context) (string, error)
	GoogleAuthorize(ctx context.Context, state, code string) (*session.Tokens, error)
	YandexLoginURL(ctx context.Context) (string, error)
	YandexAuthorize(ctx context.Context, state, code string) (*session.Tokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (*session.Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
	IsUserBlocked(ctx context.Context, userID string) (bool, error)
	BlockUser(ctx context.Context, userID string) error
	UnblockUser(ctx context.Context, userID string) error
}

type serverAPI struct {
	authv1.UnimplementedAuthServer
	authService AuthService
}

func Register(grpcServer *grpc.Server, authService AuthService) {
	authv1.RegisterAuthServer(grpcServer, &serverAPI{authService: authService})
}

func (s *serverAPI) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	return nil, nil
}

func (s *serverAPI) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return nil, nil
}

func (s *serverAPI) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal error")
	}

	if err := s.authService.Logout(ctx, req.GetRefreshToken()); err != nil {
		log.Error("Failed to refresh token", zap.Error(err))

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &authv1.LogoutResponse{}, nil
}

func (s *serverAPI) VerifyPhoneNumber(ctx context.Context, req *authv1.VerifyPhoneNumberRequest) (*authv1.VerifyPhoneNumberResponse, error) {
	return nil, nil
}

func (s *serverAPI) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal error")
	}

	tokens, err := s.authService.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		log.Error("Failed to refresh token", zap.Error(err))

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:     tokens.AccessToken,
		RefreshToken:    tokens.RefreshToken,
		AccessExpireAt:  timestamppb.New(tokens.AccessTokenExpireAt),
		RefreshExpireAt: timestamppb.New(tokens.RefreshTokenExpireAt),
	}, nil
}

func (s *serverAPI) GoogleLoginURL(ctx context.Context, req *authv1.GoogleLoginURLRequest) (*authv1.GoogleLoginURLResponse, error) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal error")
	}

	url, err := s.authService.GoogleLoginURL(ctx)
	if err != nil {
		log.Error("Failed to get google login url", zap.Error(err))

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &authv1.GoogleLoginURLResponse{
		Url: url,
	}, nil
}

func (s *serverAPI) GoogleAuthorize(ctx context.Context, req *authv1.GoogleAuthorizeRequest) (*authv1.GoogleAuthorizeResponse, error) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal error")
	}

	tokens, err := s.authService.GoogleAuthorize(ctx, req.GetState(), req.GetCode())
	if err != nil {
		log.Error("Failed to get google login url", zap.Error(err))

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &authv1.GoogleAuthorizeResponse{
		AccessToken:     tokens.AccessToken,
		RefreshToken:    tokens.RefreshToken,
		AccessExpireAt:  timestamppb.New(tokens.AccessTokenExpireAt),
		RefreshExpireAt: timestamppb.New(tokens.RefreshTokenExpireAt),
	}, nil
}

func (s *serverAPI) YandexLoginURL(ctx context.Context, req *authv1.YandexLoginURLRequest) (*authv1.YandexLoginURLResponse, error) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal error")
	}

	url, err := s.authService.YandexLoginURL(ctx)
	if err != nil {
		log.Error("Failed to get google login url", zap.Error(err))

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &authv1.YandexLoginURLResponse{
		Url: url,
	}, nil
}

func (s *serverAPI) YandexAuthorize(ctx context.Context, req *authv1.YandexAuthorizeRequest) (*authv1.YandexAuthorizeResponse, error) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal error")
	}

	tokens, err := s.authService.YandexAuthorize(ctx, req.GetState(), req.GetCode())
	if err != nil {
		log.Error("Failed to get google login url", zap.Error(err))

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &authv1.YandexAuthorizeResponse{
		AccessToken:     tokens.AccessToken,
		RefreshToken:    tokens.RefreshToken,
		AccessExpireAt:  timestamppb.New(tokens.AccessTokenExpireAt),
		RefreshExpireAt: timestamppb.New(tokens.RefreshTokenExpireAt),
	}, nil
}

func (s *serverAPI) IsUserBlocked(ctx context.Context, req *authv1.IsUserBlockedRequest) (*authv1.IsUserBlockedResponse, error) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal error")
	}

	blocked, err := s.authService.IsUserBlocked(ctx, req.GetUserId())
	if err != nil {
		log.Error("Failed to check user blocked", zap.Error(err))

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &authv1.IsUserBlockedResponse{
		Blocked: blocked,
	}, nil
}

func (s *serverAPI) BlockUser(ctx context.Context, req *authv1.BlockUserRequest) (*authv1.BlockUserResponse, error) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal error")
	}

	if err := s.authService.BlockUser(ctx, req.GetUserId()); err != nil {
		log.Error("Failed to block user", zap.Error(err))

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &authv1.BlockUserResponse{}, nil
}

func (s *serverAPI) UnblockUser(ctx context.Context, req *authv1.BlockUserRequest) (*authv1.BlockUserResponse, error) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal error")
	}

	if err := s.authService.UnblockUser(ctx, req.GetUserId()); err != nil {
		log.Error("Failed to unblock user", zap.Error(err))

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &authv1.BlockUserResponse{}, nil
}
