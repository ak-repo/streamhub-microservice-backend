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
		filename string,
		size int64,
		mime string,
		isPublic bool,
	) (uploadURL string, storagePath string, fileID string, err error)

	// Step 2: Confirm upload
	ConfirmUpload(ctx context.Context, fileID string) (*domain.File, error)

	// Step 3: Generate a download URL (presigned)
	GenerateDownloadURL(ctx context.Context, fileID string, exp int64) (downloadURL string, err error)

	// List user files
	ListFiles(ctx context.Context, ownerID string) ([]*domain.File, error)

	// Delete file (metadata + S3 object)
	DeleteFile(ctx context.Context, fileID string) error
}
