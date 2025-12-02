package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/admin_service/domain"
)

// -------- AdminService Interface --------
type AdminService interface {
	// -------- User Governance --------
	ListUsers(ctx context.Context, filter string) ([]*domain.User, error)
	BanUser(ctx context.Context, userID, reason string) error
	UnbanUser(ctx context.Context, userID, reason string) error
	UpdateRole(ctx context.Context, userID, role string) error

	// -------- Channel Management --------
	ListChannels(ctx context.Context) ([]*domain.ChannelWithMembers, error)
	FreezeChannel(ctx context.Context, channelID, reason string) error
	UnfreezeChannel(ctx context.Context, channelID string) error
	DeleteChannel(ctx context.Context, channelID string) error

	// -------- File Management --------
	ListAllFiles(ctx context.Context, adminID string) ([]*domain.File, error)
	DeleteFile(ctx context.Context, adminID, fileID string) error
	BlockUserUpload(ctx context.Context, adminID, userID string, block bool) error
}

// ADMIN REPOSITORY PORT
// Low-level DB operations
type AdminRepository interface {

	// ------------- USER MANAGEMENT -------------

	ListUsers(ctx context.Context) ([]*domain.User, error)
	ListActiveUsers(ctx context.Context) ([]*domain.User, error)
	ListBannedUsers(ctx context.Context) ([]*domain.User, error)

	UpdateRole(ctx context.Context, userID, role string) error

	BanUser(ctx context.Context, userID string, reason string) error
	UnbanUser(ctx context.Context, userID string, reason string) error

	// upload blocking
	SetUserUploadBlocked(ctx context.Context, userID string, block bool) error

	// ------------- CHANNEL MANAGEMENT -------------

	ListChannels(ctx context.Context) ([]*domain.ChannelWithMembers, error)
	ListChannelMembers(ctx context.Context, channelID string) ([]*domain.ChannelMember, error)
	FreezeChannel(ctx context.Context, channelID string, reason string) error
	UnfreezeChannel(ctx context.Context, channelID string) error
	DeleteChannel(ctx context.Context, channelID string) error

	// ------------- FILE MANAGEMENT -------------

	ListAllFiles(ctx context.Context) ([]*domain.File, error)
	DeleteFile(ctx context.Context, fileID string) error
}
