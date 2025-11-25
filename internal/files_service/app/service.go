package app

import (
	"context"
	"fmt"
	"time"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
	"github.com/ak-repo/stream-hub/internal/files_service/port"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/google/uuid"
)

// fileService implements port.FileService
type fileService struct {
	repo  port.FileRepository
	redis port.TempFileStore
	store port.FileStorage
	ttl   time.Duration
}

func NewFileService(repo port.FileRepository, redis port.TempFileStore, st port.FileStorage, ttl time.Duration) port.FileService {
	return &fileService{repo: repo, redis: redis, store: st, ttl: ttl}
}

// GenerateUploadURL creates a temp metadata entry in Redis and returns a presigned PUT URL
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

	if err := s.redis.SaveTemp(ctx, f); err != nil {
		return "", "", "", errors.New(errors.CodeInternal, fmt.Sprintf("failed to save temp metadata for file %s", fileID), err)
	}

	url, err := s.store.GenerateUploadURL(f)
	if err != nil {
		if delErr := s.redis.DeleteTemp(ctx, fileID); delErr != nil {
			fmt.Printf("warning: failed to cleanup temp metadata for file %s: %v\n", fileID, delErr)
		}
		return "", "", "", errors.New(errors.CodeInternal, fmt.Sprintf("failed to generate upload URL for file %s", fileID), err)
	}

	return url, key, fileID, nil
}

// ConfirmUpload saves file metadata to DB and deletes temp Redis entry
func (s *fileService) ConfirmUpload(ctx context.Context, fileID string) (*domain.File, error) {
	f, err := s.redis.GetTemp(ctx, fileID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, fmt.Sprintf("temp metadata missing for file %s", fileID), err)
	}

	if err := s.repo.Save(ctx, f); err != nil {
		return nil, errors.New(errors.CodeInternal, fmt.Sprintf("failed to save file metadata to DB for file %s", fileID), err)
	}

	if err := s.redis.DeleteTemp(ctx, fileID); err != nil {
		fmt.Printf("warning: failed to delete temp metadata for file %s: %v\n", fileID, err)
	}

	return f, nil
}

// GenerateDownloadURL returns a presigned GET URL if the file is public or owned by caller
func (s *fileService) GenerateDownloadURL(ctx context.Context, fileID, requesterID string, expirySeconds int64) (string, error) {
	f, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return "", errors.New(errors.CodeNotFound, fmt.Sprintf("file %s not found in DB", fileID), err)
	}

	// authorization: allow owner or public
	if !f.IsPublic && f.OwnerID != requesterID {
		return "", errors.New(errors.CodeForbidden, fmt.Sprintf("access denied for file %s", fileID), nil)
	}

	url, err := s.store.GenerateDownloadURL(f, expirySeconds)
	if err != nil {
		return "", errors.New(errors.CodeInternal, fmt.Sprintf("failed to generate download URL for file %s", fileID), err)
	}

	return url, nil
}

// ListFiles returns all files owned by a specific user
func (s *fileService) ListFiles(ctx context.Context, ownerID string) ([]*domain.File, error) {
	files, err := s.repo.GetByOwner(ctx, ownerID)
	if err != nil {
		return nil, errors.New(errors.CodeInternal, fmt.Sprintf("failed to list files for owner %s", ownerID), err)
	}
	return files, nil
}

// DeleteFile removes file from storage and DB, requires ownership
func (s *fileService) DeleteFile(ctx context.Context, fileID, requesterID string) error {
	f, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return errors.New(errors.CodeNotFound, fmt.Sprintf("file %s not found", fileID), err)
	}

	if f.OwnerID != requesterID {
		return errors.New(errors.CodeForbidden, fmt.Sprintf("access denied for deleting file %s", fileID), nil)
	}

	if err := s.store.DeleteObject(f); err != nil {
		return errors.New(errors.CodeInternal, fmt.Sprintf("failed to delete file object %s", fileID), err)
	}

	if err := s.repo.Delete(ctx, fileID); err != nil {
		return errors.New(errors.CodeInternal, fmt.Sprintf("failed to delete file metadata %s", fileID), err)
	}

	return nil
}
