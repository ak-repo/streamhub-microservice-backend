package postgres

import (
	"context"
	"fmt"

	"github.com/ak-repo/stream-hub/internal/chat_service/domain"
	"github.com/ak-repo/stream-hub/internal/chat_service/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type chatRepo struct {
	pool *pgxpool.Pool
}

func NewChatRepo(pool *pgxpool.Pool) port.ChatRepository {
	return &chatRepo{pool: pool}
}

//---------------------------For messages-----------------------

// SaveMessage persists a message to the database.
// This ensures message history is preserved even if Redis crashes.
func (r *chatRepo) SaveMessage(ctx context.Context, msg *domain.Message) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO messages (message_id, channel_id, sender_id, content, created_at) 
         VALUES ($1, $2, $3, $4, $5)`,
		msg.ID, msg.ChannelID, msg.SenderID, msg.Content, msg.CreatedAt)
	return err
}

// ListHistory retrieves message history for a channel.
func (r *chatRepo) ListHistory(ctx context.Context, channelID string, limit, offset int) ([]*domain.Message, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT message_id, channel_id, sender_id, content, created_at
         FROM messages 
         WHERE channel_id = $1 
         ORDER BY created_at DESC 
         LIMIT $2 OFFSET $3`,
		channelID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []*domain.Message{}
	for rows.Next() {
		msg := &domain.Message{}
		if err := rows.Scan(&msg.ID, &msg.ChannelID, &msg.SenderID, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

// ------------------------------ For channels __--------------

// CreateChannel creates a new chat group/channel.
func (r *chatRepo) CreateChannel(ctx context.Context, ch *domain.Channel) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO channels (channel_id, name, created_by, created_at)
         VALUES ($1, $2, $3, $4)`,
		ch.ID, ch.Name, ch.CreatedBy, ch.CreatedAt)
	return err
}

// GetChannel retrieves channel details by ID.
func (r *chatRepo) GetChannel(ctx context.Context, channelID string) (*domain.Channel, error) {
	ch := &domain.Channel{}
	err := r.pool.QueryRow(ctx, `SELECT channel_id, name, created_by, created_at 
         FROM channels 
         WHERE channel_id = $1`,
		channelID).Scan(&ch.ID, &ch.Name, &ch.CreatedBy, &ch.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("channel not found: %s", channelID)

	}
	return ch, err
}

// AddMember adds a user to a channel.
// This enables permission control - only members can see/send messages.
func (r *chatRepo) AddMember(ctx context.Context, m *domain.ChannelMember) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO channel_members (id, channel_id, user_id, joined_at)
         VALUES ($1, $2, $3, $4)
         ON CONFLICT (channel_id, user_id) DO NOTHING`,
		m.ID, m.ChannelID, m.UserID, m.JoinedAt)
	return err
}

// RemoveMember removes a user from a channel.
func (r *chatRepo) RemoveMember(ctx context.Context, channelID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM channel_members 
         WHERE channel_id = $1 AND user_id = $2`,
		channelID, userID)
	return err
}

// ListChannelMembers returns all members of a channel.
func (r *chatRepo) ListChannelMembers(ctx context.Context, channelID string) ([]*domain.ChannelMember, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, channel_id, user_id, joined_at
         FROM channel_members
         WHERE channel_id = $1
         ORDER BY joined_at ASC`,
		channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.ChannelMember
	for rows.Next() {
		m := &domain.ChannelMember{}
		if err := rows.Scan(&m.ID, &m.ChannelID, &m.UserID, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// IsUserMember checks if a user is a member of a channel.
// Used for authorization before allowing message posting.
func (r *chatRepo) IsUserMember(ctx context.Context, channelID, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
            SELECT 1 FROM channel_members 
            WHERE channel_id = $1 AND user_id = $2
        )`,
		channelID, userID).Scan(&exists)
	return exists, err
}
