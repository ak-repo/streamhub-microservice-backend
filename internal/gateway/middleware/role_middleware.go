package middleware

import "github.com/gofiber/fiber/v2"

func RoleRequired(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		role := c.Locals("role")
		if role == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "No role found in token",
			})
		}

		userRole := role.(string)

		if userRole == role {
			return c.Next()
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You are not allowed to access this resource",
		})
	}
}
