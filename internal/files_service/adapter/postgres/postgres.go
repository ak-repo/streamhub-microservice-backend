package postgres

import (
	"context"
	"time"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
	"github.com/ak-repo/stream-hub/internal/files_service/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fileRepo struct {
	pool *pgxpool.Pool
}

func NewFileRepository(pool *pgxpool.Pool) port.FileRepository {
	return &fileRepo{pool: pool}
}

// Save File Metadata
func (r *fileRepo) Save(ctx context.Context, f *domain.File) error {
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}

	_, err := r.pool.Exec(ctx, `
	INSERT INTO files (
		id, owner_id, channel_id, filename, size, mime_type,
		storage_path, is_public, created_at
	)
	VALUES ($1,$2,$3, $4,$5,$6,$7,$8,$9)
`,
		f.ID,
		f.OwnerID,
		f.ChannelID, // "" becomes NULL automatically
		f.Filename,
		f.Size,
		f.MimeType,
		f.StoragePath,
		f.IsPublic,
		f.CreatedAt,
	)

	return err
}

// List files uploaded BY the user (personal uploads)
func (r *fileRepo) GetByOwner(ctx context.Context, ownerID string) ([]*domain.File, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, owner_id, channel_id, filename, size, mime_type,
		       storage_path, is_public, created_at
		FROM files 
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*domain.File
	for rows.Next() {
		var f domain.File
		if err := rows.Scan(
			&f.ID,
			&f.OwnerID,
			&f.ChannelID,
			&f.Filename,
			&f.Size,
			&f.MimeType,
			&f.StoragePath,
			&f.IsPublic,
			&f.CreatedAt,
		); err != nil {
			return nil, err
		}
		files = append(files, &f)
	}
	return files, nil
}

// List files inside a channel
func (r *fileRepo) GetByChannel(ctx context.Context, channelID string) ([]*domain.File, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, owner_id, channel_id, filename, size, mime_type,
		       storage_path, is_public, created_at
		FROM files 
		WHERE channel_id = $1
		ORDER BY created_at DESC
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*domain.File
	for rows.Next() {
		var f domain.File
		if err := rows.Scan(
			&f.ID,
			&f.OwnerID,
			&f.ChannelID,
			&f.Filename,
			&f.Size,
			&f.MimeType,
			&f.StoragePath,
			&f.IsPublic,
			&f.CreatedAt,
		); err != nil {
			return nil, err
		}
		files = append(files, &f)
	}
	return files, nil
}

// Fetch single file
func (r *fileRepo) GetByID(ctx context.Context, id string) (*domain.File, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, owner_id, channel_id, filename, size, mime_type,
		       storage_path, is_public, created_at
		FROM files 
		WHERE id = $1
	`, id)

	var f domain.File
	if err := row.Scan(
		&f.ID,
		&f.OwnerID,
		&f.ChannelID,
		&f.Filename,
		&f.Size,
		&f.MimeType,
		&f.StoragePath,
		&f.IsPublic,
		&f.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &f, nil
}

// Delete file
func (r *fileRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM files WHERE id=$1`, id)
	return err
}

// List all files user can access (personal + channels they belong to)
func (r *fileRepo) GetUserAccessibleFiles(ctx context.Context, userID string) ([]*domain.File, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT f.id, f.owner_id, f.channel_id, f.filename, f.size, 
		       f.mime_type, f.storage_path, f.is_public, f.created_at
		FROM files f
		LEFT JOIN channel_members cm ON f.channel_id = cm.channel_id
		WHERE 
		    f.owner_id = $1         -- user uploaded the file
		    OR f.is_public = TRUE   -- public file
		    OR cm.user_id = $1      -- channel member
		ORDER BY f.created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*domain.File
	for rows.Next() {
		var f domain.File
		if err := rows.Scan(
			&f.ID,
			&f.OwnerID,
			&f.ChannelID,
			&f.Filename,
			&f.Size,
			&f.MimeType,
			&f.StoragePath,
			&f.IsPublic,
			&f.CreatedAt,
		); err != nil {
			return nil, err
		}
		files = append(files, &f)
	}

	return files, nil
}

func (r *fileRepo) IsChannelMember(ctx context.Context, channelID, userID string) (bool, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT 1 FROM channel_members
		WHERE channel_id=$1 AND user_id=$2
	`, channelID, userID)

	var one int
	err := row.Scan(&one)

	if err == pgx.ErrNoRows {
		return false, nil // not an error, simply not a member
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *fileRepo) IsChannelAdmin(ctx context.Context, channelID, userID string) (bool, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT 1 FROM channel_members
		WHERE channel_id=$1 AND user_id=$2
	`, channelID, userID)

	var one int
	err := row.Scan(&one)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
