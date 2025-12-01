package response

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Success sends a standard JSON success response
func Success(ctx *fiber.Ctx, message string, data interface{}) error {
	resp := fiber.Map{
		"success": true,
		"message": message,
	}

	if data != nil {
		if pb, ok := data.(proto.Message); ok {
			jsonBytes, err := protojson.MarshalOptions{
				UseProtoNames:   false, // camelCase output
				EmitUnpopulated: false,
			}.Marshal(pb)

			if err != nil {
				return Error(ctx, fiber.StatusInternalServerError, fiber.Map{"error": "Failed to marshal protobuf"})
			}

			// Unmarshal JSON bytes back into map for Fiber
			var jsonMap map[string]interface{}
			if err := ctx.App().Config().JSONDecoder(jsonBytes, &jsonMap); err != nil {
				return Error(ctx, fiber.StatusInternalServerError, fiber.Map{"error": "Failed to decode JSON"})
			}

			resp["data"] = jsonMap
		} else {
			// Normal struct/map → return as-is
			resp["data"] = data
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(resp)
}

// Error sends a standard JSON error response
func Error(c *fiber.Ctx, statusCode int, resp fiber.Map) error {

	return c.Status(statusCode).JSON(resp)
}

func InvalidReqBody(c *fiber.Ctx) error {
	return Error(c, fiber.StatusBadRequest, fiber.Map{"error": "invalid request body"})
}
