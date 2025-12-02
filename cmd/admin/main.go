package main

import (
	"context"
	"log"
	"net"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/adminpb"
	admingrpc "github.com/ak-repo/stream-hub/internal/admin_service/adapter/grpc"
	"github.com/ak-repo/stream-hub/internal/admin_service/adapter/postgres"
	"github.com/ak-repo/stream-hub/internal/admin_service/app"
	"github.com/ak-repo/stream-hub/internal/gateway/clients"
	"github.com/ak-repo/stream-hub/pkg/db"
	"github.com/ak-repo/stream-hub/pkg/grpc/interceptors"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/logger"
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

	// Initialize Clean Architecture layers
	repo := postgres.NewAdminRepo(pgDB.Pool)
	service := app.NewAdminService(repo, clientContainer)
	grpcHandler := admingrpc.NewAdminServer(service)

	// Start gRPC server
	port := ":" + cfg.Services.Admin.Port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.AppErrorInterceptor(), interceptors.UnaryLoggingInterceptor()))

	adminpb.RegisterAdminServiceServer(grpcServer, grpcHandler)

	log.Println("admin service started at: ", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("grpc admin server failed ", err.Error())
	}

}
