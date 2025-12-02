package handler

import (
	"log"

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
	filterBy := c.Params("filter", "all")

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()
	log.Println("filter", filterBy)
	resp, err := h.clients.Admin.ListUsers(ctx, &adminpb.ListUsersRequest{FilterBy: filterBy})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "list of all users", resp)
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

// ----------------- channel actions ----------------------
func (h *AdminHandler) ListChannels(c *fiber.Ctx) error {

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.ListChannels(ctx, &adminpb.ListChannelsRequest{})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "listing all channels", resp)
}

func (h *AdminHandler) FreezeChannel(c *fiber.Ctx) error {
	var req struct {
		ChannelID string `json:"channelId"`
		Reason    string `json:"reason"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.FreezeChannel(ctx, &adminpb.FreezeChannelRequest{ChannelId: req.ChannelID, Reason: req.Reason})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel freezed, id: "+req.ChannelID, resp)

}

func (h *AdminHandler) UnfreezeChannel(c *fiber.Ctx) error {
	var req struct {
		ChannelID string `json:"channelId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.UnfreezeChannel(ctx, &adminpb.UnfreezeChannelRequest{ChannelId: req.ChannelID})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel unfreezed, id: "+req.ChannelID, resp)

}

func (h *AdminHandler) DeleteChannel(c *fiber.Ctx) error {
	var req struct {
		ChannelID string `json:"channelId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}
	log.Println("chan:",req.ChannelID)

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.DeleteChannel(ctx, &adminpb.DeleteChannelRequest{ChannelId: req.ChannelID})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel deleted, id: "+req.ChannelID, resp)

}

func (h *AdminHandler) ListAllFiles(c *fiber.Ctx) error {

	adminID := c.Locals("userID").(string)
	log.Println("id:",adminID)
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.AdminListAllFiles(ctx, &adminpb.AdminListAllFilesRequest{AdminId: adminID})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "all files", resp)

}

func (h *AdminHandler) DeleteFile(c *fiber.Ctx) error {
	var req struct {
		FileID string `json:"fileId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	adminID := c.Locals("userId").(string)
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.AdminDeleteFile(ctx, &adminpb.AdminDeleteFileRequest{FileId: req.FileID, AdminId: adminID})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "file deleted", resp)

}

func (h *AdminHandler) BlockUserUpload(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"userId"`
		Block  bool   `json:"block"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	adminID := c.Locals("userId").(string)
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.Admin.AdminBlockUserUpload(ctx, &adminpb.AdminBlockUserUploadRequest{AdminId: adminID, Block: req.Block, UserId: req.UserID})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "user status changed", resp)

}
