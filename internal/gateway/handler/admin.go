package handler

import (
	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/adminpb"
	"github.com/ak-repo/stream-hub/internal/gateway/clients"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	clients *clients.Clients
	cfg     *config.Config
}

func NewAdminHandler(clients *clients.Clients,
	cfg *config.Config) *AdminHandler {
	return &AdminHandler{clients: clients, cfg: cfg}
}

// ---------------------- Users action-------------------
func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.GetTotalUsers(ctx, &adminpb.GetTotalUsersRequest{})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "list of all users", resp)
}

func (h *AdminHandler) ListActiveUsers(c *fiber.Ctx) error {

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.GetActiveUsers(ctx, &adminpb.GetActiveUsersRequest{})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "list of active users", resp)
}

func (h *AdminHandler) ListBannedUsers(c *fiber.Ctx) error {

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.GetBannedUsers(ctx, &adminpb.GetBannedUsersRequest{})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "list of banned users", resp)
}

func (h *AdminHandler) BanUser(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"userId"`
		Reason string `json:"reason"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.BanUser(ctx, &adminpb.BanUserRequest{
		UserId: req.UserID,
		Reason: req.Reason,
	})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel created", resp)
}

func (h *AdminHandler) UnbanUser(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"userId"`
		Reason string `json:"reason"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.UnbanUser(ctx, &adminpb.UnbanUserRequest{
		UserId: req.UserID,
		Reason: req.Reason,
	})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel created", resp)
}

func (h *AdminHandler) UpdateRole(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"userId"`
		Role   string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.UpdateRole(ctx, &adminpb.UpdateRoleRequest{
		UserId: req.UserID,
		Role:   req.Role,
	})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel created", resp)
}
