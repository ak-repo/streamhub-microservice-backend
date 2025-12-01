package middleware

import (
	"strings"

	"github.com/ak-repo/stream-hub/pkg/jwt"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(jwtManager *jwt.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {

		var tokenStr string

		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				tokenStr = parts[1]
			} else {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid Authorization header format. Expected: Bearer <token>",
				})
			}
		}

		if tokenStr == "" {
			tokenStr = c.Cookies("access")
			if tokenStr == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Missing token (no Bearer token and no access cookie)",
				})
			}
		}

		claims, err := jwtManager.ValidateToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token: " + err.Error(),
			})
		}

		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
