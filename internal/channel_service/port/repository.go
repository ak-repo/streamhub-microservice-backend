package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/channel_service/domain"
)

type ChannelRepository interface {

	// -------------------------------------------------------------------------
	// Messaging
	// -------------------------------------------------------------------------
	SaveMessage(ctx context.Context, msg *domain.Message) error
	ListHistory(ctx context.Context, channelID string, limit, offset int) ([]*domain.Message, error)
	// -------------------------------------------------------------------------
	// Channel Management
	// -------------------------------------------------------------------------
	CreateChannel(ctx context.Context, ch *domain.Channel) error
	GetChannel(ctx context.Context, channelID string) (*domain.Channel, error)
	DeleteChannel(ctx context.Context, channelID string) error
	ListUserChannels(ctx context.Context, userID string) ([]*domain.Channel, error)

	// -------------------------------------------------------------------------
	// Membership
	// -------------------------------------------------------------------------
	AddMember(ctx context.Context, member *domain.ChannelMember) error
	RemoveMember(ctx context.Context, channelID, userID string) error
	ListChannelMembers(ctx context.Context, channelID string) ([]*domain.ChannelMember, error)
	IsUserMember(ctx context.Context, channelID, userID string) (bool, error)

	// -------------------------------------------------------------------------
	// Requests (Join Requests, Invites)
	// -------------------------------------------------------------------------
	CreateRequest(ctx context.Context, req *domain.Request) error
	UpdateRequestStatus(ctx context.Context, requestID, status string) error
	ListPendingRequests(ctx context.Context, userID, channelID string) ([]*domain.Request, error)

	// -------------------------------------------------------------------------
	// Admin / Governance
	// -------------------------------------------------------------------------
	AdminListChannels(ctx context.Context, limit, offset int) ([]*domain.Channel, error)
	FreezeChannel(ctx context.Context, channelID string, freeze bool) error
}
