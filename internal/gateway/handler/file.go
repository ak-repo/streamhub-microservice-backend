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

func (h *FileHandler) GenerateUploadURL(ctx *fiber.Ctx) error {
	var req struct {
		OwnerID  string `json:"owner_id"`
		Filename string `json:"filename"`
		Size     int64  `json:"size"`
		MimeType string `json:"mime_type"`
		IsPublic bool   `json:"is_public"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}
	stctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.client.GenerateUploadURL(stctx, &filespb.GenerateUploadURLRequest{
		OwnerId:  req.OwnerID,
		Filename: req.Filename,
		Size:     req.Size,
		MimeType: req.MimeType,
		IsPublic: req.IsPublic,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}
	return response.Success(ctx, "url generated", resp)
}

func (h *FileHandler) ConfirmUpload(ctx *fiber.Ctx) error {
	var req struct {
		FileID string `json:"file_id"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}
	stctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.client.ConfirmUpload(stctx, &filespb.ConfirmUploadRequest{FileId: req.FileID})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}
	return response.Success(ctx, "file uploaded successfully", resp)
}

func (h *FileHandler) GenerateDownloadURL(ctx *fiber.Ctx) error {
	var req struct {
		FileID        string `json:"file_id"`
		ExpireSeconds int64  `json:"expire_seconds"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}
	clictx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.client.GenerateDownloadURL(clictx, &filespb.GenerateDownloadURLRequest{FileId: req.FileID, ExpireSeconds: req.ExpireSeconds})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}
	return response.Success(ctx, "generated download link", resp)
}

func (h *FileHandler) ListFiles(ctx *fiber.Ctx) error {
	owner := ctx.Params("owner_id")
	clictx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.client.ListFiles(clictx, &filespb.FileListRequest{OwnerId: owner})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}
	return response.Success(ctx, "lists", resp)
}

func (h *FileHandler) DeleteFile(ctx *fiber.Ctx) error {
	id := ctx.Params("file_id")
	clictx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.client.DeleteFile(clictx, &filespb.DeleteFileRequest{FileId: id,
		OwnerId: "122adf10-3398-46c8-bc8c-86ca83f4a177",
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}
	return response.Success(ctx, "deleted", resp)
}
