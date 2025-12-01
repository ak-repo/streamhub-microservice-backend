package domain

import "time"

// FileAttachment holds metadata for a file shared in a chat message.
type FileAttachment struct {
	ID       string // Unique  (UUID)
	FileID   string // Unique ID of the uploaded file
	FileName string
	URL      string // Public or signed URL to access the file
	MimeType string // e.g., "image/png", "application/pdf"
	Size     int64  // Size in bytes
}

// Message represents a persisted chat message in a group/channel.
type Message struct {
	ID         string          // Unique message identifier (UUID)
	ChannelID  string          // References which group/channel this belongs to
	SenderID   string          // User who sent the message
	Username   string          // Display name of the sender
	Content    string          // The actual message text
	Attachment *FileAttachment // Optional file attachment metadata
	CreatedAt  time.Time       // When the message was created
}
