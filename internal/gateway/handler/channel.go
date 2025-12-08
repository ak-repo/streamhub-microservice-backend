package handler

import (
	"fmt"
	"strconv"

	"github.com/ak-repo/stream-hub/gen/channelpb"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type ChannelHandler struct {
	client channelpb.ChannelServiceClient
}

func NewChannelHandler(client channelpb.ChannelServiceClient) *ChannelHandler {
	return &ChannelHandler{client: client}
}

// -------------------- Leave Channel --------------------
func (h *ChannelHandler) LeaveChannel(c *fiber.Ctx) error {
	req := new(channelpb.RemoveMemberRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	req.RemovedBy = uid
	req.UserId = uid
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()
	resp, err := h.client.RemoveMember(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "leaved", resp)
}

// -------------------- Create Channel --------------------
func (h *ChannelHandler) CreateChannel(c *fiber.Ctx) error {
	req := new(channelpb.CreateChannelRequest)

	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.CreateChannel(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel created", resp)
}

// -------------------- List Channels --------------------
func (h *ChannelHandler) ListChannels(c *fiber.Ctx) error {
	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.ListUserChannels(ctx, &channelpb.ListUserChannelsRequest{UserId: uid})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channels", resp)
}

// -------------------- Get Channel --------------------
func (h *ChannelHandler) GetChannel(c *fiber.Ctx) error {
	channelID := c.Params("channelId")
	if channelID == "" {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.GetChannel(ctx, &channelpb.GetChannelRequest{ChannelId: channelID})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, fmt.Sprintf("channel id: %s", channelID), resp)
}

// -------------------- Delete Channel --------------------
func (h *ChannelHandler) DeleteChannel(c *fiber.Ctx) error {
	req := new(channelpb.DeleteChannelRequest)
	req.ChannelId = c.Query("channel_id")
	req.RequesterId = c.Query("requester_id")

	if req.ChannelId == "" || req.RequesterId == "" {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.DeleteChannel(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel deleted", resp)
}

// -------------------- List Members --------------------
func (h *ChannelHandler) ListMembers(c *fiber.Ctx) error {
	channelID := c.Params("channelId")
	if channelID == "" {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.ListMembers(ctx, &channelpb.ListMembersRequest{ChannelId: channelID})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "member list", resp)
}

// -------------------- List Messages --------------------
func (h *ChannelHandler) ListMessages(c *fiber.Ctx) error {
	channelID := c.Params("channelId")
	if channelID == "" {
		return response.InvalidReqBody(c)
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.ListMessages(ctx, &channelpb.ListMessagesRequest{
		ChannelId: channelID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "previous messages", resp)
}

// -------------------Request JOIN & INVITE ------------

func (h *ChannelHandler) SendInvite(c *fiber.Ctx) error {
	req := new(channelpb.SendInviteRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}
	if req.ChannelId == "" || req.TargetUserId == "" {
		return response.InvalidReqBody(c)
	}

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	req.SenderId = uid

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.SendInvite(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "Invite sented", resp)
}

func (h *ChannelHandler) SendJoin(c *fiber.Ctx) error {
	req := new(channelpb.SendJoinRequest)
	if err := c.BodyParser(req); err != nil || req.ChannelId == "" || req.UserId == "" {
		return response.InvalidReqBody(c)
	}
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.SendJoin(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "Invite sented", resp)
}

func (h *ChannelHandler) ListUserInvites(c *fiber.Ctx) error {
	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()
	resp, err := h.client.ListUserInvites(ctx, &channelpb.ListUserInvitesRequest{UserId: uid})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "user invites", resp)
}

func (h *ChannelHandler) ListChannelJoins(c *fiber.Ctx) error {
	channelID := c.Params("id")
	if channelID == "" {
		return response.InvalidReqBody(c)
	}
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()
	resp, err := h.client.ListChannelJoins(ctx, &channelpb.ListChannelJoinsRequest{ChannelId: channelID})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "channel join request", resp)
}

func (h *ChannelHandler) UpdateRequestStatus(c *fiber.Ctx) error {
	req := new(channelpb.UpdateRequestStatusRequest)
	if err := c.BodyParser(&req); err != nil || req.Status == "" || req.RequestId == "" {
		return response.InvalidReqBody(c)
	}
	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	req.UserId = uid
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()
	resp, err := h.client.UpdateRequestStatus(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "request updated", resp)
}



