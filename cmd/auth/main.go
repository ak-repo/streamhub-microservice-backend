package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/authpb"
	authgrpc "github.com/ak-repo/stream-hub/internal/auth_service/adapter/grpc"
	"github.com/ak-repo/stream-hub/internal/auth_service/adapter/postgres"
	"github.com/ak-repo/stream-hub/internal/auth_service/app"

	"github.com/ak-repo/stream-hub/pkg/db"
	"github.com/ak-repo/stream-hub/pkg/grpc/interceptors"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/jwt"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}
	helper.OverrideLocal(cfg)

	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	defer logger.Sync()

	//db
	pgDB, err := db.NewPostgresDB(context.Background(), cfg)
	if err != nil {
		log.Fatal("failed to connect db:", zap.Error(err))
	}
	defer pgDB.Close()

	// jwt manager
	tokenExpiry := 10 * time.Minute
	jwtMan := jwt.NewJWTManager(cfg.JWT.Secret, tokenExpiry, tokenExpiry)

	// repo -> service -> server
	repo := postgres.NewUserRepository(pgDB.Pool)
	service := app.NewAuthService(repo, jwtMan, cfg)
	server := authgrpc.NewAuthServer(service)

	addr := ":" + cfg.Services.Auth.Port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("listen failed", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.AppErrorInterceptor(), interceptors.UnaryLoggingInterceptor()))

	authpb.RegisterAuthServiceServer(grpcServer, server)

	log.Println("auth-service started at:", cfg.Services.Auth.Host+addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("grpc auth server failed ", zap.Error(err))
	}
}
