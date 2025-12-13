package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
)

// UserRepository defines storage operations for users.
type UserRepository interface {
	// -------------------------------------------------------------------------
	// Core Identity Management (CRUD)
	// -------------------------------------------------------------------------
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error

	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)

	// -------------------------------------------------------------------------
	// Profile & Security Specifics
	// -------------------------------------------------------------------------
	UpdatePassword(ctx context.Context, email string, passwordHash string) error
	UpdateAvatar(ctx context.Context, userID string, avatarURL string) error

	// -------------------------------------------------------------------------
	// Discovery & Listing
	// -------------------------------------------------------------------------
	// SearchUsers filters users by username or email partial match.
	SearchUsers(ctx context.Context, query string) ([]*domain.User, error)

	// ListUsers returns all users (typically ordered by creation date).
	ListUsers(ctx context.Context, filer string, limit, offset int32) ([]*domain.User, int32, error)

	// -------------------------------------------------------------------------
	// Governance (Admin Actions)
	// -------------------------------------------------------------------------
	UpdateRole(ctx context.Context, userID string, role string) error
	BanUser(ctx context.Context, userID string, reason string) error
	UnbanUser(ctx context.Context, userID string, reason string) error
	SetUserUploadBlocked(ctx context.Context, userID string, blocked bool) error
	DeleteUser(ctx context.Context, id string) error
}
