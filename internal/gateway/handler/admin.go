package handler

import (
	"log"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/authpb"
	"github.com/ak-repo/stream-hub/gen/channelpb"
	"github.com/ak-repo/stream-hub/gen/filespb"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/grpc/clients"
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
	resp, err := h.clients.AdminAuth.AdminListUsers(ctx, &authpb.AdminListUsersRequest{FilterQuery: filterBy})
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

	resp, err := h.clients.AdminAuth.AdminBanUser(ctx, &authpb.AdminBanUserRequest{
		TargetUserId: req.UserID,
		Reason:       req.Reason,
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

	resp, err := h.clients.AdminAuth.AdminUnbanUser(ctx, &authpb.AdminUnbanUserRequest{
		TargetUserId: req.UserID,
		Reason:       req.Reason,
	})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel created", resp)
}

func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {

	// ctx, cancel := helper.WithGRPCTimeout()
	// defer cancel()

	// resp, err := h.clients.Admin.

	// if err != nil {
	// 	code, body := errors.GRPCToFiber(err)
	// 	return response.Error(c, code, body)
	// }
	return response.Success(c, "user deleted", nil)
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

	resp, err := h.clients.AdminAuth.AdminUpdateRole(ctx, &authpb.AdminUpdateRoleRequest{
		TargetUserId: req.UserID,
		NewRole:      req.Role,
	})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel created", resp)
}

// ----------------- channel actions ----------------------
func (h *AdminHandler) ListChannels(c *fiber.Ctx) error {
	req := new(channelpb.AdminListChannelsRequest)
	req.Limit = helper.StringToInt32(c.Query("limit", "10"))
	req.Offset = helper.StringToInt32(c.Query("offset", "0"))

	log.Println("limit: ", req.Limit, " offset: ", req.Offset)

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}

	req.AdminId = uid
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminChannel.AdminListChannels(ctx, req)

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

	resp, err := h.clients.AdminChannel.AdminFreezeChannel(ctx, &channelpb.AdminFreezeChannelRequest{ChannelId: req.ChannelID, Reason: req.Reason})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel freezed, id: "+req.ChannelID, resp)

}

func (h *AdminHandler) DeleteChannel(c *fiber.Ctx) error {
	channelID := c.Params("id")
	adminID := c.Locals("userID").(string)

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminChannel.AdminDeleteChannel(ctx, &channelpb.AdminDeleteChannelRequest{ChannelId: channelID, AdminId: adminID})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel deleted, id: "+channelID, resp)

}

func (h *AdminHandler) ListAllFiles(c *fiber.Ctx) error {

	adminID := c.Locals("userID").(string)
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminFile.AdminListFiles(ctx, &filespb.AdminListFilesRequest{AdminId: adminID})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "all files", resp)

}

func (h *AdminHandler) DeleteFile(c *fiber.Ctx) error {
	fileID := c.Params("id")
	adminID := c.Locals("userID").(string)

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminFile.AdminDeleteFile(ctx, &filespb.AdminDeleteFileRequest{FileId: fileID, AdminId: adminID})

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

	resp, err := h.clients.AdminFile.AdminBlockUploads(ctx, &filespb.AdminBlockUploadsRequest{AdminId: adminID, Block: req.Block, TargetUserId: req.UserID})

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "user status changed", resp)

}
