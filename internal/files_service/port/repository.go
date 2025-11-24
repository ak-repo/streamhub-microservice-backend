package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
)

type FileRepository interface {
	Save(ctx context.Context, f *domain.File) error
	GetByOwner(ctx context.Context, ownerID string) ([]*domain.File, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.File, error)
}
