package storage

import (
	"bytes"
	"context"
	"time"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
	client     *minio.Client
	bucket     string
	useSSL     bool
	presignTTL time.Duration
}

func NewS3Storage(endpoint, accessKey, secretKey, bucket string, useSSL bool, presignTTL time.Duration) (*S3Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	// ensure bucket exists
	ctx := context.Background()
	ok, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !ok {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}
	return &S3Storage{client: client, bucket: bucket, useSSL: useSSL, presignTTL: presignTTL}, nil
}

func (s *S3Storage) GenerateUploadURL(file *domain.File) (string, error) {
	// Presigned PUT
	ctx := context.Background()
	// We can add headers or req params if needed
	u, err := s.client.PresignedPutObject(ctx, s.bucket, file.StoragePath, s.presignTTL)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *S3Storage) GenerateDownloadURL(file *domain.File, expirySeconds int64) (string, error) {
	ctx := context.Background()
	u, err := s.client.PresignedGetObject(ctx, s.bucket, file.StoragePath, time.Duration(expirySeconds)*time.Second, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *S3Storage) DeleteObject(file *domain.File) error {
	ctx := context.Background()
	return s.client.RemoveObject(ctx, s.bucket, file.StoragePath, minio.RemoveObjectOptions{})
}



func (s *S3Storage) Upload(file *domain.File, data []byte) error {
	_, err := s.client.PutObject(context.Background(), s.bucket, file.StoragePath, bytes.NewReader(data), file.Size, minio.PutObjectOptions{
		ContentType: file.MimeType,
	})
	return err
}

func (s *S3Storage) Download(file *domain.File) ([]byte, error) {
	obj, err := s.client.GetObject(context.Background(), s.bucket, file.StoragePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(obj)
	return buf.Bytes(), err
}
