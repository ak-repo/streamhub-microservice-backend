package domain

import (
	"time"
)

// Channel represents a chat group where users can exchange messages.
type Channel struct {
	ID           string    // Unique channel identifier (UUID)
	Name         string    // Human-readable channel name (e.g., "general", "random")
	OwnerID      string    // User ID of the channel creator
	CreatedAt    time.Time // When the channel was created
	Description  string
	Visibility   string // public / private
	IsFrozen     bool
	OwnerName    string `json:"owner_name"`
	MembersCount int32
}

// ChannelMember represents a user's membership in a channel.
type ChannelMember struct {
	ID        string // Unique membership record ID
	ChannelID string // Which channel this membership is for
	UserID    string // Which user is a member
	Username  string
	Role      string
	JoinedAt  time.Time // When the user joined the channel
}

type ChannelWithMembers struct {
	Channel *Channel
	Members []*ChannelMember
}

// Message represents a persisted chat message in a group/channel.
type Message struct {
	ID        string    // Unique message identifier (UUID)
	ChannelID string    // References which group/channel this belongs to
	SenderID  string    // User who sent the message
	Username  string    // Display name of the sender
	Content   string    // The actual message text
	CreatedAt time.Time // When the message was created
}

type Request struct {
	ID        string
	UserID    string
	ChannelID string
	Type      string
	Status    string
	CreatedAt time.Time
}
