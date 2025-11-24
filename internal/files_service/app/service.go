package app

import (
	"context"
	"fmt"
	"time"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
	"github.com/ak-repo/stream-hub/internal/files_service/port"
	"github.com/google/uuid"
)

// fileService implements port.FileService
type fileService struct {
	repo port.FileRepository
	ts   port.TempFileStore
	st   port.FileStorage
	ttl  time.Duration
}

func NewFileService(repo port.FileRepository, ts port.TempFileStore, st port.FileStorage, ttl time.Duration) port.FileService {
	return &fileService{repo: repo, ts: ts, st: st, ttl: ttl}
}

// GenerateUploadURL: store temp metadata in Redis, return presigned PUT
func (s *fileService) GenerateUploadURL(ctx context.Context, ownerID, filename string, size int64, mime string, isPublic bool) (string, string, string, error) {
	fileID := uuid.NewString()
	key := fmt.Sprintf("uploads/%s_%s", fileID, filename)
	f := &domain.File{
		ID:          fileID,
		OwnerID:     ownerID,
		Filename:    filename,
		Size:        size,
		MimeType:    mime,
		StoragePath: key,
		IsPublic:    isPublic,
		CreatedAt:   time.Now().UTC(),
	}
	// save temp
	if err := s.ts.SaveTemp(ctx, f); err != nil {
		return "", "", "", err
	}
	// generate presigned url
	url, err := s.st.GenerateUploadURL(f)
	if err != nil {
		// clean temp on error
		_ = s.ts.DeleteTemp(ctx, fileID)
		return "", "", "", err
	}
	return url, key, fileID, nil
}

// ConfirmUpload: read temp metadata, save to DB, delete temp entry
func (s *fileService) ConfirmUpload(ctx context.Context, fileID string) (*domain.File, error) {
	f, err := s.ts.GetTemp(ctx, fileID)
	if err != nil {
		return nil, err
	}
	// Optional: you may stat object in S3 to validate size/content-type
	if err := s.repo.Save(ctx, f); err != nil {
		return nil, err
	}
	_ = s.ts.DeleteTemp(ctx, fileID)
	return f, nil
}

func (s *fileService) GenerateDownloadURL(ctx context.Context, fileID string, expirySeconds int64) (string, error) {
	f, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return "", err
	}
	// only allow owner or public; authorization is done in gateway by token
	return s.st.GenerateDownloadURL(f, expirySeconds)
}

func (s *fileService) ListFiles(ctx context.Context, ownerID string) ([]*domain.File, error) {
	return s.repo.GetByOwner(ctx, ownerID)
}

func (s *fileService) DeleteFile(ctx context.Context, fileID string) error {
	f, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return err
	}
	if err := s.st.DeleteObject(f); err != nil {
		return err
	}
	return s.repo.Delete(ctx, fileID)
}
