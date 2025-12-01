package grpcserver

import (
	"context"
	"log"
	"time"

	pb "github.com/ak-repo/stream-hub/gen/filespb"
	"github.com/ak-repo/stream-hub/internal/files_service/port"
	"github.com/ak-repo/stream-hub/pkg/logger"
)

type FileServer struct {
	pb.UnimplementedFileServiceServer
	service port.FileService
}

func NewFileServer(svc port.FileService) *FileServer {
	return &FileServer{service: svc}
}

func (s *FileServer) GenerateUploadURL(ctx context.Context, req *pb.GenerateUploadURLRequest) (*pb.GenerateUploadURLResponse, error) {
	uploadURL, storagePath, fileID, err := s.service.GenerateUploadURL(
		ctx,
		req.OwnerId,
		req.ChannelId,
		req.Filename,
		req.Size,
		req.MimeType,
		req.IsPublic,
	)

	if err != nil {
		return nil, err
	}

	logger.Log.Info("upload URL generated: " + uploadURL)

	return &pb.GenerateUploadURLResponse{
		UploadUrl:   uploadURL,
		StoragePath: storagePath,
		FileId:      fileID,
	}, nil
}

func (s *FileServer) ConfirmUpload(ctx context.Context, req *pb.ConfirmUploadRequest) (*pb.ConfirmUploadResponse, error) {
	f, err := s.service.ConfirmUpload(ctx, req.FileId)
	if err != nil {
		return nil, err
	}

	return &pb.ConfirmUploadResponse{
		File: &pb.File{
			Id:          f.ID,
			OwnerId:     f.OwnerID,
			ChannelId:   f.ChannelID,
			Filename:    f.Filename,
			Size:        f.Size,
			MimeType:    f.MimeType,
			StoragePath: f.StoragePath,
			IsPublic:    f.IsPublic,
			CreatedAt:   f.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *FileServer) GenerateDownloadURL(ctx context.Context, req *pb.GenerateDownloadURLRequest) (*pb.GenerateDownloadURLResponse, error) {
	log.Println("req: ", req.FileId, " +", req.RequesterId)

	exp := int64(300) // default
	if req.ExpireSeconds > 0 {
		exp = req.ExpireSeconds
	}

	url, err := s.service.GenerateDownloadURL(ctx, req.FileId, req.RequesterId, exp)
	if err != nil {
		return nil, err
	}

	return &pb.GenerateDownloadURLResponse{
		DownloadUrl:   url,
		ExpireSeconds: exp,
	}, nil
}

func (s *FileServer) ListFiles(ctx context.Context, req *pb.FileListRequest) (*pb.FileListResponse, error) {
	files, err := s.service.ListFiles(ctx, req.RequesterId, req.ChannelId)
	if err != nil {
		return nil, err
	}

	resp := &pb.FileListResponse{}

	for _, f := range files {
		resp.Files = append(resp.Files, &pb.File{
			Id:          f.ID,
			OwnerId:     f.OwnerID,
			ChannelId:   f.ChannelID,
			Filename:    f.Filename,
			Size:        f.Size,
			MimeType:    f.MimeType,
			StoragePath: f.StoragePath,
			IsPublic:    f.IsPublic,
			CreatedAt:   f.CreatedAt.Format(time.RFC3339),
		})
	}

	return resp, nil
}

func (s *FileServer) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {

	if err := s.service.DeleteFile(ctx, req.FileId, req.RequesterId); err != nil {
		return nil, err
	}

	return &pb.DeleteFileResponse{
		Success: true,
	}, nil
}
