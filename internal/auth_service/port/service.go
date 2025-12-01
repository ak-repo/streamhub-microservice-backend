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
}
