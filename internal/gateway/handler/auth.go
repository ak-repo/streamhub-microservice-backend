package handler

import (
	"context"
	"io"
	"log"
	"time"

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

func (h *AuthHandler) ForgetPassword(c *fiber.Ctx) error {
	req := new(authpb.PasswordResetRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.PasswordReset(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "OTP sented into email: "+req.Email, resp)

}

func (h *AuthHandler) VerifyOTPForPasswordReset(c *fiber.Ctx) error {

	req := new(authpb.VerifyPasswordResetRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	log.Println("email: for pr", req.Email)
	log.Println("token:", req.Token)
	log.Println("pass: ", req.NewPassword)
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.VerifyPasswordReset(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "Password updated "+req.Email, resp)

}

func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	req := new(authpb.UpdateProfileRequest)

	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}
	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	req.UserId = uid

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.UpdateProfile(ctx, req)

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "profile updated", resp)
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	req := new(authpb.ChangePasswordRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	req.UserId = uid

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.ChangePassword(ctx, req)

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "password changed", resp)

}

func (h *AuthHandler) SearchUsers(c *fiber.Ctx) error {
	req := new(authpb.SearchUsersRequest)
	req.Query = c.Query("query")
	req.Pagination = &authpb.PaginationRequest{
		Offset: helper.StringToInt32(c.Query("offset")),
		Limit:  helper.StringToInt32(c.Query("limit")),
	}
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.SearchUsers(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "users", resp)

}

func (h *AuthHandler) UploadAvatar(c *fiber.Ctx) error {

	c.Set("Content-Type", "application/json")

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file upload error: " + err.Error(),
		})
	}

	if fileHeader.Size > 5<<20 { // enforce 5MB
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"error": "file too large; max 5MB",
		})
	}

	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to open uploaded file",
		})
	}
	defer file.Close()
	// Read bytes
	data, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to read file",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	resp, err := h.client.UploadAvatar(ctx, &authpb.UploadAvatarRequest{
		UserId:   uid,
		FileData: data,
		Filename: fileHeader.Filename,
		MimeType: fileHeader.Header.Get("Content-Type"),
	})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "profile pic updated", resp)

}
