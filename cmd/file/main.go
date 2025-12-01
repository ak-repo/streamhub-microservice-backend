package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/filespb"
	filegrpc "github.com/ak-repo/stream-hub/internal/files_service/adapter/grpc"
	"github.com/ak-repo/stream-hub/internal/files_service/adapter/postgres"
	redisstore "github.com/ak-repo/stream-hub/internal/files_service/adapter/redis"
	"github.com/ak-repo/stream-hub/internal/files_service/adapter/storage"
	"github.com/ak-repo/stream-hub/internal/files_service/app"
	"github.com/ak-repo/stream-hub/pkg/db"
	"github.com/ak-repo/stream-hub/pkg/grpc/interceptors"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"github.com/redis/go-redis/v9"
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

	// MinIO storage
	s3, err := storage.NewS3Storage(cfg.MinIO.Endpoint, cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey, cfg.MinIO.Bucket, cfg.MinIO.UseSSL, 15*time.Minute)
	if err != nil {
		log.Fatal("failed to connect MinIo storage:", zap.Error(err))
	}

	// Redis
	rdbAddr := cfg.Redis.Host + ":" + cfg.Redis.Port
	rdb := redis.NewClient(&redis.Options{Addr: rdbAddr})
	tempStore := redisstore.NewTempStore(rdb, 15*time.Minute)

	// repo - service - server
	repo := postgres.NewFileRepository(pgDB.Pool)
	service := app.NewFileService(repo, tempStore, s3, 15*time.Minute)
	server := filegrpc.NewFileServer(service)

	addr := ":" + cfg.Services.File.Port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("listen failed", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.AppErrorInterceptor(), interceptors.UnaryLoggingInterceptor()))

	filespb.RegisterFileServiceServer(grpcServer, server)

	log.Println("file-service started at:", cfg.Services.Auth.Host+addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("grpc file server failed ", zap.Error(err))
	}
}
