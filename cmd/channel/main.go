package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/channelpb"
	channelgrpc "github.com/ak-repo/stream-hub/internal/channel_service/adapter/grpc"
	"github.com/ak-repo/stream-hub/internal/channel_service/adapter/postgres"
	chatredis "github.com/ak-repo/stream-hub/internal/channel_service/adapter/redis"
	"github.com/ak-repo/stream-hub/internal/channel_service/app"
	"github.com/ak-repo/stream-hub/pkg/db"
	"github.com/ak-repo/stream-hub/pkg/grpc/clients"
	"github.com/ak-repo/stream-hub/pkg/grpc/interceptors"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/logger"
	redisclient "github.com/ak-repo/stream-hub/pkg/redis"
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

	// ---- Create gRPC Client Container ----
	clientContainer := clients.NewClients(cfg)
	defer clientContainer.CloseAll()

	// Redis
	rAddr := cfg.Redis.Host + ":" + cfg.Redis.Port
	redisclient.Init(rAddr)
	rClient := redisclient.Client

	// 4. Initialize Clean Architecture layers
	repo := postgres.NewChannelRepo(pgDB.Pool)
	ps := chatredis.NewRedisPubSub(rClient)
	svc := app.NewChannelService(repo, ps, clientContainer, cfg)
	grpcHandler := channelgrpc.NewServer(svc)

	// 5. Start gRPC server
	port := ":" + cfg.Services.Chat.Port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.AppErrorInterceptor(), interceptors.UnaryLoggingInterceptor()))
	channelpb.RegisterChannelServiceServer(grpcServer, grpcHandler)
	channelpb.RegisterAdminChannelServiceServer(grpcServer, grpcHandler)

	//  Graceful shutdown
	go func() {
		log.Printf("Channel Service (gRPC) running on %s", port)
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
