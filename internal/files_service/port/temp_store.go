package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
)

type TempFileStore interface {
	SaveTemp(ctx context.Context, f *domain.File) error
	GetTemp(ctx context.Context, fileID string) (*domain.File, error)
	DeleteTemp(ctx context.Context, fileID string) error
}
