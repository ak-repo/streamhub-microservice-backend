package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ak-repo/stream-hub/internal/channel_service/domain"
	"github.com/ak-repo/stream-hub/internal/channel_service/port"
)

type channelRepo struct {
	db *pgxpool.Pool
}

func NewChannelRepo(pool *pgxpool.Pool) port.ChannelRepository {
	return &channelRepo{db: pool}
}

//
// CHANNEL CRUD
//

func (r *channelRepo) CreateChannel(ctx context.Context, c *domain.Channel) error {
	const q = `
		INSERT INTO channels (id, name, description, visibility, created_by, created_at, is_frozen)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, q,
		c.ID, c.Name, c.Description, c.Visibility,
		c.OwnerID, c.CreatedAt, c.IsFrozen,
	)
	return err
}

func (r *channelRepo) GetChannel(ctx context.Context, channelID string) (*domain.Channel, error) {
	const q = `
		SELECT id, name, description, visibility, created_by, created_at, is_frozen
		FROM channels
		WHERE id = $1
	`

	ch := &domain.Channel{}
	err := r.db.QueryRow(ctx, q, channelID).Scan(
		&ch.ID, &ch.Name, &ch.Description, &ch.Visibility,
		&ch.OwnerID, &ch.CreatedAt, &ch.IsFrozen,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}

	return ch, nil
}

func (r *channelRepo) DeleteChannel(ctx context.Context, channelID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM channels WHERE id = $1`, channelID)
	return err
}

func (r *channelRepo) ListUserChannels(ctx context.Context, userID string) ([]*domain.Channel, error) {
	const q = `
		SELECT c.id, c.name, c.description, c.visibility, c.created_by, c.created_at, c.is_frozen
		FROM channels c
		JOIN channel_members cm ON cm.channel_id = c.id
		WHERE cm.user_id = $1
		ORDER BY c.created_at DESC
	`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Channel
	for rows.Next() {
		ch := new(domain.Channel)
		if err := rows.Scan(
			&ch.ID, &ch.Name, &ch.Description, &ch.Visibility,
			&ch.OwnerID, &ch.CreatedAt, &ch.IsFrozen,
		); err != nil {
			return nil, err
		}
		list = append(list, ch)
	}

	return list, nil
}

func (r *channelRepo) SearchChannels(ctx context.Context, filter string, limit, offset int32) ([]*domain.Channel, error) {
	log.Println("qu: ", filter, "limit", limit, "off:", offset)
	query := `
		SELECT id, name, description, visibility, created_by, created_at, is_frozen
		FROM channels WHERE visibility = 'public'
	`
	var args []interface{}
	argIndex := 1

	// Optional filter
	if filter != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argIndex)
		args = append(args, "%"+filter+"%")
		argIndex++
	}

	// Sorting
	query += " ORDER BY created_at DESC"

	// Pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*domain.Channel
	for rows.Next() {
		ch := &domain.Channel{}
		err := rows.Scan(&ch.ID, &ch.Name, &ch.Description, &ch.Visibility, &ch.OwnerID, &ch.CreatedAt, &ch.IsFrozen)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}

	return channels, nil
}

//
// MEMBER MANAGEMENT
//

func (r *channelRepo) AddMember(ctx context.Context, m *domain.ChannelMember) error {
	const q = `
		INSERT INTO channel_members (id, channel_id, user_id, joined_at, role)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (channel_id, user_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, q, m.ID, m.ChannelID, m.UserID, m.JoinedAt, m.Role)
	return err
}

func (r *channelRepo) RemoveMember(ctx context.Context, channelID, userID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM channel_members WHERE channel_id = $1 AND user_id = $2`,
		channelID, userID,
	)
	return err
}

func (r *channelRepo) IsUserMember(ctx context.Context, channelID, userID string) (bool, error) {
	const q = `
		SELECT EXISTS(
			SELECT 1 FROM channel_members
			WHERE channel_id = $1 AND user_id = $2
		)
	`
	var exists bool
	err := r.db.QueryRow(ctx, q, channelID, userID).Scan(&exists)
	return exists, err
}

func (r *channelRepo) ListChannelMembers(ctx context.Context, channelID string) ([]*domain.ChannelMember, error) {
	const q = `
		SELECT cm.channel_id, cm.user_id, u.username, cm.role, cm.joined_at
		FROM channel_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.channel_id = $1
		ORDER BY cm.joined_at ASC
	`

	rows, err := r.db.Query(ctx, q, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.ChannelMember
	for rows.Next() {
		m := new(domain.ChannelMember)
		if err := rows.Scan(
			&m.ChannelID,
			&m.UserID,
			&m.Username,
			&m.Role,
			&m.JoinedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

//
// MESSAGING
//

func (r *channelRepo) SaveMessage(ctx context.Context, msg *domain.Message) error {
	const q = `
		INSERT INTO messages (id, channel_id, sender_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, q,
		msg.ID, msg.ChannelID, msg.SenderID,
		msg.Content, msg.CreatedAt,
	)
	return err
}

func (r *channelRepo) ListHistory(ctx context.Context, channelID string, limit, offset int) ([]*domain.Message, error) {
	const q = `
		SELECT m.id, m.channel_id, m.sender_id, u.username, m.content, m.created_at
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.channel_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, q, channelID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Message
	for rows.Next() {
		msg := new(domain.Message)
		if err := rows.Scan(
			&msg.ID,
			&msg.ChannelID,
			&msg.SenderID,
			&msg.Username,
			&msg.Content,
			&msg.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, msg)
	}

	return list, nil
}

//
// REQUESTS
//

func (r *channelRepo) CreateRequest(ctx context.Context, req *domain.Request) error {
	const q = `
		INSERT INTO requests (id, user_id, channel_id, type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, q,
		req.ID, req.UserID, req.ChannelID,
		req.Type, req.Status, req.CreatedAt,
	)
	return err
}
func (r *channelRepo) UpdateRequestStatus(ctx context.Context, requestID, status string) (*domain.Request, error) {
	_, err := r.db.Exec(ctx,
		`UPDATE requests SET status = $1 WHERE id = $2`,
		status, requestID,
	)
	if err != nil {
		return nil, err
	}

	// Fetch updated request
	row := r.db.QueryRow(ctx,
		`SELECT id, channel_id, user_id, type, status, created_at 
		 FROM requests 
		 WHERE id = $1`, requestID)

	var req domain.Request
	if err := row.Scan(
		&req.ID,
		&req.ChannelID,
		&req.UserID,
		&req.Type,
		&req.Status,
		&req.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *channelRepo) ListPendingRequests(ctx context.Context, userID, channelID string) ([]*domain.Request, error) {
	q := `
		SELECT id, user_id, channel_id, type, status, created_at
		FROM requests
		WHERE status = 'pending'
	`
	var args []any
	i := 1

	if userID != "" {
		q += fmt.Sprintf(" AND user_id = $%d", i)
		args = append(args, userID)
		i++
	}

	if channelID != "" {
		q += fmt.Sprintf(" AND channel_id = $%d", i)
		args = append(args, channelID)
		i++
	}

	q += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Request
	for rows.Next() {
		req := new(domain.Request)
		if err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.ChannelID,
			&req.Type,
			&req.Status,
			&req.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, req)
	}

	return list, nil
}

//
// ADMIN
//

func (r *channelRepo) AdminListChannels(ctx context.Context, limit, offset int32) ([]*domain.Channel, error) {
	const q = `
		SELECT id, name, description, visibility, created_by, created_at, is_frozen
		FROM channels
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Channel
	for rows.Next() {
		c := new(domain.Channel)
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Description, &c.Visibility,
			&c.OwnerID, &c.CreatedAt, &c.IsFrozen,
		); err != nil {
			return nil, err
		}
		list = append(list, c)
	}

	return list, nil
}

func (r *channelRepo) FreezeChannel(ctx context.Context, channelID string, freeze bool) error {
	_, err := r.db.Exec(ctx,
		`UPDATE channels SET is_frozen = $1 WHERE id = $2`,
		freeze, channelID,
	)
	return err
}
