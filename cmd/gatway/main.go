package main

import (
	"log"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/authpb"
	"github.com/ak-repo/stream-hub/gen/filespb"
	"github.com/ak-repo/stream-hub/internal/gateway/clients"
	"github.com/ak-repo/stream-hub/internal/gateway/routes"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/zap"

	"google.golang.org/grpc"
)

func main() {
	// ---- Load Config ----
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	// ---- Local Development Overrides ----
	helper.OverrideLocal(cfg)

	// ---- Logger ----
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	defer logger.Sync()

	// ---- Fiber App ----
	app := fiber.New()
	app.Use(fiberlogger.New())

	// ---- Create gRPC Client Container ----
	clientContainer := clients.NewContainer()

	// Initialize Auth Client
	initClient(
		clientContainer,
		cfg.Services.Auth.Host,
		cfg.Services.Auth.Port,
		func(conn *grpc.ClientConn) interface{} { return authpb.NewAuthServiceClient(conn) },
		&clientContainer.Auth,
	)

	// Initialize File Client
	initClient(
		clientContainer,
		cfg.Services.File.Host,
		cfg.Services.File.Port,
		func(conn *grpc.ClientConn) interface{} { return filespb.NewFileServiceClient(conn) },
		&clientContainer.File,
	)

	// Clean up gRPC connections on exit
	defer clientContainer.CloseAll()

	// ---- Register Routes ----
	routes.New(app, cfg, clientContainer)

	// ---- Start Gateway HTTP Server ----
	addr := cfg.Services.Gateway.Host + ":" + cfg.Services.Gateway.Port
	log.Println("gateway started: ", addr)

	if err := app.Listen(addr); err != nil {
		log.Fatal("gateway startup failed ", err.Error())
	}
}

// initClient initializes a gRPC client in a clean, reusable way
func initClient(
	container *clients.Clients,
	host, port string,
	factory func(*grpc.ClientConn) interface{},
	target interface{},
) {
	cli, conn, err := clients.NewClient(host, port, factory)
	if err != nil {
		logger.Log.Fatal("gRPC client initialization failed", zap.Error(err))
	}

	// Assign the concrete type via pointer to interface
	switch t := target.(type) {
	case *authpb.AuthServiceClient:
		*t = cli.(authpb.AuthServiceClient)
	case *filespb.FileServiceClient:
		*t = cli.(filespb.FileServiceClient)
	default:
		logger.Log.Fatal("unsupported client type")
	}

	container.AddConn(conn)
}
