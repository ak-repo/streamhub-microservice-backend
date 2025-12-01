package main

import (
	"log"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/internal/gateway/clients"
	"github.com/ak-repo/stream-hub/internal/gateway/routes"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
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
	clientContainer := clients.NewClients(cfg)
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
