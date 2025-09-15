package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	config "github.com/hesoyamTM/nbf-auth/internal/config"
	v1 "github.com/hesoyamTM/nbf-auth/internal/delivery/gRPC/v1"
	cfgtools "github.com/hesoyamTM/nbf-auth/pkg/config"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
)

func main() {
	cfg := cfgtools.MustParseConfig[config.Config]()
	ctx, err := logger.SetupLogger(context.Background(), cfg.Env)
	if err != nil {
		panic(err)
	}

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		panic(err)
	}

	log.Debug("Logger is working")

	grpcApp, err := v1.NewGrpcApp(ctx, cfg)
	if err != nil {
		panic(err)
	}

	go grpcApp.MustStart(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	grpcApp.MustStop(ctx)
	log.Info("Application stopped")
}
