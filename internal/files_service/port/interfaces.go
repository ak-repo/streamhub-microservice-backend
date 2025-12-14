package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
)

// --- Persistence Interfaces (Database and Redis) ---
// FileRepository defines the contract for interacting with the primary data store (e.g., SQL DB).
type FileRepository interface {
	Save(ctx context.Context, file *domain.File) error

	Delete(ctx context.Context, fileID string) error

	// --- File Listing & Access ---
	GetByID(ctx context.Context, fileID string) (*domain.File, error)
	GetByOwner(ctx context.Context, ownerID string) ([]*domain.File, error)
	GetByChannel(ctx context.Context, channelID string) ([]*domain.File, error)
	GetUserAccessibleFiles(ctx context.Context, userID string) ([]*domain.File, error)

	// --- Channel/Access Checks ---
	IsChannelMember(ctx context.Context, channelID, userID string) (bool, error)
	IsChannelAdmin(ctx context.Context, channelID, userID string) (bool, error)


	



	// IsUserBlocked checks if a user is blocked from uploading.
	IsUserBlocked(ctx context.Context, userID string) (bool, error)

	// SetUserBlocked sets or unsets the upload block status for a user.
	SetUserBlocked(ctx context.Context, userID string, block bool) error

	// --- Admin Listings & Stats ---
	ListAllFiles(ctx context.Context, limit int32, offset int32) ([]*domain.File, error)

}

// FileStorage defines the contract for interacting with the object storage service (e.g., S3, GCS).
type FileStorage interface {
	GenerateUploadURL(file *domain.File) (string, error)
	GenerateDownloadURL(file *domain.File, expireSeconds int64) (string, error)
	DeleteObject(file *domain.File) error
}

// TempFileStore defines the contract for temporary, fast state storage (e.g., Redis).
type TempFileStore interface {
	SaveTemp(ctx context.Context, file *domain.File) error
	GetTemp(ctx context.Context, fileID string) (*domain.File, error)
	DeleteTemp(ctx context.Context, fileID string) error
}

// --- Service Interface (Business Logic) ---

// FileService defines the high-level business logic contract exposed by the service.
type FileService interface {
	// User Operations (FileService gRPC)
	GenerateUploadURL(ctx context.Context, ownerID, channelID, filename string, size int64, mime string, isPublic bool) (url, path, fileID string, err error)
	ConfirmUpload(ctx context.Context, fileID string) (*domain.File, error)
	GenerateDownloadURL(ctx context.Context, fileID, requesterID string, expirySeconds int64) (string, error)
	ListFiles(ctx context.Context, requesterID, channelID string) ([]*domain.File, error)
	DeleteFile(ctx context.Context, fileID, requesterID string) error

	

	// Admin Operations (AdminFileService gRPC)
	AdminListFiles(ctx context.Context, limit, offset int32) ([]*domain.File, error)
	AdminDeleteFile(ctx context.Context, fileID, adminID string, force bool) error
	// AdminSetStorageLimit(ctx context.Context, channelID string, maxBytes int64) (prevLimit int64, err error)
	AdminBlockUploads(ctx context.Context, targetID string, block bool) error
// 	AdminGetStats(ctx context.Context) (*domain.StorageStats, error)
}
