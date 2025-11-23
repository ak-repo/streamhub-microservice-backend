package middleware

import (
	"strings"

	"github.com/ak-repo/stream-hub/pkg/jwt"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(jwtManager *jwt.JWTManager) fiber.Handler {
    return func(c *fiber.Ctx) error {

        // Get Authorization header
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Missing Authorization header",
            })
        }

        // Expected: Bearer <token>
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Invalid Authorization header format",
            })
        }

        tokenStr := parts[1]

        // Validate token with your jwt manager
        claims, err := jwtManager.ValidateToken(tokenStr)
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": err.Error(),
            })
        }

        // Store claims for next handlers
        c.Locals("claims", claims)

        // Continue
        return c.Next()
    }
}
