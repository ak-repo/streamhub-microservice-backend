package postgres

import (
	"context"
	"time"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
	"github.com/ak-repo/stream-hub/internal/files_service/port"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fileRepo struct {
	pool *pgxpool.Pool
}

func NewFileRepository(pool *pgxpool.Pool) port.FileRepository {
	return &fileRepo{pool: pool}
}

func (r *fileRepo) Save(ctx context.Context, f *domain.File) error {
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO files (id, owner_id, filename, size, mime_type, storage_path, is_public, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		f.ID, f.OwnerID, f.Filename, f.Size, f.MimeType, f.StoragePath, f.IsPublic, f.CreatedAt)
	return err
}

func (r *fileRepo) GetByOwner(ctx context.Context, ownerID string) ([]*domain.File, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, owner_id, filename, size, mime_type, storage_path, is_public, created_at
		FROM files WHERE owner_id=$1`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []*domain.File
	for rows.Next() {
		var f domain.File
		if err := rows.Scan(&f.ID, &f.OwnerID, &f.Filename, &f.Size, &f.MimeType, &f.StoragePath, &f.IsPublic, &f.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, &f)
	}
	return files, nil
}

func (r *fileRepo) GetByID(ctx context.Context, id string) (*domain.File, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, owner_id, filename, size, mime_type, storage_path, is_public, created_at
		FROM files WHERE id=$1`, id)
	var f domain.File
	if err := row.Scan(&f.ID, &f.OwnerID, &f.Filename, &f.Size, &f.MimeType, &f.StoragePath, &f.IsPublic, &f.CreatedAt); err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *fileRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM files WHERE id=$1`, id)
	return err
}
