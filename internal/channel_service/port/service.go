package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/channel_service/domain"
)

// ChannelService defines the business logic interface for chat operations.
type ChannelService interface {

	// =========================================================================
	// Messaging (Real-time & Persistent)
	// =========================================================================
	PostMessage(ctx context.Context, senderID, channelID, content string) (*domain.Message, error)
	GetHistory(ctx context.Context, channelID string, limit, offset int) ([]*domain.Message, error)

	// SubscribeToChannel returns a read-only channel for real-time updates
	SubscribeToChannel(ctx context.Context, channelID string) (<-chan *domain.Message, error)

	// =========================================================================
	// Channel Management
	// =========================================================================
	CreateChannel(ctx context.Context, name, description, visibility, creatorID string) (*domain.Channel, error)
	GetChannel(ctx context.Context, channelID string) (*domain.Channel, error)
	ListUserChannels(ctx context.Context, userID string) ([]*domain.Channel, error)
	DeleteChannel(ctx context.Context, channelID, requesterID string) error
	SearchChannels(ctx context.Context, filter string, limit, offset int32) ([]*domain.Channel, error)

	// =========================================================================
	// Membership Management
	// =========================================================================
	AddMember(ctx context.Context, channelID, userID string) (*domain.ChannelMember, error)
	RemoveMember(ctx context.Context, channelID, userID, requesterID string) error
	ListMembers(ctx context.Context, channelID string) ([]*domain.ChannelMember, error)
	CheckMembership(ctx context.Context, channelID, userID string) (bool, error)

	// =========================================================================
	// Request Flow (Invites & Joins)
	// =========================================================================
	SendInvite(ctx context.Context, targetUserID, channelID, senderID string) error
	SendJoin(ctx context.Context, userID, channelID string) error
	RespondToRequest(ctx context.Context, requestID, userID, status string) error

	ListUserInvites(ctx context.Context, userID string) ([]*domain.Request, error)
	ListChannelJoins(ctx context.Context, channelID string) ([]*domain.Request, error)

	NotifyAdminUserJoined(ctx context.Context, channelID, newUserID string) error
	// =========================================================================
	// Administration (System Level)
	// =========================================================================
	AdminListChannels(ctx context.Context, limit, offset int32) ([]*domain.ChannelWithMembers, error)
	AdminFreezeChannel(ctx context.Context, channelID string, freeze bool, reason string) error
	AdminDeleteChannel(ctx context.Context, channelID string) error
}
