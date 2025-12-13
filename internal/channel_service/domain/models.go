package domain

import (
	"time"
)

// Channel represents a chat group where users can exchange messages.
type Channel struct {
	ID        string    `json:"id"`          // Channel UUID
	Name      string    `json:"name"`        // Channel name
	CreatedBy string    `json:"created_by"`  // Owner user ID

	Description string `json:"description"`

	Visibility string `json:"visibility"` // public / private / hidden
	IsFrozen   bool   `json:"is_frozen"`
	IsArchived bool   `json:"is_archived"`

	// PLAN & STORAGE (IMPORTANT)
	ActivePlanID   string `json:"active_plan_id"`   // FK → channel_storage_plans.id
	StorageLimitMB int64  `json:"storage_limit_mb"` // Total allowed storage
	StorageUsedMB  int64  `json:"storage_used_mb"`  // Used storage

	// METADATA
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// READ-ONLY / JOINED FIELDS (NOT STORED IN channels TABLE)
	OwnerName    string `json:"owner_name,omitempty"`
	MembersCount int32  `json:"members_count,omitempty"`
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
