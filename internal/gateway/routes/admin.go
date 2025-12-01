package routes

import (
	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/internal/gateway/clients"
	"github.com/ak-repo/stream-hub/internal/gateway/handler"
	"github.com/ak-repo/stream-hub/pkg/jwt"
	"github.com/gofiber/fiber/v2"
)

func adminRoutes(api fiber.Router, clients *clients.Clients, cfg *config.Config) {
	r := api.Group("/admin")

	jwtMan := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.Expiry*7)
	// Handlers
	auth := handler.NewAuthHandler(clients.Auth, jwtMan)

	r.Post("/login",auth.Login)

	
}
