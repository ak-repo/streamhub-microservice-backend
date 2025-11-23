package routes

import (
	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/authpb"
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
	authRoutes(api, clients.Auth, jwtMan)

}

func authRoutes(api fiber.Router, authClient authpb.AuthServiceClient, jwtMan *jwt.JWTManager) {

	auth := handler.NewAuthHandler(authClient, jwtMan)

	r := api.Group("/auth")
	r.Post("/login", auth.Login)
	r.Post("/register", auth.Register)
	r.Post("/verify-gen", auth.SendMagicLink)
	r.Get("/verify-link", auth.VerifyMagicLink)

}
