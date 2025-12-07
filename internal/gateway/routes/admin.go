package routes

import (
	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/internal/gateway/clients"
	"github.com/ak-repo/stream-hub/internal/gateway/handler"
	"github.com/ak-repo/stream-hub/internal/gateway/middleware"
	"github.com/ak-repo/stream-hub/pkg/jwt"
	"github.com/gofiber/fiber/v2"
)

func adminRoutes(api fiber.Router, clients *clients.Clients, cfg *config.Config) {

	jwtMan := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.Expiry*7)

	adminR := api.Group("/admin")
	adminR.Use(middleware.AuthMiddleware(jwtMan))

	handler := handler.NewAdminHandler(clients, cfg)

	// users actions
	adminR.Get("/users", handler.ListUsers)
	adminR.Post("/users/ban", handler.BanUser)
	adminR.Post("/users/unban", handler.UnbanUser)
	adminR.Post("/users/change-role", handler.UpdateRole)
	adminR.Post("/users/uploads-block", handler.BlockUserUpload)
	adminR.Delete("/users/:id", handler.DeleteUser)

	// channels actions
	adminR.Get("/channels", handler.ListChannels)
	// adminR.Get("/channels/:id", handler.GetChannelById)
	adminR.Post("/channels/freeze", handler.FreezeChannel)
	adminR.Post("/channels/unfreeze", handler.UnfreezeChannel)
	adminR.Delete("/channels/:id", handler.DeleteChannel)

	// files actions
	adminR.Get("/files", handler.ListAllFiles)
	adminR.Delete("/files/:id", handler.DeleteFile)

}
