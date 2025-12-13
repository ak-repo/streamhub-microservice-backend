package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
	"github.com/ak-repo/stream-hub/internal/files_service/port"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/grpc/clients"
	"github.com/google/uuid"
)

type fileService struct {
	repo    port.FileRepository
	redis   port.TempFileStore
	store   port.FileStorage
	ttl     time.Duration
	clients clients.Clients
}

func NewFileService(
	repo port.FileRepository,
	redis port.TempFileStore,
	st port.FileStorage,
	ttl time.Duration,
	clients clients.Clients,
) port.FileService {
	return &fileService{
		repo:    repo,
		redis:   redis,
		store:   st,
		ttl:     ttl,
		clients: clients,
	}
}

// =============================================================================
// USER OPERATIONS
// =============================================================================

func (s *fileService) GenerateUploadURL(ctx context.Context, ownerID, channelID, filename string, size int64, mime string, isPublic bool) (string, string, string, error) {
	// 1. Check if user is blocked
	blocked, err := s.repo.IsUserBlocked(ctx, ownerID)
	if err != nil {
		return "", "", "", errors.New(errors.CodeInternal, "failed to check block status", err)
	}
	if blocked {
		return "", "", "", errors.New(errors.CodeForbidden, "upload blocked for this user", nil)
	}

	// 2. Check storage limits
	used, limit, err := s.repo.GetStorageUsage(ctx, ownerID)
	log.Println("limi: ",limit," used: ",used)
	if err != nil {
		return "", "", "", errors.New(errors.CodeInternal, "failed to check storage usage", err)
	}
	if limit > 0 && (used+size) > limit {
		return "", "", "", errors.New(errors.CodeForbidden, "storage limit exceeded", nil)
	}

	// 3. Check Channel Membership
	if channelID != "" {
		isMember, err := s.repo.IsChannelMember(ctx, channelID, ownerID)
		if err != nil {
			return "", "", "", errors.New(errors.CodeInternal, "membership check failed", err)
		}
		if !isMember {
			return "", "", "", errors.New(errors.CodeForbidden, "not a member of channel", nil)
		}
	}

	// 4. Prepare Metadata
	fileID := uuid.New().String()
	storagePath := fmt.Sprintf("uploads/%s/%s", ownerID, filename)

	f := &domain.File{
		ID:          fileID,
		OwnerID:     ownerID,
		ChannelID:   channelID,
		Filename:    filename,
		Size:        size,
		MimeType:    mime,
		StoragePath: storagePath,
		IsPublic:    isPublic,
		CreatedAt:   time.Now().UTC(),
	}

	// 5. Save Temp & Generate URL
	if err := s.redis.SaveTemp(ctx, f); err != nil {
		return "", "", "", errors.New(errors.CodeInternal, "failed to save temp metadata", err)
	}

	url, err := s.store.GenerateUploadURL(f)
	if err != nil {
		s.redis.DeleteTemp(ctx, fileID)
		return "", "", "", errors.New(errors.CodeInternal, "failed to generate signed URL", err)
	}

	return url, storagePath, fileID, nil
}

func (s *fileService) ConfirmUpload(ctx context.Context, fileID string) (*domain.File, error) {
	f, err := s.redis.GetTemp(ctx, fileID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "upload session expired or invalid", err)
	}

	if err := s.repo.Save(ctx, f); err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to persist file metadata", err)
	}

	s.redis.DeleteTemp(ctx, fileID)
	return f, nil
}

func (s *fileService) GenerateDownloadURL(ctx context.Context, fileID, requesterID string, expirySeconds int64) (string, error) {
	f, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return "", errors.New(errors.CodeNotFound, "file not found", err)
	}

	// Access Control
	if f.ChannelID != "" {
		isMember, err := s.repo.IsChannelMember(ctx, f.ChannelID, requesterID)
		if err != nil {
			return "", errors.New(errors.CodeInternal, "membership check failed", err)
		}
		if !isMember {
			return "", errors.New(errors.CodeForbidden, "access denied", nil)
		}
	} else if !f.IsPublic && f.OwnerID != requesterID {
		return "", errors.New(errors.CodeForbidden, "access denied", nil)
	}

	return s.store.GenerateDownloadURL(f, expirySeconds)
}

func (s *fileService) ListFiles(ctx context.Context, requesterID, channelID string) ([]*domain.File, error) {
	// Simple access check for channel files
	if channelID != "" {
		isMember, err := s.repo.IsChannelMember(ctx, channelID, requesterID)
		if err != nil {
			return nil, errors.New(errors.CodeInternal, "membership check failed", err)
		}
		if !isMember {
			return nil, errors.New(errors.CodeForbidden, "access denied", nil)
		}
	}
	return s.repo.GetByChannel(ctx, channelID)
}

func (s *fileService) DeleteFile(ctx context.Context, fileID, requesterID string) error {
	f, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return errors.New(errors.CodeNotFound, "file not found", err)
	}

	// Check if requester is owner OR admin
	isOwner := f.OwnerID == requesterID
	isAdmin, _ := s.repo.IsChannelAdmin(ctx, f.ChannelID, requesterID)

	if isOwner || isAdmin {
		// Delete from Object Storage
		if err := s.store.DeleteObject(f); err != nil {
			return errors.New(errors.CodeInternal, "failed to delete from storage", err)
		}

		return s.repo.Delete(ctx, fileID)
	}
	return errors.New(errors.CodeForbidden, "permission denied", nil)

}

// for channel
func (s *fileService) GetStorageUsage(ctx context.Context, channelID string) (int64, int64, error) {
	return s.repo.GetStorageUsage(ctx, channelID)
}


// =============================================================================
// ADMIN OPERATIONS
// =============================================================================

func (s *fileService) AdminListFiles(ctx context.Context, limit, offset int32) ([]*domain.File, error) {
	log.Println("list files: ", limit, "off: ", offset)
	if limit >= 0 {
		limit = 10
	}
	return s.repo.ListAllFiles(ctx, limit, offset)
}

func (s *fileService) AdminDeleteFile(ctx context.Context, fileID, adminID string, force bool) error {
	f, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return errors.New(errors.CodeNotFound, "file not found", err)
	}

	// Delete from Object Storage
	if err := s.store.DeleteObject(f); err != nil {
		return errors.New(errors.CodeInternal, "failed to delete from storage", err)
	}
	return s.repo.Delete(ctx, fileID)
}

func (s *fileService) AdminSetStorageLimit(ctx context.Context, channelID string, maxBytes int64) (int64, error) {
	_, currentLimit, _ := s.repo.GetStorageUsage(ctx, channelID)
	if err := s.repo.SetStorageLimit(ctx, channelID, maxBytes); err != nil {
		return 0, errors.New(errors.CodeInternal, "failed to set limit", err)
	}
	return currentLimit, nil
}

func (s *fileService) AdminBlockUploads(ctx context.Context, targetID string, block bool) error {
	return s.repo.SetUserBlocked(ctx, targetID, block)
}

func (s *fileService) AdminGetStats(ctx context.Context) (*domain.StorageStats, error) {
	return s.repo.GetGlobalStats(ctx)
}
