package handler

import (
	"github.com/ak-repo/stream-hub/gen/authpb"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/jwt"
	"github.com/ak-repo/stream-hub/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	client     authpb.AuthServiceClient
	jwtManager *jwt.JWTManager
}

func NewAuthHandler(cli authpb.AuthServiceClient, jwt *jwt.JWTManager) *AuthHandler {
	return &AuthHandler{client: cli, jwtManager: jwt}
}

// -------------------- Login Handler --------------------
func (h *AuthHandler) Login(ctx *fiber.Ctx) error {
	req := &authpb.LoginRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	gc, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.Login(gc, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}

	access, aExp, err := h.jwtManager.GenerateAccessToken(resp.User.Id, resp.User.Email)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "token creation failed"})
	}

	refresh, rExp, err := h.jwtManager.GenerateRefreshToken(resp.User.Id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "token creation failed"})
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh",
		Value:    refresh,
		Path:     "/",
		Expires:  rExp,
		HTTPOnly: true,
	})

	// Use success response helper
	return response.Success(ctx, "login successful", fiber.Map{
		"token": access,
		"exp":   aExp,
	})
}

// -------------------- Register Handler --------------------
func (h *AuthHandler) Register(ctx *fiber.Ctx) error {
	req := &authpb.RegisterRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	gc, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.Register(gc, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}

	return response.Success(ctx, "registration successful", resp)
}

// -------------------- Send Magic Link Handler --------------------
func (h *AuthHandler) SendMagicLink(ctx *fiber.Ctx) error {
	req := &authpb.SendMagicLinkRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	gc, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.SendMagicLink(gc, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}

	return response.Success(ctx, "magic link sent successfully", resp)
}

// -------------------- Verify Magic Link Handler --------------------
func (h *AuthHandler) VerifyMagicLink(ctx *fiber.Ctx) error {
	req := &authpb.VerifyMagicLinkRequest{
		Email: ctx.Query("email"),
		Token: ctx.Query("token"),
	}

	if req.Email == "" || req.Token == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing email or token"})
	}

	gc, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.VerifyMagicLink(gc, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}

	return response.Success(ctx, "magic link verified successfully", resp)
}
