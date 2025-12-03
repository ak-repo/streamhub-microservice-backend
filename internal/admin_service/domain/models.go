package domain

import "time"

// =======================
// USER ENTITY
// =======================
type User struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	CreatedAt     time.Time `json:"created_at"`
	EmailVerified bool      `json:"email_verified"`
	IsBanned      bool      `json:"is_banned"`
}

// =======================
// CHANNEL ENTITY
// =======================
type Channel struct {
	ID          string    `json:"id"`
	CreatedBy   string    `json:"created_by"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsFrozen    bool      `json:"is_frozen"`
	CreatedAt   time.Time `json:"created_at"`
	OwnerName   string    `json:"owner_name"`
}

type ChannelMember struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joined_at"`
}

type ChannelWithMembers struct {
	Channel *Channel         `json:"channel"`
	Members []*ChannelMember `json:"members"`
}

// =======================
// FILE ENTITY
// =======================
type File struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	ChannelID   string    `json:"channel_id"`
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	MimeType    string    `json:"mime_type"`
	StoragePath string    `json:"storage_path"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	OwnerName   string    `json:"owner_name"`
	ChannelName string    `json:"channel_name"`
}
