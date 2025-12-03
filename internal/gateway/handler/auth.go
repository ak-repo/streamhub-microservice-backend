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

func (h *AuthHandler) PasswordReset(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.PasswordReset(ctx, &authpb.PasswordResetRequest{Email: req.Email})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "OTP sented into email: "+req.Email, resp)

}

func (h *AuthHandler) VerifyPasswordReset(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.VerifyPasswordReset(ctx, &authpb.PasswordResetVerifyRequest{Email: req.Email, Token: req.Token, NewPassword: req.Password})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "Password updated "+req.Email, resp)

}

func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.UpdateProfile(ctx, &authpb.UpdateProfileRequest{
		Username: req.Username,
		Email:    req.Email,
		Id:       uid,
	})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "profile updated", resp)
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {

	var req struct {
		Password    string `json:"password"`
		NewPassword string `json:"new_password"`
	}

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.ChangePassword(ctx, &authpb.ChangePasswordRequest{
		Id:          uid,
		Password:    req.Password,
		NewPassword: req.NewPassword,
	})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "password changed", resp)

}
