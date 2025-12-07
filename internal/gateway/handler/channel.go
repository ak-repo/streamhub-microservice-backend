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

// -------------------- Create Channel --------------------
func (h *ChannelHandler) CreateChannel(c *fiber.Ctx) error {
	var req struct {
		Name      string `json:"name"`
		CreatorID string `json:"creatorId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.CreateChannel(ctx, &channelpb.CreateChannelRequest{
		Name:      req.Name,
		CreatorId: req.CreatorID,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "channel created", resp)
}

// -------------------- Join Channel --------------------
func (h *ChannelHandler) JoinChannel(c *fiber.Ctx) error {
	var req struct {
		ChannelID string `json:"channelId"`
		UserID    string `json:"userId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)

	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.AddMember(ctx, &channelpb.AddMemberRequest{
		ChannelId: req.ChannelID,
		UserId:    req.UserID,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)

	}

	return response.Success(c, fmt.Sprintf("joined channel: %s", req.ChannelID), resp)
}

// -------------------- Leave Channel --------------------
func (h *ChannelHandler) LeaveChannel(c *fiber.Ctx) error {
	var req struct {
		ChannelID string `json:"channelId"`
		UserID    string `json:"userId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.RemoveMember(ctx, &channelpb.RemoveMemberRequest{
		ChannelId: req.ChannelID,
		UserId:    req.UserID,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, fmt.Sprintf("left channel: %s", req.ChannelID), resp)
}

// -------------------- List Channels --------------------
func (h *ChannelHandler) ListChannels(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	if userID == "" {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.ListChannels(ctx, &channelpb.ListChannelsRequest{UserId: userID})
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
	channelID := c.Query("channelId")
	requesterID := c.Query("requesterId")

	if channelID == "" || requesterID == "" {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.DeleteChannel(ctx, &channelpb.DeleteChannelRequest{
		ChannelId:   channelID,
		RequesterId: requesterID,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, fmt.Sprintf("channel deleted: %s", channelID), resp)
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

func (h *ChannelHandler) AddMember(c *fiber.Ctx) error {
	var req struct {
		ChannelID string `json:"channelId"`
		UserID    string `json:"userId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)

	}

	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.AddMember(ctx, &channelpb.AddMemberRequest{
		ChannelId: req.ChannelID,
		UserId:    req.UserID,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)

	}

	return response.Success(c, fmt.Sprintf("joined channel: %s", req.ChannelID), resp)

}

// -------------------Request JOIN & INVITE ------------
func (h *ChannelHandler) SendInvite(c *fiber.Ctx) error {
	var req struct {
		ChannelID string `json:"channelId"`
		UserID    string `json:"userId"`
	}
	if err := c.BodyParser(&req); err != nil || req.ChannelID == "" || req.UserID == "" {
		return response.InvalidReqBody(c)
	}
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.SendInvite(ctx, &channelpb.SendInviteRequest{UserId: req.UserID, ChannelId: req.ChannelID})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "Invite sented", resp)
}

func (h *ChannelHandler) SendJoin(c *fiber.Ctx) error {
	var req struct {
		ChannelID string `json:"channelId"`
		UserID    string `json:"userId"`
	}
	if err := c.BodyParser(&req); err != nil || req.ChannelID == "" || req.UserID == "" {
		return response.InvalidReqBody(c)
	}
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()

	resp, err := h.client.SendJoin(ctx, &channelpb.SendJoinRequest{UserId: req.UserID, ChannelId: req.ChannelID})
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
	resp, err := h.client.ListUserInvites(ctx, &channelpb.ListUserInviteRequest{UserId: uid})
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
	resp, err := h.client.ListChannelJoins(ctx, &channelpb.ListChannelJoinRequest{ChannelId: channelID})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "channel join request", resp)
}

func (h *ChannelHandler) UpdateRequestStatus(c *fiber.Ctx) error {
	var req struct {
		ReqID  string `json:"reqId"`
		Status string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil || req.Status == "" || req.ReqID == "" {
		return response.InvalidReqBody(c)
	}
	ctx, cancel := helper.WithGRPCTimeout()
	defer cancel()
	resp, err := h.client.UpdateRequestStatus(ctx, &channelpb.StatusUpdateRequest{Id: req.ReqID, Status: req.Status})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "request updated", resp)
}
