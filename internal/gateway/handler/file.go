package handler

import (
	"context"
	"time"

	"github.com/ak-repo/stream-hub/gen/filespb"
	"github.com/ak-repo/stream-hub/pkg/errors"
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
func (h *FileHandler) GenerateUploadURL(c *fiber.Ctx) error {
	var req struct {
		OwnerID   string `json:"ownerId"`
		Filename  string `json:"filename"`
		Size      int64  `json:"size"`
		MimeType  string `json:"mimeType"`
		IsPublic  bool   `json:"isPublic"`
		ChannelID string `json:"channelId"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GenerateUploadURL(ctx, &filespb.GenerateUploadURLRequest{
		OwnerId:   req.OwnerID,
		ChannelId: req.ChannelID,
		Filename:  req.Filename,
		Size:      req.Size,
		MimeType:  req.MimeType,
		IsPublic:  req.IsPublic,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "upload URL generated", resp)
}

// -------------------- Confirm Upload --------------------
func (h *FileHandler) ConfirmUpload(c *fiber.Ctx) error {
	var req struct {
		FileID string `json:"fileId"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.ConfirmUpload(ctx, &filespb.ConfirmUploadRequest{
		FileId: req.FileID,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "file uploaded successfully", resp)
}

// -------------------- Generate Download URL --------------------
func (h *FileHandler) GenerateDownloadURL(c *fiber.Ctx) error {
	var req struct {
		FileID        string `query:"fileId"`
		ExpireSeconds int64  `query:"expireSeconds"`
		RequesterID   string `query:"requesterId"`
	}

	if err := c.QueryParser(&req); err != nil {
		return response.InvalidReqBody(c)
	}

	if req.FileID == "" {
		return response.InvalidReqBody(c)
	}
	if req.RequesterID == "" {
		return response.InvalidReqBody(c)
	}
	if req.ExpireSeconds <= 0 {
		req.ExpireSeconds = 60
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GenerateDownloadURL(ctx, &filespb.GenerateDownloadURLRequest{
		FileId:        req.FileID,
		RequesterId:   req.RequesterID,
		ExpireSeconds: req.ExpireSeconds,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "download URL generated", resp)
}

// -------------------- List Files --------------------
func (h *FileHandler) ListFiles(c *fiber.Ctx) error {
	requesterID := c.Query("requesterId")
	channelID := c.Query("channelId")

	if requesterID == "" || channelID == "" {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.ListFiles(ctx, &filespb.FileListRequest{
		RequesterId: requesterID,
		ChannelId:   channelID,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "file list retrieved", resp)
}

// -------------------- Delete File --------------------
func (h *FileHandler) DeleteFile(c *fiber.Ctx) error {
	fileID := c.Query("fileId")
	requesterID := c.Query("requesterId")

	if fileID == "" || requesterID == "" {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.DeleteFile(ctx, &filespb.DeleteFileRequest{
		FileId:      fileID,
		RequesterId: requesterID,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "file deleted successfully", resp)
}
