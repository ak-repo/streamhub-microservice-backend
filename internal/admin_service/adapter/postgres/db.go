package postgres

import (
	"context"
	"fmt"

	"github.com/ak-repo/stream-hub/internal/admin_service/domain"
	"github.com/ak-repo/stream-hub/internal/admin_service/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type adminRepo struct {
	pool *pgxpool.Pool
}

func NewAdminRepo(pool *pgxpool.Pool) port.AdminRepository {
	return &adminRepo{pool: pool}
}

//
// ---------------------------------------------------------
// SCANNERS (Correct Ordering)
// ---------------------------------------------------------
//

func scanUser(row pgx.Row, u *domain.User) error {
	return row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Role,
		&u.CreatedAt,
		&u.EmailVerified,
		&u.IsBanned,
	)
}

func scanChannel(row pgx.Row, c *domain.Channel) error {
	return row.Scan(
		&c.ID,
		&c.CreatedBy,
		&c.Name,
		&c.Description,
		&c.IsFrozen,
		&c.CreatedAt,
		&c.OwnerName,
	)
}

func scanMember(row pgx.Row, m *domain.ChannelMember) error {
	return row.Scan(
		&m.ID,
		&m.UserID,
		&m.Username,
		&m.JoinedAt,
	)
}

func scanFile(row pgx.Row, f *domain.File) error {
	return row.Scan(
		&f.ID,
		&f.OwnerID,
		&f.ChannelID,
		&f.Filename,
		&f.Size,
		&f.MimeType,
		&f.StoragePath,
		&f.IsPublic,
		&f.CreatedAt,
		&f.OwnerName,
		&f.ChannelName,
	)
}

//
// ---------------------------------------------------------
// USER MANAGEMENT
// ---------------------------------------------------------
//

const userColumns = `
	id, username, email, role, created_at, email_verified, is_banned
`

func (r *adminRepo) fetchUsers(ctx context.Context, query string, args ...any) ([]*domain.User, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetchUsers: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := new(domain.User)
		if err := scanUser(rows, u); err != nil {
			return nil, fmt.Errorf("scanUser: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *adminRepo) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return r.fetchUsers(ctx, `
		SELECT `+userColumns+`
		FROM users
		ORDER BY created_at DESC
	`)
}

func (r *adminRepo) ListActiveUsers(ctx context.Context) ([]*domain.User, error) {
	return r.fetchUsers(ctx, `
		SELECT `+userColumns+`
		FROM users
		WHERE is_banned = false
		ORDER BY created_at DESC
	`)
}

func (r *adminRepo) ListBannedUsers(ctx context.Context) ([]*domain.User, error) {
	return r.fetchUsers(ctx, `
		SELECT `+userColumns+`
		FROM users
		WHERE is_banned = true
		ORDER BY created_at DESC
	`)
}

func (r *adminRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	u := new(domain.User)
	row := r.pool.QueryRow(ctx, `
		SELECT `+userColumns+`
		FROM users
		WHERE id = $1
	`, id)

	if err := scanUser(row, u); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user %s not found", id)
		}
		return nil, fmt.Errorf("GetUserByID: %w", err)
	}
	return u, nil
}

func (r *adminRepo) BanUser(ctx context.Context, id, reason string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET is_banned = true WHERE id = $1
	`, id)
	return err
}

func (r *adminRepo) UnbanUser(ctx context.Context, id, reason string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET is_banned = false WHERE id = $1
	`, id)
	return err
}

func (r *adminRepo) UpdateRole(ctx context.Context, id, role string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET role = $1 WHERE id = $2
	`, role, id)
	return err
}

func (r *adminRepo) SetUserUploadBlocked(ctx context.Context, id string, blocked bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET upload_blocked = $1 WHERE id = $2
	`, blocked, id)
	return err
}

//
// ---------------------------------------------------------
// CHANNEL MANAGEMENT
// ---------------------------------------------------------
//

func (r *adminRepo) ListChannels(ctx context.Context) ([]*domain.ChannelWithMembers, error) {

	rows, err := r.pool.Query(ctx, `
		SELECT 
			c.id,
			c.created_by,
			c.name,
			c.description,
			c.is_frozen,
			c.created_at,
			u.username AS owner_name
		FROM channels c
		JOIN users u ON u.id = c.created_by
		ORDER BY c.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer rows.Close()

	var channels []*domain.Channel
	var channelIDs []string

	for rows.Next() {
		c := new(domain.Channel)
		if err := scanChannel(rows, c); err != nil {
			return nil, fmt.Errorf("scanChannel: %w", err)
		}
		channels = append(channels, c)
		channelIDs = append(channelIDs, c.ID)
	}

	if len(channels) == 0 {
		return []*domain.ChannelWithMembers{}, nil
	}

	memberRows, err := r.pool.Query(ctx, `
		SELECT
			cm.id,
			cm.user_id,
			u.username,
			cm.joined_at,
			cm.channel_id
		FROM channel_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.channel_id = ANY($1)
		ORDER BY cm.joined_at ASC
	`, channelIDs)
	if err != nil {
		return nil, fmt.Errorf("list channel members: %w", err)
	}
	defer memberRows.Close()

	membersByChannel := make(map[string][]*domain.ChannelMember)

	for memberRows.Next() {
		var channelID string
		m := new(domain.ChannelMember)

		if err := memberRows.Scan(
			&m.ID,
			&m.UserID,
			&m.Username,
			&m.JoinedAt,
			&channelID,
		); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		membersByChannel[channelID] = append(membersByChannel[channelID], m)
	}

	var output []*domain.ChannelWithMembers
	for _, c := range channels {
		output = append(output, &domain.ChannelWithMembers{
			Channel: c,
			Members: membersByChannel[c.ID],
		})
	}

	return output, nil
}

func (r *adminRepo) ListChannelMembers(ctx context.Context, channelID string) ([]*domain.ChannelMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			cm.id,
			cm.user_id,
			u.username,
			cm.joined_at
		FROM channel_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.channel_id = $1
		ORDER BY cm.joined_at ASC
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.ChannelMember
	for rows.Next() {
		m := new(domain.ChannelMember)
		if err := scanMember(rows, m); err != nil {
			return nil, fmt.Errorf("scanMember: %w", err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *adminRepo) FreezeChannel(ctx context.Context, channelID, reason string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE channels SET is_frozen = true WHERE id = $1
	`, channelID)
	return err
}

func (r *adminRepo) UnfreezeChannel(ctx context.Context, channelID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE channels SET is_frozen = false WHERE id = $1
	`, channelID)
	return err
}

// func (r *adminRepo) DeleteChannel(ctx context.Context, channelID string) error {
// 	_, err := r.pool.Exec(ctx, `
// 		DELETE FROM channels WHERE id = $1
// 	`, channelID)
// 	return err
// }

//
// ---------------------------------------------------------
// FILE MANAGEMENT
// ---------------------------------------------------------
//

func (r *adminRepo) ListAllFiles(ctx context.Context) ([]*domain.File, error) {
	rows, err := r.pool.Query(ctx, `
    SELECT 
        f.id,
        f.owner_id,
        f.channel_id,
        f.filename,
        f.size,
        f.mime_type,
        f.storage_path,
        f.is_public,
        f.created_at,
        u.username,
        c.name
    FROM files f
    JOIN users u ON f.owner_id = u.id
    JOIN channels c ON f.channel_id = c.id
`)

	if err != nil {
		return nil, fmt.Errorf("list all files: %w", err)
	}
	defer rows.Close()

	var files []*domain.File
	for rows.Next() {
		f := new(domain.File)
		if err := scanFile(rows, f); err != nil {
			return nil, fmt.Errorf("scanFile: %w", err)
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func (r *adminRepo) DeleteFile(ctx context.Context, fileID string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM files WHERE id = $1
	`, fileID)
	return err
}
