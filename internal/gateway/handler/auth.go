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
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	req := &authpb.LoginRequest{}
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	gc, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.Login(gc, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	access, aExp, err := h.jwtManager.GenerateAccessToken(resp.User.Id, resp.User.Email, resp.User.Role)
	if err != nil {

		return response.Error(c, fiber.StatusInternalServerError, fiber.Map{"error": "failed to generate access token"})
	}

	refresh, rExp, err := h.jwtManager.GenerateRefreshToken(resp.User.Id, resp.User.Role)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, fiber.Map{"error": "failed to generate refresh token"})
	}

	// Set cookies
	c.Cookie(&fiber.Cookie{
		Name:     "refresh",
		Value:    refresh,
		Path:     "/",
		Expires:  rExp,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "None",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "access",
		Value:    access,
		Path:     "/",
		Expires:  aExp,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "None",
	})

	return response.Success(c, "login successful", fiber.Map{
		"token": access,
		"user":  resp.User,
	})
}

// -------------------- Register Handler --------------------
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	req := &authpb.RegisterRequest{}
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	gc, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.Register(gc, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "registration successful", resp)
}

// -------------------- Send Magic Link Handler --------------------
func (h *AuthHandler) SendMagicLink(c *fiber.Ctx) error {
	req := &authpb.SendMagicLinkRequest{}
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	gc, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.SendMagicLink(gc, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "magic link sent successfully", resp)
}

// -------------------- Verify Magic Link Handler --------------------
func (h *AuthHandler) VerifyMagicLink(c *fiber.Ctx) error {
	req := &authpb.VerifyMagicLinkRequest{
		Email: c.Query("email"),
		Token: c.Query("token"),
	}

	if req.Email == "" || req.Token == "" {
		return response.InvalidReqBody(c)
	}

	gc, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.VerifyMagicLink(gc, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "magic link verified successfully", resp)
}
