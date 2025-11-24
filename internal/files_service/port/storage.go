package port

import "github.com/ak-repo/stream-hub/internal/files_service/domain"

type FileStorage interface {
	// Generate DNS-presigned or S3 presigned PUT URL for this file
	GenerateUploadURL(file *domain.File) (string, error)

	// After upload, generate presigned GET for download
	GenerateDownloadURL(file *domain.File, expirySeconds int64) (string, error)

	// Remove object
	DeleteObject(file *domain.File) error

	// Direct upload -> keeping for future use
	Upload(file *domain.File, data []byte) error
	Download(file *domain.File) ([]byte, error)
}


