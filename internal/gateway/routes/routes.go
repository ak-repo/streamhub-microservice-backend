package routes

import (
	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/internal/gateway/clients"
	"github.com/ak-repo/stream-hub/internal/gateway/handler"
	"github.com/ak-repo/stream-hub/internal/gateway/middleware"
	"github.com/ak-repo/stream-hub/pkg/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
)

func New(app *fiber.App, cfg *config.Config, clients *clients.Clients) {
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, http://localhost:3000, http://localhost:3001,http://localhost:3002",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Authorization",
	}))

	api := app.Group("/api/v1")
	adminRoutes(api, clients, cfg)
	userRoutes(api, cfg, clients)

}

func userRoutes(api fiber.Router, cfg *config.Config, clients *clients.Clients) {
	jwtMan := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.Expiry*7)

	// --------------------------
	// AUTH HANDLER
	// --------------------------
	auth := handler.NewAuthHandler(clients.Auth, jwtMan)

	authR := api.Group("/auth")
	authR.Post("/login", auth.Login)
	authR.Post("/register", auth.Register)
	authR.Post("/verify-gen", auth.SendMagicLink)
	authR.Get("/verify-link", auth.VerifyMagicLink)

	// --------------------------
	// FILE HANDLER
	// --------------------------
	file := handler.NewFileHandler(clients.File)

	fileR := api.Group("/files")
	fileR.Use(middleware.AuthMiddleware(jwtMan))
	fileR.Post("/upload-url", file.GenerateUploadURL)
	fileR.Post("/confirm", file.ConfirmUpload)
	fileR.Get("/download-url", file.GenerateDownloadURL)
	fileR.Get("/", file.ListFiles)
	fileR.Delete("/delete", file.DeleteFile)

	// --------------------------
	// CHANNEL HANDLER
	// --------------------------
	channel := handler.NewChannelHandler(clients.Channel)

	// Websocket route
	api.Get("/ws", websocket.New(channel.WsHandler))

	ch := api.Group("/channels")
	ch.Use(middleware.AuthMiddleware(jwtMan))

	ch.Get("/", channel.ListChannels)
	ch.Post("/create", channel.CreateChannel)
	ch.Post("/join", channel.JoinChannel)
	ch.Post("/leave", channel.LeaveChannel)
	ch.Delete("/delete", channel.DeleteChannel)

	ch.Get("/channel/:channelId", channel.GetChannel)
	ch.Get("/members/:channelId", channel.ListMembers)
	ch.Get("/:channelId/history", channel.ListMessages)
}
