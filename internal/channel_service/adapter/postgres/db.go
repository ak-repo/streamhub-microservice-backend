package postgres

import (
	"context"
	"fmt"

	"github.com/ak-repo/stream-hub/internal/channel_service/domain"
	"github.com/ak-repo/stream-hub/internal/channel_service/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type chatRepo struct {
	pool *pgxpool.Pool
}

func NewChatRepo(pool *pgxpool.Pool) port.ChannelRepository {
	return &chatRepo{pool: pool}
}

//---------------------------For messages-----------------------

// SaveMessage persists a message to the database.
// This ensures message history is preserved even if Redis crashes.
func (r *chatRepo) SaveMessage(ctx context.Context, msg *domain.Message) error {
	var attachmentID *string
	if msg.Attachment != nil {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO file_attachments (id, file_name, file_url, mime_type, size, uploaded_by, uploaded_at)
             VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			msg.Attachment.ID, msg.Attachment.FileName, msg.Attachment.URL,
			msg.Attachment.MimeType, msg.Attachment.Size, msg.SenderID, msg.CreatedAt)
		if err != nil {
			return err
		}
		attachmentID = &msg.Attachment.ID
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO messages (id, channel_id, sender_id, content, created_at, attachment_id)
         VALUES ($1,$2,$3,$4,$5,$6)`,
		msg.ID, msg.ChannelID, msg.SenderID, msg.Content, msg.CreatedAt, attachmentID)
	return err
}

// ListHistory retrieves message history for a channel.
// ListHistory retrieves message history for a channel.
func (r *chatRepo) ListHistory(ctx context.Context, channelID string, limit, offset int) ([]*domain.Message, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT m.id, m.channel_id, m.sender_id, m.content, m.created_at, u.username,
		        f.id, f.file_name, f.file_url, f.mime_type, f.size
         FROM messages m
         JOIN users u ON m.sender_id = u.id
         LEFT JOIN file_attachments f ON m.attachment_id = f.id
         WHERE m.channel_id = $1
         ORDER BY m.created_at ASC
         LIMIT $2 OFFSET $3`,
		channelID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}
		var attachmentID, fileName, fileURL, mimeType *string
		var size *int64

		if err := rows.Scan(&msg.ID, &msg.ChannelID, &msg.SenderID, &msg.Content, &msg.CreatedAt,
			&msg.Username, &attachmentID, &fileName, &fileURL, &mimeType, &size); err != nil {
			return nil, err
		}

		if attachmentID != nil {
			msg.Attachment = &domain.FileAttachment{
				ID:       *attachmentID,
				FileName: *fileName,
				URL:      *fileURL,
				MimeType: *mimeType,
				Size:     *size,
			}
		}

		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// ------------------------------ For channels __--------------

// CreateChannel creates a new chat group/channel.
func (r *chatRepo) CreateChannel(ctx context.Context, ch *domain.Channel) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO channels (id, name, created_by, created_at)
         VALUES ($1, $2, $3, $4)`,
		ch.ID, ch.Name, ch.CreatedBy, ch.CreatedAt)
	return err
}

// GetChannel retrieves channel details by ID.
func (r *chatRepo) GetChannel(ctx context.Context, channelID string) (*domain.Channel, error) {
	ch := &domain.Channel{}
	err := r.pool.QueryRow(ctx, `SELECT id, name, created_by, created_at 
         FROM channels 
         WHERE id = $1`,
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
		`SELECT 
            cm.id,
            cm.channel_id,
            cm.user_id,
            u.username,
            cm.joined_at
         FROM channel_members cm
         JOIN users u ON cm.user_id = u.id
         WHERE cm.channel_id = $1
         ORDER BY cm.joined_at ASC`,
		channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.ChannelMember
	for rows.Next() {
		m := &domain.ChannelMember{}
		if err := rows.Scan(&m.ID, &m.ChannelID, &m.UserID, &m.Username, &m.JoinedAt); err != nil {
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

// List all channels to a specific user
// List all channels to a specific user
func (r *chatRepo) ListChannels(ctx context.Context, userID string) (map[string]*domain.ChannelWithMembers, error) {

	// STEP 1: get channel IDs (FIXED)
	rows, err := r.pool.Query(ctx,
		`SELECT channel_id FROM channel_members WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channelIDs []string
	for rows.Next() {
		var cid string
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		channelIDs = append(channelIDs, cid)
	}

	if len(channelIDs) == 0 {
		return map[string]*domain.ChannelWithMembers{}, nil
	}

	// Final output map
	result := make(map[string]*domain.ChannelWithMembers)

	// STEP 2: get channels
	chanRows, err := r.pool.Query(ctx,
		`SELECT id, name, created_by, created_at
         FROM channels
         WHERE id = ANY($1)`, channelIDs)
	if err != nil {
		return nil, err
	}
	defer chanRows.Close()

	for chanRows.Next() {
		ch := &domain.Channel{}
		if err := chanRows.Scan(&ch.ID, &ch.Name, &ch.CreatedBy, &ch.CreatedAt); err != nil {
			return nil, err
		}

		result[ch.ID] = &domain.ChannelWithMembers{
			Channel: ch,
			Members: []*domain.ChannelMember{},
		}
	}

	// STEP 3: get members for all channels
	memRows, err := r.pool.Query(ctx,
		`SELECT id, channel_id, user_id, joined_at
         FROM channel_members
         WHERE channel_id = ANY($1)`, channelIDs)
	if err != nil {
		return nil, err
	}
	defer memRows.Close()

	for memRows.Next() {
		m := &domain.ChannelMember{}
		if err := memRows.Scan(&m.ID, &m.ChannelID, &m.UserID, &m.JoinedAt); err != nil {
			return nil, err
		}

		if group, ok := result[m.ChannelID]; ok {
			group.Members = append(group.Members, m)
		}
	}

	return result, nil
}

func (r *chatRepo) DeleteChannel(ctx context.Context, channelID, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM channels WHERE id = $1 AND created_by=$2`, channelID, userID)
	return err
}
