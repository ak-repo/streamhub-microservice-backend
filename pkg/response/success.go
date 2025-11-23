package response

import (
	"github.com/gofiber/fiber/v2"
)

// Success sends a standard JSON success response
func Success(ctx *fiber.Ctx, message string, data interface{}) error {
	resp := fiber.Map{
		"success": true,
		"message": message,
	}
	if data != nil {
		resp["data"] = data
	}
	return ctx.Status(fiber.StatusOK).JSON(resp)
}
