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

	req := new(authpb.AdminListUsersRequest)
	req.Pagination = &authpb.PaginationRequest{
		Limit:  helper.StringToInt32(c.Query("limit", "10")),
		Offset: helper.StringToInt32(c.Query("offset", "0")),
	}
	req.FilterQuery = c.Query("filter")

	log.Println("limit: ", req.Pagination.Limit, "off:", req.Pagination.Offset)

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}

	req.AdminId = uid

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()
	resp, err := h.clients.AdminAuth.AdminListUsers(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "listing users", resp)
}

func (h *AdminHandler) BanUser(c *fiber.Ctx) error {
	req := new(authpb.AdminBanUserRequest)

	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}

	req.AdminId = uid

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminAuth.AdminBanUser(ctx, req)

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel created", resp)
}

func (h *AdminHandler) UnbanUser(c *fiber.Ctx) error {
	req := new(authpb.AdminUnbanUserRequest)

	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}
	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}

	req.AdminId = uid

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminAuth.AdminUnbanUser(ctx, req)

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "user banned", resp)
}

func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	req := new(authpb.AdminDeleteUserRequest)
	req.TargetUserId = c.Params("target_user_id")

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}

	req.AdminId = uid

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminAuth.AdminDeleteUser(ctx, req)

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "user deleted", resp)
}

func (h *AdminHandler) UpdateRole(c *fiber.Ctx) error {
	req := new(authpb.AdminUpdateRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}

	req.AdminId = uid
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminAuth.AdminUpdateRole(ctx, req)

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
	req := new(channelpb.AdminFreezeChannelRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminChannel.AdminFreezeChannel(ctx, req)

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel freezed", resp)

}

func (h *AdminHandler) DeleteChannel(c *fiber.Ctx) error {
	req := new(channelpb.AdminDeleteChannelRequest)
	req.ChannelId = c.Params("channel_id")

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	req.AdminId = uid

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminChannel.AdminDeleteChannel(ctx, req)

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel deleted", resp)

}

// ------------------------------------- Files ---------------------
func (h *AdminHandler) ListAllFiles(c *fiber.Ctx) error {

	req := new(filespb.AdminListFilesRequest)
	req.Limit = helper.StringToInt32(c.Query("limit", "10"))
	req.Offset = helper.StringToInt32(c.Query("offset", "0"))

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	log.Println("offff:ss ", req.Limit)

	req.AdminId = uid
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminFile.AdminListFiles(ctx, req)

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "all files", resp)

}

func (h *AdminHandler) DeleteFile(c *fiber.Ctx) error {
	req := new(filespb.AdminDeleteFileRequest)
	req.FileId = c.Params("id")

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}

	req.AdminId = uid
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminFile.AdminDeleteFile(ctx, req)

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "file deleted", resp)

}

func (h *AdminHandler) BlockUserUpload(c *fiber.Ctx) error {
	req := new(filespb.AdminBlockUploadsRequest)

	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}
	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}

	req.AdminId = uid
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.clients.AdminFile.AdminBlockUploads(ctx, req)

	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "user status changed", resp)

}
