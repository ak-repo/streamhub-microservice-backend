package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
)

type FileService interface {
	// Step 1: Generate a presigned upload URL
	GenerateUploadURL(
		ctx context.Context,
		ownerID string,
		channelID string,
		filename string,
		size int64,
		mime string,
		isPublic bool,
	) (uploadURL string, storagePath string, fileID string, err error)

	// Step 2: Confirm upload
	ConfirmUpload(ctx context.Context, fileID string) (*domain.File, error)

	// Step 3: Generate a download URL (presigned)
	GenerateDownloadURL(
		ctx context.Context,
		fileID string,
		requesterID string,
		expirySeconds int64,
	) (downloadURL string, err error)

	// List  files for channel
	ListFiles(ctx context.Context, requesterID, channelID string) ([]*domain.File, error)

	// Delete file (metadata + S3 object)
	// Allowed for: file owner OR channel admin
	DeleteFile(ctx context.Context, fileID string, requesterID string) error
}
