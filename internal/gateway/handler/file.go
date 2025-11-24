package handler

import (
	"context"
	"time"

	"github.com/ak-repo/stream-hub/gen/filespb"
	"github.com/gofiber/fiber/v2"
)

type FileHandler struct {
	client filespb.FileServiceClient
}

func NewFileHandler(cli filespb.FileServiceClient) *FileHandler {
	return &FileHandler{client: cli}
}

func (h *FileHandler) GenerateUploadURL(c *fiber.Ctx) error {
	var body struct {
		OwnerID  string `json:"owner_id"`
		Filename string `json:"filename"`
		Size     int64  `json:"size"`
		MimeType string `json:"mime_type"`
		IsPublic bool   `json:"is_public"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.client.GenerateUploadURL(ctx, &filespb.GenerateUploadURLRequest{
		OwnerId:  body.OwnerID,
		Filename: body.Filename,
		Size:     body.Size,
		MimeType: body.MimeType,
		IsPublic: body.IsPublic,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(resp)
}

func (h *FileHandler) ConfirmUpload(c *fiber.Ctx) error {
	var body struct {
		FileID string `json:"file_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.client.ConfirmUpload(ctx, &filespb.ConfirmUploadRequest{FileId: body.FileID})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(resp.File)
}

func (h *FileHandler) GenerateDownloadURL(c *fiber.Ctx) error {
	var body struct {
		FileID        string `json:"file_id"`
		ExpireSeconds int64  `json:"expire_seconds"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.client.GenerateDownloadURL(ctx, &filespb.GenerateDownloadURLRequest{FileId: body.FileID, ExpireSeconds: body.ExpireSeconds})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(resp)
}

func (h *FileHandler) ListFiles(c *fiber.Ctx) error {
	owner := c.Params("owner_id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.client.ListFiles(ctx, &filespb.FileListRequest{OwnerId: owner})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(resp.Files)
}

func (h *FileHandler) DeleteFile(c *fiber.Ctx) error {
	id := c.Params("file_id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := h.client.DeleteFile(ctx, &filespb.DeleteFileRequest{FileId: id})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendString("deleted")
}
