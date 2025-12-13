package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
	"github.com/ak-repo/stream-hub/internal/files_service/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NOTE:
//   - fileColumnsSelect is used in SELECT queries and includes the table alias `f.`
//     to avoid ambiguous column references when joining other tables.
//   - fileColumnsInsert is used in INSERT statements (no alias).
const fileColumnsSelect = `
    f.id, f.owner_id, f.channel_id, f.filename, f.size, f.mime_type,
    f.storage_path, f.is_public, f.created_at
`

const fileColumnsInsert = `
    id, owner_id, channel_id, filename, size, mime_type,
    storage_path, is_public, created_at
`

type fileRepo struct {
	pool *pgxpool.Pool
}

func NewFileRepository(pool *pgxpool.Pool) port.FileRepository {
	return &fileRepo{pool: pool}
}

func (r *fileRepo) ListAllFiles(ctx context.Context, limit int32, offset int32) ([]*domain.File, error) {
	query := fmt.Sprintf(`
		SELECT 
			%s,
			u.username,
			c.name
		FROM files f
		JOIN users u ON f.owner_id = u.id
		LEFT JOIN channels c ON f.channel_id = c.id
		ORDER BY f.created_at DESC
		LIMIT $1 OFFSET $2
	`, fileColumnsSelect)

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ListAllFiles: query: %w", err)
	}
	defer rows.Close()

	var files []*domain.File
	for rows.Next() {
		f := new(domain.File)
		if err := scanFileJoined(rows, f); err != nil {
			return nil, fmt.Errorf("ListAllFiles: scanFileJoined: %w", err)
		}
		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListAllFiles: rows: %w", err)
	}

	return files, nil
}

// Save inserts a new file row. If CreatedAt is zero, set to now UTC.
func (r *fileRepo) Save(ctx context.Context, f *domain.File) error {
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}

	query := fmt.Sprintf(`
		INSERT INTO files (%s)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, fileColumnsInsert)

	_, err := r.pool.Exec(ctx, query,
		f.ID,
		f.OwnerID,
		f.ChannelID,
		f.Filename,
		f.Size,
		f.MimeType,
		f.StoragePath,
		f.IsPublic,
		f.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("Save: exec: %w", err)
	}
	return nil
}

// GetByOwner lists files uploaded BY the user (personal uploads).
func (r *fileRepo) GetByOwner(ctx context.Context, ownerID string) ([]*domain.File, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM files
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`, fileColumnsInsert)

	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("GetByOwner: query: %w", err)
	}
	defer rows.Close()

	return r.scanMultipleFiles(rows)
}

func (r *fileRepo) GetByChannel(ctx context.Context, channelID string) ([]*domain.File, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM files
		WHERE channel_id = $1
		ORDER BY created_at DESC
	`, fileColumnsInsert)

	rows, err := r.pool.Query(ctx, query, channelID)
	if err != nil {
		return nil, fmt.Errorf("GetByChannel: query: %w", err)
	}
	defer rows.Close()

	return r.scanMultipleFiles(rows)
}

// GetByID fetches a single file by its ID.
func (r *fileRepo) GetByID(ctx context.Context, id string) (*domain.File, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM files
		WHERE id = $1
	`, fileColumnsInsert)

	row := r.pool.QueryRow(ctx, query, id)

	f := new(domain.File)
	if err := scanFileSimple(row, f); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no file found")
		}
		return nil, fmt.Errorf("GetByID: scan: %w", err)
	}

	return f, nil
}

// GetUserAccessibleFiles lists all files a user can access (personal + public + channel member).
func (r *fileRepo) GetUserAccessibleFiles(ctx context.Context, userID string) ([]*domain.File, error) {
	query := fmt.Sprintf(`
		SELECT f.%s
		FROM files f
		LEFT JOIN channel_members cm ON f.channel_id = cm.channel_id
		WHERE 
			f.owner_id = $1
			OR f.is_public = TRUE
			OR cm.user_id = $1
		ORDER BY f.created_at DESC
	`, fileColumnsSelect)

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserAccessibleFiles: query: %w", err)
	}
	defer rows.Close()

	return r.scanMultipleFiles(rows)
}

func (r *fileRepo) Delete(ctx context.Context, id string) error {
	cmdTag, err := r.pool.Exec(ctx, `DELETE FROM files WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("Delete: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no file found")
	}
	return nil
}

// IsChannelMember checks if a user is a member of a channel.
func (r *fileRepo) IsChannelMember(ctx context.Context, channelID, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM channel_members
            WHERE channel_id=$1 AND user_id=$2
        )
    `, channelID, userID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("IsChannelMember: %w", err)
	}
	return exists, nil
}

// IsChannelAdmin checks if a user has the admin role in the channel.
// Assumes channel_members.role contains values like 'admin' / 'member'.
func (r *fileRepo) IsChannelAdmin(ctx context.Context, channelID, userID string) (bool, error) {
	var role pgtype.Text
	err := r.pool.QueryRow(ctx, `
        SELECT role FROM channel_members
        WHERE channel_id=$1 AND user_id=$2
    `, channelID, userID).Scan(&role)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("IsChannelAdmin: %w", err)
	}

	// role can be NULL; handle it safely

	return role.String == "admin", nil
}

// Helper to scan multiple files (rows -> []*domain.File)
func (r *fileRepo) scanMultipleFiles(rows pgx.Rows) ([]*domain.File, error) {
	var files []*domain.File
	for rows.Next() {
		f := new(domain.File)
		if err := scanFileSimple(rows, f); err != nil {
			return nil, fmt.Errorf("scanMultipleFiles: %w", err)
		}
		files = append(files, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scanMultipleFiles: rows: %w", err)
	}
	return files, nil
}

func scanFileJoined(row pgx.Row, f *domain.File) error {
	var channelID pgtype.Text
	var channelName pgtype.Text

	var ownerName pgtype.Text

	err := row.Scan(
		&f.ID,
		&f.OwnerID,
		&channelID,
		&f.Filename,
		&f.Size,
		&f.MimeType,
		&f.StoragePath,
		&f.IsPublic,
		&f.CreatedAt,
		&ownerName,
		&channelName,
	)
	if err != nil {
		return err
	}
	f.ChannelID = channelID.String
	f.ChannelName = channelName.String
	f.OwnerName = ownerName.String

	return nil
}

func scanFileSimple(row pgx.Row, f *domain.File) error {
	var channelID pgtype.Text

	err := row.Scan(
		&f.ID,
		&f.OwnerID,
		&channelID,
		&f.Filename,
		&f.Size,
		&f.MimeType,
		&f.StoragePath,
		&f.IsPublic,
		&f.CreatedAt,
	)
	if err != nil {
		return err
	}
	f.ChannelID = channelID.String
	// OwnerName and ChannelName remain empty unless joined query used

	return nil
}

func (r *fileRepo) GetStorageUsage(
	ctx context.Context,
	channelID string,
) (usedMB int64, limitMB int64, err error) {

	const query = `
        SELECT storage_used_mb, storage_limit_mb
        FROM channels
        WHERE id = $1
    `

	row := r.pool.QueryRow(ctx, query, channelID)

	if err = row.Scan(&usedMB, &limitMB); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, fmt.Errorf("no channels found")
		}
		return 0, 0, err
	}

	return usedMB, limitMB, nil
}



func (r *fileRepo) IsUserBlocked(ctx context.Context, userID string) (bool, error) {
	// TODO: implement block check
	return false, nil
}

func (r *fileRepo) SetUserBlocked(ctx context.Context, userID string, block bool) error {
	// TODO: implement
	return nil
}

func (r *fileRepo) SetStorageLimit(ctx context.Context, ownerID string, limit int64) error {
	// TODO: implement
	return nil
}

func (r *fileRepo) GetGlobalStats(ctx context.Context) (*domain.StorageStats, error) {
	// TODO: implement real stats
	return &domain.StorageStats{}, nil
}
