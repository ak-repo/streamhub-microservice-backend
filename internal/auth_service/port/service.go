package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
)

// AuthService describes business-level auth operations.
type AuthService interface {
	Register(ctx context.Context, email, username, password string) error
	Login(ctx context.Context, email, password string) (*domain.User, error)
	SendMagicLink(email string) (string, string, error)
	VerifyMagicLink(ctx context.Context, token, email string) error
	FindUser(ctx context.Context, identifier, method string) (*domain.User, error)
	PasswordReset(ctx context.Context, email string) error
	VerifyPasswordReset(ctx context.Context, otp, password, email string) error
	UpdateProfile(ctx context.Context, userID, username, email string) (*domain.User, error)
	ChangePassword(ctx context.Context, userID, password, newPassword string) error
	FindAllUsers(ctx context.Context, query string) ([]*domain.User, error)
}
