package routes

import (
	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/authpb"
	"github.com/ak-repo/stream-hub/gen/filespb"
	"github.com/ak-repo/stream-hub/internal/gateway/clients"
	"github.com/ak-repo/stream-hub/internal/gateway/handler"
	"github.com/ak-repo/stream-hub/pkg/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func New(app *fiber.App, cfg *config.Config, clients *clients.Clients) {
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Authorization",
	}))
	api := app.Group("/api/v1")

	jwtMan := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.Expiry*7)

	// routes
	authRoutes(api, clients.Auth, jwtMan)
	fileRoutes(api, clients.File)

}

func authRoutes(api fiber.Router, client authpb.AuthServiceClient, jwtMan *jwt.JWTManager) {

	auth := handler.NewAuthHandler(client, jwtMan)

	r := api.Group("/auth")
	r.Post("/login", auth.Login)
	r.Post("/register", auth.Register)
	r.Post("/verify-gen", auth.SendMagicLink)
	r.Get("/verify-link", auth.VerifyMagicLink)

}

func fileRoutes(api fiber.Router, client filespb.FileServiceClient) {

	file := handler.NewFileHandler(client)

	r := api.Group("/files")

	r.Post("/upload-url", file.GenerateUploadURL)
	r.Post("/confirm", file.ConfirmUpload)
	r.Get("/download-url", file.GenerateDownloadURL)
	r.Get("/:owner_id", file.ListFiles)
	r.Delete("/:file_id",file.DeleteFile)

}
