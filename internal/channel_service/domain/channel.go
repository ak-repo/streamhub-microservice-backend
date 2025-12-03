package domain

import (
	"time"
)

// Channel represents a chat group where users can exchange messages.
type Channel struct {
	ID          string    // Unique channel identifier (UUID)
	Name        string    // Human-readable channel name (e.g., "general", "random")
	CreatedBy   string    // User ID of the channel creator
	CreatedAt   time.Time // When the channel was created
	Description string
	Visibility  string // public / private
	IsFrozen    bool
	OwnerName   string `json:"owner_name"`
}

// ChannelMember represents a user's membership in a channel.
type ChannelMember struct {
	ID        string // Unique membership record ID
	ChannelID string // Which channel this membership is for
	UserID    string // Which user is a member
	Username  string
	JoinedAt  time.Time // When the user joined the channel
}

type ChannelWithMembers struct {
	Channel *Channel
	Members []*ChannelMember
}
