package postgres

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/admin_service/domain"
	"github.com/ak-repo/stream-hub/internal/admin_service/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type adminRepo struct {
	db *pgxpool.Pool
}

func NewAdminRepo(db *pgxpool.Pool) port.AdminRepository {
	return &adminRepo{db: db}
}

// ---------------------------------------------------------
// Helpers
// ---------------------------------------------------------

const userColumns = `
    id, username, email, password_hash, role, 
    created_at, updated_at, email_verified, is_banned
`

func scanUser(row pgx.Row, u *domain.User) error {
	return row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.EmailVerified,
		&u.IsBanned,
	)
}

func (r *adminRepo) fetchUsers(ctx context.Context, query string, args ...any) ([]*domain.User, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := new(domain.User)
		if err := scanUser(rows, u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// ---------------------------------------------------------
// Queries
// ---------------------------------------------------------

func (r *adminRepo) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return r.fetchUsers(ctx, `SELECT `+userColumns+` FROM users`)
}

func (r *adminRepo) ListActiveUsers(ctx context.Context) ([]*domain.User, error) {
	return r.fetchUsers(ctx, `SELECT `+userColumns+` FROM users WHERE is_banned = false`)
}

func (r *adminRepo) ListBannedUsers(ctx context.Context) ([]*domain.User, error) {
	return r.fetchUsers(ctx, `SELECT `+userColumns+` FROM users WHERE is_banned = true`)
}

func (r *adminRepo) BanUser(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET is_banned = true WHERE id = $1`,
		userID,
	)
	return err
}

func (r *adminRepo) UnbanUser(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET is_banned = false WHERE id = $1`,
		userID,
	)
	return err
}

func (r *adminRepo) UpdateRole(ctx context.Context, userID, role string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET role = $1 WHERE id = $2`,
		role, userID,
	)
	return err
}
