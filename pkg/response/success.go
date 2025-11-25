package response

import "github.com/gofiber/fiber/v2"

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

// Error sends a standard JSON error response
func Error(ctx *fiber.Ctx, statusCode int, message string) error {
	resp := fiber.Map{
		"success": false,
		"message": message,
	}

	return ctx.Status(statusCode).JSON(resp)
}
