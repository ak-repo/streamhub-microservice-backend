package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/chatpb"
	chatgrpc "github.com/ak-repo/stream-hub/internal/chat_service/adapter/grpc"
	"github.com/ak-repo/stream-hub/internal/chat_service/adapter/postgres"
	chatredis "github.com/ak-repo/stream-hub/internal/chat_service/adapter/redis"
	"github.com/ak-repo/stream-hub/internal/chat_service/app"
	"github.com/ak-repo/stream-hub/pkg/db"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {

	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	helper.OverrideLocal(cfg)
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	defer logger.Sync()

	ctx := context.Background()

	// 2. Initialize PostgreSQL
	pgDB, err := db.NewPostgresDB(ctx, cfg)
	if err != nil {
		log.Fatal("failed to connect to database:", zap.Error(err))
	}
	defer pgDB.Close()

	// Redis
	rdbAddr := cfg.Redis.Host + ":" + cfg.Redis.Port
	rdb := redis.NewClient(&redis.Options{Addr: rdbAddr})

	// 4. Initialize Clean Architecture layers
	repo := postgres.NewChatRepo(pgDB.Pool)
	ps := chatredis.NewRedisPubSub(rdb)
	svc := app.NewChatService(repo, ps)
	grpcHandler := chatgrpc.NewChatServer(svc)

	// 5. Start gRPC server
	port := ":" + cfg.Services.Chat.Port
	log.Println("port:", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	chatpb.RegisterChatServiceServer(grpcServer, grpcHandler)

	// 6. Graceful shutdown
	go func() {
		log.Printf("Chat Service (gRPC) running on %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("Server stopped")

}
