package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
)

// AuthService describes business-level auth operations.
type AuthService interface {
	// -------------------------------------------------------------------------
	// Authentication & Identity
	// -------------------------------------------------------------------------
	Register(ctx context.Context, email, username, password string) error
	Login(ctx context.Context, email, password string) (*domain.User, error)

	// Magic Link Flow
	SendMagicLink(email string) (string, string, error)
	VerifyMagicLink(ctx context.Context, token, email string) error

	// -------------------------------------------------------------------------
	// User Management (Self)
	// -------------------------------------------------------------------------
	FindUser(ctx context.Context, identifier, method string) (*domain.User, error)

	// Profile
	UpdateProfile(ctx context.Context, userID, username string) (*domain.User, error)
	UploadAvatar(ctx context.Context, userID string, fileBytes []byte, filename string, contentType string) (string, error)

	// Security
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	PasswordReset(ctx context.Context, email string) error
	VerifyPasswordReset(ctx context.Context, otp, newPassword, email string) error

	// -------------------------------------------------------------------------
	// Discovery
	// -------------------------------------------------------------------------
	// SearchUsers finds users by partial name/email (Public facing)
	SearchUsers(ctx context.Context, query string) ([]*domain.User, error)

	// -------------------------------------------------------------------------
	// Administration & Governance
	// -------------------------------------------------------------------------
	// ListUsers returns users based on filter (active/banned/all)
	ListUsers(ctx context.Context, filter string) ([]*domain.User, error)

	UpdateRole(ctx context.Context, userID, role string) error
	BanUser(ctx context.Context, userID, reason string) error
	UnbanUser(ctx context.Context, userID, reason string) error
	BlockUserUpload(ctx context.Context, adminID, userID string, block bool) error
}
