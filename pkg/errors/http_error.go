package errors

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GRPCToFiber(err error) (int, fiber.Map) {
	if err == nil {
		return fiber.StatusOK, nil
	}

	st, ok := status.FromError(err)
	if !ok {
		// not a gRPC error → internal
		return fiber.StatusInternalServerError, fiber.Map{
			"error": "internal server error",
		}
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return fiber.StatusBadRequest, fiber.Map{"error": st.Message()}
	case codes.NotFound:
		return fiber.StatusNotFound, fiber.Map{"error": st.Message()}
	case codes.AlreadyExists:
		return fiber.StatusConflict, fiber.Map{"error": st.Message()}
	case codes.Unauthenticated:
		return fiber.StatusUnauthorized, fiber.Map{"error": st.Message()}
	case codes.Internal:
		return fiber.StatusInternalServerError, fiber.Map{"error": st.Message()}
	case codes.Unavailable:
		return fiber.StatusServiceUnavailable, fiber.Map{"error": "service unavailable"}
	default:
		return fiber.StatusInternalServerError, fiber.Map{"error": st.Message()}
	}
}
