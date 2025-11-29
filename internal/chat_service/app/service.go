package app

import (
	"context"
	"fmt"
	"time"

	"github.com/ak-repo/stream-hub/internal/chat_service/domain"
	"github.com/ak-repo/stream-hub/internal/chat_service/port"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type chatService struct {
	repo   port.ChatRepository
	pubsub port.PubSub
}

func NewChatService(repo port.ChatRepository, pubsub port.PubSub) port.ChatService {
	return &chatService{repo: repo, pubsub: pubsub}
}

// PostMessage handles the message posting flow:
// 1. Persist to database (durable storage)
// 2. Broadcast via Redis (real-time delivery)
func (s *chatService) PostMessage(ctx context.Context, senderID, channelID, content string) (*domain.Message, error) {
	// Optional: Verify sender is a channel member
	isMember, err := s.repo.IsUserMember(ctx, channelID, senderID)
	if err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to check membership", err)
	}
	if !isMember {
		return nil, errors.New(errors.CodeUnauthorized, fmt.Sprintf("user %s is not a member of channel %s", senderID, channelID), nil)
	}

	msg := &domain.Message{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		SenderID:  senderID,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}

	// Save to database first (critical path)
	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to save message", err)
	}

	// Broadcast to subscribers (non-critical - message is already saved)
	if err := s.pubsub.Publish(ctx, channelID, msg); err != nil {
		// Log but don't fail the request - message is already persisted
		logger.Log.Error("failed to publish message to Redis", zap.Error(err))
	}
	return msg, nil
}

func (s *chatService) GetHistory(ctx context.Context, channelID string) ([]*domain.Message, error) {
	messages, err := s.repo.ListHistory(ctx, channelID, 50, 0)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "message on this channels not found", err)
	}
	return messages, nil
}

func (s *chatService) SubscribeToChannel(ctx context.Context, channelID string) (<-chan *domain.Message, error) {
	msgChan, err := s.pubsub.Subscribe(ctx, channelID)
	if err != nil {
		return nil, errors.New(errors.CodeInternal, fmt.Sprintf("Failed to subscribe to channel %s", channelID), err)
	}
	return msgChan, nil
}

// CreateChannel creates a new group chat channel and automatically
// adds the creator as the first member.
func (s *chatService) CreateChannel(ctx context.Context, name, creatorID string) (*domain.Channel, error) {
	ch := &domain.Channel{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedBy: creatorID,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateChannel(ctx, ch); err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to create channel", err)
	}

	// Auto-add creator as first member
	_, err := s.AddMember(ctx, ch.ID, creatorID)
	if err != nil {
		// Channel exists but creator not added - log and continue
		fmt.Printf("Warning: failed to add creator as member: %v\n", err)
	}

	return ch, nil
}

// Get All channels of a user
func (s *chatService) ListChannels(ctx context.Context, userID string) (map[string]*domain.ChannelWithMembers, error) {

	channels, err := s.repo.ListChannels(ctx, userID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "channels not found", err)
	}
	return channels, nil

}

func (s *chatService) AddMember(ctx context.Context, channelID, userID string) (*domain.ChannelMember, error) {
	// Verify channel exists
	_, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "channel not found", err)
	}

	m := &domain.ChannelMember{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    userID,
		JoinedAt:  time.Now().UTC(),
	}

	if err := s.repo.AddMember(ctx, m); err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to add member", err)
	}

	return m, nil
}

func (s *chatService) GetChannel(ctx context.Context, channelID string) (*domain.Channel, error) {
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "channel not found", err)
	}
	return channel, nil
}

func (s *chatService) RemoveMember(ctx context.Context, channelID, userID string) error {
	if err := s.repo.RemoveMember(ctx, channelID, userID); err != nil {
		return errors.New(errors.CodeInternal, "member removing failed", err)
	}
	return nil
}

func (s *chatService) ListMembers(ctx context.Context, channelID string) ([]*domain.ChannelMember, error) {
	return s.repo.ListChannelMembers(ctx, channelID)
}

func (s *chatService) CheckMembership(ctx context.Context, channelID, userID string) (bool, error) {
	return s.repo.IsUserMember(ctx, channelID, userID)
}

func (s *chatService) DeleteChannel(ctx context.Context, channelID, userID string) error {

	return s.repo.DeleteChannel(ctx, channelID, userID)

}
