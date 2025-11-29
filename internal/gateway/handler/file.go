package handler

import (
	"context"
	"log"
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
		OwnerID   string `json:"ownerId"`
		Filename  string `json:"filename"`
		Size      int64  `json:"size"`
		MimeType  string `json:"mimeType"`
		IsPublic  bool   `json:"isPublic"`
		ChannelID string `json:"channelId"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid body")
	}

	log.Println("on handlerL:", req.ChannelID)
	stCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GenerateUploadURL(stCtx, &filespb.GenerateUploadURLRequest{
		OwnerId:   req.OwnerID,
		ChannelId: req.ChannelID,
		Filename:  req.Filename,
		Size:      req.Size,
		MimeType:  req.MimeType,
		IsPublic:  req.IsPublic,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}
	return response.Success(ctx, "url generated", resp)
}

func (h *FileHandler) ConfirmUpload(ctx *fiber.Ctx) error {
	var req struct {
		FileID string `json:"fileId"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid body")

	}

	stCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.ConfirmUpload(stCtx, &filespb.ConfirmUploadRequest{FileId: req.FileID})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}
	return response.Success(ctx, "file uploaded successfully", resp)
}
func (h *FileHandler) GenerateDownloadURL(ctx *fiber.Ctx) error {
	req := struct {
		FileID        string `query:"fileId"`
		ExpireSeconds int64  `query:"expireSeconds"`
		RequesterID   string `query:"requesterId"`
	}{}

	if err := ctx.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid query params")
	}

	if req.FileID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "fileId is required")
	}
	if req.RequesterID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "requesterId is required")
	}
	if req.ExpireSeconds <= 0 {
		req.ExpireSeconds = 60
	}

	cliCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GenerateDownloadURL(cliCtx, &filespb.GenerateDownloadURLRequest{
		FileId:        req.FileID,
		ExpireSeconds: req.ExpireSeconds,
		RequesterId:   req.RequesterID,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}

	return response.Success(ctx, "generated download link", resp)
}


func (h *FileHandler) ListFiles(ctx *fiber.Ctx) error {
	requesterID := ctx.Query("requesterId")
	channelID := ctx.Query("channelId")
	if requesterID == "" || channelID == "" {

		return fiber.NewError(fiber.StatusBadRequest, "Missing query parameters")
	}

	cliCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.ListFiles(cliCtx, &filespb.FileListRequest{
		RequesterId: requesterID,
		ChannelId:   channelID,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}
	return response.Success(ctx, "lists", resp)
}

func (h *FileHandler) DeleteFile(ctx *fiber.Ctx) error {
	var req struct {
		FileID      string `json:"fileId"`
		RequesterID string `json:"requesterId"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid body")
	}

	cliCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.DeleteFile(cliCtx, &filespb.DeleteFileRequest{
		FileId:      req.FileID,
		RequesterId: req.RequesterID,
	})
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return ctx.Status(code).JSON(body)
	}
	return response.Success(ctx, "deleted", resp)
}

//GRPC

// type FileListRequest struct {
//     state         protoimpl.MessageState `protogen:"open.v1"`
//     RequesterId   string                 `protobuf:"bytes,1,opt,name=requester_id,json=requesterId,proto3" json:"requester_id,omitempty"` // Files uploaded by the user
//     ChannelId     string                 `protobuf:"bytes,2,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`       // Optional: files in a channel
//     unknownFields protoimpl.UnknownFields
//     sizeCache     protoimpl.SizeCache
// }
// ------------------------------ LIST FILES FOR USER OR CHANNEL ------------------------------

// func (*filespb.FileListRequest) Descriptor() ([]byte, []int)
// func (x *filespb.FileListRequest) GetChannelId() string
