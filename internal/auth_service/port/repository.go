package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
)

// UserRepository defines storage operations for users.
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
}
