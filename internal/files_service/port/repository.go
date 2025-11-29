package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
)

type FileRepository interface {
	// Save metadata
	Save(ctx context.Context, f *domain.File) error

	// List personal uploads
	GetByOwner(ctx context.Context, ownerID string) ([]*domain.File, error)

	// List files inside a specific channel
	GetByChannel(ctx context.Context, channelID string) ([]*domain.File, error)

	// List all files user can access: personal + channel + public
	GetUserAccessibleFiles(ctx context.Context, userID string) ([]*domain.File, error)

	// Get single file metadata
	GetByID(ctx context.Context, id string) (*domain.File, error)

	// Delete file
	Delete(ctx context.Context, id string) error

	IsChannelMember(ctx context.Context, channelID, userID string) (bool, error)
	IsChannelAdmin(ctx context.Context, channelID, userID string) (bool, error)
}
