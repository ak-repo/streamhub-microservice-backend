package domain

import "time"

type File struct {
	ID          string `db:"id" json:"id"`
	OwnerID     string `db:"owner_id" json:"ownerId"`
	ChannelID   string `db:"channel_id" json:"channelId,omitempty"` // nullable if file is personal
	Filename    string `db:"filename" json:"filename"`
	Size        int64  `db:"size" json:"size"`
	MimeType    string `db:"mime_type" json:"mimeType"`
	StoragePath string `db:"storage_path" json:"storagePath"`
	OwnerName   string
	ChannelName string
	IsPublic    bool      `db:"is_public" json:"isPublic"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
}

type StorageStats struct {
	TotalFilesCount   int64
	TotalStorageBytes int64
	PublicFilesCount  int64
	PrivateFilesCount int64
}


