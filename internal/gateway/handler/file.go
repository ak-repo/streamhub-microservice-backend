package handler

import (
	"context"
	"log"
	"time"

	"github.com/ak-repo/stream-hub/gen/filespb"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type FileHandler struct {
	client filespb.FileServiceClient
}

func NewFileHandler(cli filespb.FileServiceClient) *FileHandler {
	return &FileHandler{client: cli}
}

// -------------------- Generate Upload URL --------------------
func (h *FileHandler) CreateUploadUrl(c *fiber.Ctx) error {
	req := new(filespb.CreateUploadUrlRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.CreateUploadUrl(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "upload URL generated", resp)
}

// -------------------- Confirm Upload --------------------
func (h *FileHandler) CompleteUpload(c *fiber.Ctx) error {
	req := new(filespb.CompleteUploadRequest)
	if err := c.BodyParser(req); err != nil || req.FileId == "" {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.CompleteUpload(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "file uploaded successfully", resp)
}

// -------------------- Generate Download URL --------------------
func (h *FileHandler) GenerateDownloadURL(c *fiber.Ctx) error {
	req := new(filespb.GetDownloadUrlRequest)
	req.FileId = c.Query("file_id")
	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	req.RequesterId = uid

	if req.FileId == "" || req.RequesterId == "" {
		return response.InvalidReqBody(c)
	}
	if req.ExpireSeconds <= 0 {
		req.ExpireSeconds = 240
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetDownloadUrl(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "download URL generated", resp)
}

// -------------------- List Files --------------------
func (h *FileHandler) ListFiles(c *fiber.Ctx) error {
	req := new(filespb.ListFilesRequest)
	req.ChannelId = c.Query("channel_id")
	req.Limit = helper.StringToInt32(c.Query("limit"))
	req.Offset = helper.StringToInt32(c.Query(""))

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	req.RequesterId = uid

	if req.ChannelId == "" || req.RequesterId == "" {
		log.Println("req: ", req.RequesterId, " file:", req.ChannelId)

		return response.InvalidReqBody(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.ListFiles(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "file list retrieved", resp)
}

// -------------------- Delete File --------------------
func (h *FileHandler) DeleteFile(c *fiber.Ctx) error {
	req := new(filespb.DeleteFileRequest)

	req.FileId = c.Query("file_id")
	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	req.RequesterId = uid

	if req.FileId == "" || req.RequesterId == "" {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.DeleteFile(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "file deleted successfully", resp)
}
