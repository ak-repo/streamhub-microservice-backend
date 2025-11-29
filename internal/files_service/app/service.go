package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
	"github.com/ak-repo/stream-hub/internal/files_service/port"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/google/uuid"
)

type fileService struct {
	repo  port.FileRepository
	redis port.TempFileStore
	store port.FileStorage
	ttl   time.Duration
}

func NewFileService(repo port.FileRepository, redis port.TempFileStore, st port.FileStorage, ttl time.Duration) port.FileService {
	return &fileService{
		repo:  repo,
		redis: redis,
		store: st,
		ttl:   ttl,
	}
}

// GenerateUploadURL
func (s *fileService) GenerateUploadURL(ctx context.Context, ownerID, channelID, filename string, size int64, mime string, isPublic bool) (string, string, string, error) {

	if channelID != "" {
		isMember, err := s.repo.IsChannelMember(ctx, channelID, ownerID)
		if err != nil {
			return "", "", "", errors.New(errors.CodeInternal, "failed to check channel membership", err)
		}
		if !isMember {
			return "", "", "", errors.New(errors.CodeForbidden, "you are not a member of this channel", nil)
		}
	}

	fileID := uuid.New().String()
	key := fmt.Sprintf("uploads/%s_%s", fileID, filename)

	f := &domain.File{
		ID:          fileID,
		OwnerID:     ownerID,
		ChannelID:   channelID,
		Filename:    filename,
		Size:        size,
		MimeType:    mime,
		StoragePath: key,
		IsPublic:    isPublic,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.redis.SaveTemp(ctx, f); err != nil {
		return "", "", "", errors.New(errors.CodeInternal,
			fmt.Sprintf("failed to save temp metadata for file %s", fileID), err)
	}

	url, err := s.store.GenerateUploadURL(f)
	if err != nil {
		s.redis.DeleteTemp(ctx, fileID)
		return "", "", "", errors.New(errors.CodeInternal,
			fmt.Sprintf("failed to generate upload URL for file %s", fileID), err)
	}

	return url, key, fileID, nil
}

// ConfirmUpload
func (s *fileService) ConfirmUpload(ctx context.Context, fileID string) (*domain.File, error) {
	f, err := s.redis.GetTemp(ctx, fileID)

	if err != nil {
		return nil, errors.New(errors.CodeNotFound,
			fmt.Sprintf("temp metadata missing for file %s", fileID), err)
	}
	log.Println("channelid: confirm ", f.ChannelID)

	if err := s.repo.Save(ctx, f); err != nil {
		return nil, errors.New(errors.CodeInternal,
			fmt.Sprintf("failed to save file metadata to DB for file %s", fileID), err)
	}

	s.redis.DeleteTemp(ctx, fileID)
	return f, nil
}

// GenerateDownloadURL
func (s *fileService) GenerateDownloadURL(ctx context.Context, fileID, requesterID string, expirySeconds int64) (string, error) {
	log.Println("fie:", fileID, " req:", requesterID)
	f, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return "", errors.New(errors.CodeNotFound,
			fmt.Sprintf("file %s not found", fileID), err)
	}

	if f.ChannelID != "" {
		isMember, err := s.repo.IsChannelMember(ctx, f.ChannelID, requesterID)
		if err != nil {
			return "", errors.New(errors.CodeInternal, "failed to check channel membership", err)
		}
		if !isMember {
			return "", errors.New(errors.CodeForbidden, "you are not a member of this channel", nil)
		}
	}

	url, err := s.store.GenerateDownloadURL(f, expirySeconds)
	if err != nil {
		return "", errors.New(errors.CodeInternal,
			fmt.Sprintf("failed to generate download URL for file %s", fileID), err)
	}

	return url, nil
}

// ListFiles
func (s *fileService) ListFiles(ctx context.Context, requesterID, channelID string) ([]*domain.File, error) {
	isMember, err := s.repo.IsChannelMember(ctx, channelID, requesterID)
	if err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to check channel membership", err)
	}
	if !isMember {
		return nil, errors.New(errors.CodeForbidden, "you are not allowed to access this file", nil)
	}

	files, err := s.repo.GetByChannel(ctx, channelID)
	log.Println("files:", files)
	if err != nil {
		return nil, errors.New(errors.CodeInternal,
			fmt.Sprintf("failed to list files for channels %s", channelID), err)
	}
	return files, nil
}

// DeleteFile
func (s *fileService) DeleteFile(ctx context.Context, fileID, requesterID string) error {
	f, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return errors.New(errors.CodeNotFound,
			fmt.Sprintf("file %s not found", fileID), err)
	}

	// Only owner or channel admin can delete
	// isAdmin, err := s.repo.IsChannelAdmin(ctx, f.ChannelID, requesterID)
	// if err != nil {
	// 	return errors.New(errors.CodeInternal, "failed to check admin access", err)
	// }

	if requesterID != f.OwnerID {
		return errors.New(errors.CodeForbidden,
			"only owner or channel admin can delete this file", nil)
	}

	if err := s.store.DeleteObject(f); err != nil {
		return errors.New(errors.CodeInternal,
			fmt.Sprintf("failed to delete file object %s", fileID), err)
	}

	if err := s.repo.Delete(ctx, fileID); err != nil {
		return errors.New(errors.CodeInternal,
			fmt.Sprintf("failed to delete file metadata %s", fileID), err)
	}

	return nil
}
