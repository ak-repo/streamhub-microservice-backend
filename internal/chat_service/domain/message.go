package domain

import "time"

// Message represents a persisted chat message in a group/channel.
type Message struct {
	ID        string    // Unique message identifier (UUID)
	ChannelID string    // References which group/channel this belongs to
	SenderID  string    // User who sent the message
	Content   string    // The actual message text
	CreatedAt time.Time // When the message was created
}
