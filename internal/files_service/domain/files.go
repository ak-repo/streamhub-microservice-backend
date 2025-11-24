package domain

import "time"

type File struct {
	ID          string    `db:"id" json:"id"`
	OwnerID     string    `db:"owner_id" json:"owner_id"`
	Filename    string    `db:"filename" json:"filename"`
	Size        int64     `db:"size" json:"size"`
	MimeType    string    `db:"mime_type" json:"mime_type"`
	StoragePath string    `db:"storage_path" json:"storage_path"`
	IsPublic    bool      `db:"is_public" json:"is_public"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type Access int

const (
	AccessPrivate Access = iota
	AccessPublic
)
