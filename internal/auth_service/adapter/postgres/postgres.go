package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
	"github.com/ak-repo/stream-hub/internal/auth_service/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// userRepo implements port.UserRepository
type userRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new Postgres user repository
func NewUserRepository(pool *pgxpool.Pool) port.UserRepository {
	return &userRepo{pool: pool}
}

// -----------------------------------------------------------------------------
// INTERNAL HELPERS
// -----------------------------------------------------------------------------

// commonUserColumns defines the standard columns to select for a User.
// NOTE: Ensure order matches the scanUser helper below.
const commonUserColumns = `
	id, 
	username, 
	email, 
	password_hash, 
	role, 
	email_verified, 
	is_banned, 
	upload_blocked, 
	avatar_url, 
	created_at, 
	updated_at
`

// scanUser maps a pgx row to a domain.User struct.
// It assumes the query selected columns in the exact order of commonUserColumns.
func scanUser(row pgx.Row, u *domain.User) error {
	return row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.EmailVerified,
		&u.IsBanned,
		&u.UploadBlocked,
		&u.AvatarURL,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
}

// -----------------------------------------------------------------------------
// CREATE & UPDATE
// -----------------------------------------------------------------------------

func (r *userRepo) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (
			id, username, email, password_hash, role,
			email_verified, is_banned, upload_blocked, avatar_url
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at, avatar_url
	`
	return r.pool.QueryRow(
		ctx,
		query,
		u.ID,
		u.Username,
		u.Email,
		u.PasswordHash,
		u.Role,
		u.EmailVerified,
		u.IsBanned,
		u.UploadBlocked,
		u.AvatarURL,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt, &u.AvatarURL)
}

func (r *userRepo) Update(ctx context.Context, u *domain.User) error {
	log.Printf("Updating user: %s", u.ID)

	query := `
		UPDATE users 
		SET username=$1, email=$2, password_hash=$3, role=$4, 
			email_verified=$5, is_banned=$6, upload_blocked=$7,
			avatar_url=$8, updated_at=NOW()
		WHERE id=$9
	`
	_, err := r.pool.Exec(ctx,
		query,
		u.Username,
		u.Email,
		u.PasswordHash,
		u.Role,
		u.EmailVerified,
		u.IsBanned,
		u.UploadBlocked,
		u.AvatarURL,
		u.ID,
	)
	return err
}

func (r *userRepo) UpdatePassword(ctx context.Context, email, hash string) error {
	log.Println("email: ", email, " hashL: ", hash)
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE email = $2`
	result, err := r.pool.Exec(ctx, query, hash, email)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no user found with email: %s", email)
	}
	return nil
}

func (r *userRepo) UpdateAvatar(ctx context.Context, userID, url string) error {
	query := `UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, url, userID)
	return err
}

// -----------------------------------------------------------------------------
// READ (Single User)
// -----------------------------------------------------------------------------

func (r *userRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE id=$1", commonUserColumns)
	row := r.pool.QueryRow(ctx, query, id)

	u := &domain.User{}
	if err := scanUser(row, u); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil, nil if not found (idiomatic)
		}
		return nil, fmt.Errorf("FindByID: %w", err)
	}
	return u, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE email=$1", commonUserColumns)
	row := r.pool.QueryRow(ctx, query, email)

	u := &domain.User{}
	if err := scanUser(row, u); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindByEmail: %w", err)
	}
	return u, nil
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE username=$1", commonUserColumns)
	row := r.pool.QueryRow(ctx, query, username)

	u := &domain.User{}
	if err := scanUser(row, u); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindByUsername: %w", err)
	}
	return u, nil
}

// -----------------------------------------------------------------------------
// READ (List / Search)
// -----------------------------------------------------------------------------

// fetchUsers is a generic helper to reduce code duplication for list queries
func (r *userRepo) fetchUsers(ctx context.Context, query string, args ...any) ([]*domain.User, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetchUsers query: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := new(domain.User)
		if err := scanUser(rows, u); err != nil {
			return nil, fmt.Errorf("scanUser error: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepo) SearchUsers(ctx context.Context, filter string) ([]*domain.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users", commonUserColumns)
	var args []interface{}

	if filter != "" {
		query += " WHERE username ILIKE $1 OR email ILIKE $1"
		args = append(args, "%"+filter+"%")
	}

	query += " ORDER BY created_at DESC"
	return r.fetchUsers(ctx, query, args...)
}

func (r *userRepo) ListActiveUsers(ctx context.Context) ([]*domain.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE is_banned = false ORDER BY created_at DESC", commonUserColumns)
	return r.fetchUsers(ctx, query)
}

func (r *userRepo) ListBannedUsers(ctx context.Context) ([]*domain.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE is_banned = true ORDER BY created_at DESC", commonUserColumns)
	return r.fetchUsers(ctx, query)
}

// -----------------------------------------------------------------------------
// ADMIN ACTIONS
// -----------------------------------------------------------------------------

// var total int
// err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
// if err != nil {
// 	return nil, 0, fmt.Errorf("count users: %w", err)
// }

func (r *userRepo) ListUsers(ctx context.Context, filter string, limit, offset int32) ([]*domain.User, int32, error) {
	query := fmt.Sprintf("SELECT %s FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2", commonUserColumns)

	users, err := r.fetchUsers(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	var total int
	err = r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	return users, int32(total), err
}

func (r *userRepo) BanUser(ctx context.Context, id, reason string) error {

	query := `UPDATE users SET is_banned = true, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *userRepo) UnbanUser(ctx context.Context, id, reason string) error {
	query := `UPDATE users SET is_banned = false, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *userRepo) UpdateRole(ctx context.Context, id, role string) error {
	query := `UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, role, id)
	return err
}

func (r *userRepo) SetUserUploadBlocked(ctx context.Context, id string, blocked bool) error {
	query := `UPDATE users SET upload_blocked = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, blocked, id)
	return err
}

func (r *userRepo) DeleteUser(ctx context.Context, id string) error {
	c, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)

	if c.RowsAffected() == 0 {
		return fmt.Errorf("no user found with this user id")
	}
	return err
}
