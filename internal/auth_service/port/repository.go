package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
)

// UserRepository defines storage operations for users.
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
	UpdatePassword(ctx context.Context, email, hash string) error
	FindAll(ctx context.Context, query string) ([]*domain.User, error)
}
