package grpc

import (
	"context"
	"log"
	"time"

	"github.com/ak-repo/stream-hub/gen/filespb"
	"github.com/ak-repo/stream-hub/internal/files_service/domain"
	"github.com/ak-repo/stream-hub/internal/files_service/port"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"go.uber.org/zap"
)

// Server implements both FileService and AdminFileService
type Server struct {
	filespb.UnimplementedFileServiceServer
	filespb.UnimplementedAdminFileServiceServer
	service port.FileService
}

func NewServer(svc port.FileService) *Server {
	return &Server{service: svc}
}

// =============================================================================
// PUBLIC FILE SERVICE
// =============================================================================

func (s *Server) CreateUploadUrl(ctx context.Context, req *filespb.CreateUploadUrlRequest) (*filespb.CreateUploadUrlResponse, error) {
	url, _, fileID, err := s.service.GenerateUploadURL(
		ctx, req.OwnerId, req.ChannelId, req.Filename,
		req.SizeBytes, req.MimeType, req.IsPublic,
	)
	if err != nil {
		return nil, err
	}

	logger.Log.Info("upload initiated", zap.String("file_id", fileID), zap.String("user", req.OwnerId))

	return &filespb.CreateUploadUrlResponse{
		UploadUrl:     url,
		FileId:        fileID,
		ExpireSeconds: 3600, // Configurable in real app
	}, nil
}

func (s *Server) CompleteUpload(ctx context.Context, req *filespb.CompleteUploadRequest) (*filespb.CompleteUploadResponse, error) {
	// TODO success handling
	// if !req.Success {
	// 	logger.Log.Info("client reported upload failure", zap.String("file_id", req.FileId))
	// 	return &filespb.CompleteUploadResponse{},
	// }

	f, err := s.service.ConfirmUpload(ctx, req.FileId)
	if err != nil {
		return nil, err
	}

	return &filespb.CompleteUploadResponse{File: mapFileToProto(f)}, nil
}

func (s *Server) GetDownloadUrl(ctx context.Context, req *filespb.GetDownloadUrlRequest) (*filespb.GetDownloadUrlResponse, error) {
	seconds := req.ExpireSeconds
	if seconds <= 0 {
		seconds = 300 // default 5 mins
	}

	url, err := s.service.GenerateDownloadURL(ctx, req.FileId, req.RequesterId, seconds)
	if err != nil {
		return nil, err
	}

	return &filespb.GetDownloadUrlResponse{
		DownloadUrl:   url,
		ExpireSeconds: seconds,
	}, nil
}

func (s *Server) ListFiles(ctx context.Context, req *filespb.ListFilesRequest) (*filespb.ListFilesResponse, error) {
	// Currently ignoring page_token/size for user lists as per service logic (returning all)
	// You can enhance service later to support pagination here too.
	files, err := s.service.ListFiles(ctx, req.RequesterId, req.ChannelId)
	if err != nil {
		return nil, err
	}

	var pbFiles []*filespb.FileMetadata
	for _, f := range files {
		pbFiles = append(pbFiles, mapFileToProto(f))
	}

	return &filespb.ListFilesResponse{Files: pbFiles}, nil
}

func (s *Server) DeleteFile(ctx context.Context, req *filespb.DeleteFileRequest) (*filespb.DeleteFileResponse, error) {
	err := s.service.DeleteFile(ctx, req.FileId, req.RequesterId)
	if err != nil {
		return nil, err
	}
	return &filespb.DeleteFileResponse{Success: true}, nil
}

func (s *Server) GetStorageUsage(ctx context.Context, req *filespb.GetStorageUsageRequest) (*filespb.GetStorageUsageResponse, error) {
	used, limit, err := s.service.GetStorageUsage(ctx, req.ChannelId)
	if err != nil {
		return nil, err
	}

	return &filespb.GetStorageUsageResponse{
		UsedBytes:  used,
		LimitBytes: limit,
	}, nil
}



// =============================================================================
// ADMIN FILE SERVICE
// =============================================================================

func (s *Server) AdminListFiles(ctx context.Context, req *filespb.AdminListFilesRequest) (*filespb.AdminListFilesResponse, error) {
	files, err := s.service.AdminListFiles(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	log.Println("files: ", files)
	var pbFiles []*filespb.FileMetadata
	for _, f := range files {
		pbFiles = append(pbFiles, mapFileToProto(f))
	}

	return &filespb.AdminListFilesResponse{
		Files: pbFiles,
	}, nil
}

func (s *Server) AdminDeleteFile(ctx context.Context, req *filespb.AdminDeleteFileRequest) (*filespb.AdminDeleteFileResponse, error) {
	err := s.service.AdminDeleteFile(ctx, req.FileId, req.AdminId, req.ForceDelete)
	if err != nil {
		return nil, err
	}
	return &filespb.AdminDeleteFileResponse{Success: true}, nil
}

func (s *Server) AdminSetStorageLimit(ctx context.Context, req *filespb.AdminSetStorageLimitRequest) (*filespb.AdminSetStorageLimitResponse, error) {
	prev, err := s.service.AdminSetStorageLimit(ctx, req.TargetId, req.MaxBytes)
	if err != nil {
		return nil, err
	}
	return &filespb.AdminSetStorageLimitResponse{Success: true, PreviousLimit: prev}, nil
}

func (s *Server) AdminBlockUploads(ctx context.Context, req *filespb.AdminBlockUploadsRequest) (*filespb.AdminBlockUploadsResponse, error) {
	err := s.service.AdminBlockUploads(ctx, req.TargetUserId, req.Block)
	if err != nil {
		return nil, err
	}
	return &filespb.AdminBlockUploadsResponse{Success: true}, nil
}

func (s *Server) AdminGetStats(ctx context.Context, req *filespb.AdminGetStatsRequest) (*filespb.AdminGetStatsResponse, error) {
	stats, err := s.service.AdminGetStats(ctx)
	if err != nil {
		return nil, err
	}
	return &filespb.AdminGetStatsResponse{Stats: mapStatsToProto(stats)}, nil
}

// =============================================================================
// HELPERS
// =============================================================================

func mapFileToProto(f *domain.File) *filespb.FileMetadata {
	if f == nil {
		return nil
	}
	return &filespb.FileMetadata{
		Id:          f.ID,
		OwnerId:     f.OwnerID,
		ChannelId:   f.ChannelID,
		Filename:    f.Filename,
		SizeBytes:   f.Size,
		MimeType:    f.MimeType,
		StoragePath: f.StoragePath,
		IsPublic:    f.IsPublic,
		CreatedAt:   f.CreatedAt.Format(time.RFC3339),
		OwerName:    f.OwnerName,
		Channelname: f.ChannelName,
	}
}

func mapStatsToProto(s *domain.StorageStats) *filespb.StorageStats {
	if s == nil {
		return nil
	}
	return &filespb.StorageStats{
		TotalFilesCount:   s.TotalFilesCount,
		TotalStorageBytes: s.TotalStorageBytes,
		PublicFilesCount:  s.PublicFilesCount,
		PrivateFilesCount: s.PrivateFilesCount,
	}
}
