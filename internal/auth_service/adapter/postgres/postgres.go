package postgres

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
	"github.com/ak-repo/stream-hub/internal/auth_service/port"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) port.UserRepository {
	return &userRepo{pool: pool}
}

// Create User
func (r *userRepo) Create(ctx context.Context, u *domain.User) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash, role, email_verified, is_banned)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 RETURNING id, created_at, updated_at`,
		u.Username, u.Email, u.PasswordHash, u.Role, u.EmailVerified, u.IsBanned,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// ExistsByEmail
func (r *userRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, email,
	).Scan(&exists)
	return exists, err
}

// FindByEmail
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, role, email_verified, is_banned, created_at, updated_at
		 FROM users WHERE email=$1`, email,
	)

	u := &domain.User{}
	if err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.Role, &u.EmailVerified, &u.IsBanned,
		&u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return u, nil
}

// FindByID
func (r *userRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, role, email_verified, is_banned, created_at, updated_at
		 FROM users WHERE id=$1`, id,
	)

	u := &domain.User{}
	if err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.Role, &u.EmailVerified, &u.IsBanned,
		&u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return u, nil
}

// Update User
// ----------------------
func (r *userRepo) Update(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users 
		 SET username=$1, email=$2, password_hash=$3, role=$4, email_verified=$5, 
		     is_banned=$6, updated_at=$7
		 WHERE id=$8`,
		u.Username, u.Email, u.PasswordHash, u.Role,
		u.EmailVerified, u.IsBanned, u.UpdatedAt, u.ID,
	)
	return err
}
