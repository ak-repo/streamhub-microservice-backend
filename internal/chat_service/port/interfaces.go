package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/chat_service/domain"
)

// ChatRepository defines database operations for chat functionality.
type ChatRepository interface {
	// Message operations
	SaveMessage(ctx context.Context, msg *domain.Message) error
	ListHistory(ctx context.Context, channelID string, limit, offset int) ([]*domain.Message, error)

	// Channel operations - enables group management
	CreateChannel(ctx context.Context, ch *domain.Channel) error
	GetChannel(ctx context.Context, channelID string) (*domain.Channel, error)
	ListChannels(ctx context.Context, userID string) (map[string]*domain.ChannelWithMembers, error)

	// Membership operations - controls who can access channels
	AddMember(ctx context.Context, m *domain.ChannelMember) error
	RemoveMember(ctx context.Context, channelID, userID string) error
	ListChannelMembers(ctx context.Context, channelID string) ([]*domain.ChannelMember, error)
	IsUserMember(ctx context.Context, channelID, userID string) (bool, error)
}

// PubSub defines real-time message broadcasting capabilities.
type PubSub interface {
	Publish(ctx context.Context, channelID string, msg *domain.Message) error
	Subscribe(ctx context.Context, channelID string) (<-chan *domain.Message, error)
}

// ChatService defines the business logic interface for chat operations.
type ChatService interface {
	// Messaging
	PostMessage(ctx context.Context, senderID, channelID, content string) (*domain.Message, error)
	GetHistory(ctx context.Context, channelID string) ([]*domain.Message, error)
	SubscribeToChannel(ctx context.Context, channelID string) (<-chan *domain.Message, error)

	// Channel management
	CreateChannel(ctx context.Context, name, creatorID string) (*domain.Channel, error)
	GetChannel(ctx context.Context, channelID string) (*domain.Channel, error)
	ListChannels(ctx context.Context, userID string) (map[string]*domain.ChannelWithMembers, error)

	// Membership management
	AddMember(ctx context.Context, channelID, userID string) (*domain.ChannelMember, error)
	RemoveMember(ctx context.Context, channelID, userID string) error
	ListMembers(ctx context.Context, channelID string) ([]*domain.ChannelMember, error)
	CheckMembership(ctx context.Context, channelID, userID string) (bool, error)
}
