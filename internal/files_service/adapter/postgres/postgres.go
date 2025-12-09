package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
	"github.com/ak-repo/stream-hub/internal/files_service/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype" // Use pgtype for robust null handling
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL column list for simple file selects (excluding joins)
const fileColumns = `
    id, owner_id, channel_id, filename, size, mime_type,
    storage_path, is_public, created_at
`

// fileRepo implements the port.FileRepository interface.
type fileRepo struct {
	pool *pgxpool.Pool
}

// NewFileRepository is the constructor for the repository.
func NewFileRepository(pool *pgxpool.Pool) port.FileRepository {
	return &fileRepo{pool: pool}
}

// --- Repository Methods ---

// ListAllFiles retrieves all files metadata, including owner and channel names.
func (r *fileRepo) ListAllFiles(ctx context.Context, limit int32, offset int32) ([]*domain.File, error) {
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT 
			f.%s, 
			u.username,
			c.name
		FROM files f
		JOIN users u ON f.owner_id = u.id
		LEFT JOIN channels c ON f.channel_id = c.id -- Use LEFT JOIN since channel_id can be NULL
	`, fileColumns))

	if err != nil {
		return nil, fmt.Errorf("ListAllFiles: %w", err)
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
	return files, rows.Err()
}

// Save File Metadata
func (r *fileRepo) Save(ctx context.Context, f *domain.File) error {
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}

	_, err := r.pool.Exec(ctx, fmt.Sprintf(`
		INSERT INTO files (%s)
		VALUES ($1,$2,$3, $4,$5,$6,$7,$8,$9)
	`, fileColumns),
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

	return err
}

// GetByOwner lists files uploaded BY the user (personal uploads).
func (r *fileRepo) GetByOwner(ctx context.Context, ownerID string) ([]*domain.File, error) {
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT %s
		FROM files 
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`, fileColumns), ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMultipleFiles(rows)
}

func (r *fileRepo) GetByChannel(ctx context.Context, channelID string) ([]*domain.File, error) {
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT %s
		FROM files 
		WHERE channel_id = $1
		ORDER BY created_at DESC
	`, fileColumns), channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMultipleFiles(rows)
}

// GetByID fetches a single file by its ID.
func (r *fileRepo) GetByID(ctx context.Context, id string) (*domain.File, error) {
	row := r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s
		FROM files 
		WHERE id = $1
	`, fileColumns), id)

	f := new(domain.File)
	if err := scanFileSimple(row, f); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no file found") // Custom error for domain layer
		}
		return nil, err
	}

	return f, nil
}

// GetUserAccessibleFiles lists all files a user can access (personal + public + channel member).
func (r *fileRepo) GetUserAccessibleFiles(ctx context.Context, userID string) ([]*domain.File, error) {
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT f.%s
		FROM files f
		LEFT JOIN channel_members cm ON f.channel_id = cm.channel_id
		WHERE 
			f.owner_id = $1         -- user uploaded the file
			OR f.is_public = TRUE   -- public file
			OR cm.user_id = $1      -- channel member (via LEFT JOIN)
		ORDER BY f.created_at DESC
	`, fileColumns), userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMultipleFiles(rows)
}

func (r *fileRepo) Delete(ctx context.Context, id string) error {
	cmdTag, err := r.pool.Exec(ctx, `DELETE FROM files WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no file found") // Indicate that no row was deleted
	}
	return nil
}

// IsChannelMember checks if a user is a member of a channel.
func (r *fileRepo) IsChannelMember(ctx context.Context, channelID, userID string) (bool, error) {
	var exists bool
	log.Println("chan: ", channelID, "us:", userID)
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (
        SELECT 1 FROM channel_members
        WHERE channel_id=$1 AND user_id=$2
    )`,
		channelID,
		userID,
	).Scan(&exists)
	log.Println("ex: ", exists)

	if err != nil {
		return false, fmt.Errorf("IsChannelMember: %w", err)
	}

	return exists, nil
}

func (r *fileRepo) IsChannelAdmin(ctx context.Context, channelID, userID string) (bool, error) {
	var isAdmin bool
	err := r.pool.QueryRow(ctx, `
        SELECT created_by FROM channel_members
        WHERE channel_id=$1 AND user_id=$2
    `, channelID, userID).Scan(&isAdmin)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil // Not a member/admin
	}
	if err != nil {
		return false, fmt.Errorf("IsChannelAdmin: %w", err)
	}
	return isAdmin, nil
}

// Helper to handle repetitive row iteration and scanning.
func (r *fileRepo) scanMultipleFiles(rows pgx.Rows) ([]*domain.File, error) {
	var files []*domain.File
	for rows.Next() {
		f := new(domain.File)
		if err := scanFileSimple(rows, f); err != nil {
			return nil, fmt.Errorf("scanFileSimple: %w", err)
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

// --- Scanning Functions (Reduced Duplication) ---

// scanFileJoined scans rows that include owner username and channel name (used by ListAllFiles).
func scanFileJoined(row pgx.Row, f *domain.File) error {
	// Use pgtype.Text for nullable columns like ChannelID and ChannelName
	var channelID pgtype.Text
	var channelName pgtype.Text

	err := row.Scan(
		&f.ID,
		&f.OwnerID,
		&channelID, // Scan into pgtype.Text
		&f.Filename,
		&f.Size,
		&f.MimeType,
		&f.StoragePath,
		&f.IsPublic,
		&f.CreatedAt,
		&f.OwnerName,
		&channelName, // Scan into pgtype.Text
	)

	if err != nil {
		return err
	}

	// Convert pgtype.Text back to string, handling NULLs
	f.ChannelID = channelID.String
	f.ChannelName = channelName.String

	return nil
}

// scanFileSimple scans rows that ONLY select columns from the 'files' table.
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
	// Note: f.OwnerName and f.ChannelName are intentionally left as zero values ("")

	return nil
}

func (r *fileRepo) GetStorageUsage(ctx context.Context, channelID string) (used int64, limit int64, err error) {
	return int64(7), int64(0), nil
}

func (r *fileRepo) IsUserBlocked(ctx context.Context, userID string) (bool, error) {
	return false, nil
}

func (r *fileRepo) SetUserBlocked(ctx context.Context, userID string, block bool) error {
	return nil
}

func (r *fileRepo) SetStorageLimit(ctx context.Context, ownerID string, limit int64) error {
	return nil
}

func (r *fileRepo) GetGlobalStats(ctx context.Context) (*domain.StorageStats, error) {

	return &domain.StorageStats{}, nil
}
