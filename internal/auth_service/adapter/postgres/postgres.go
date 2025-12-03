package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
	"github.com/ak-repo/stream-hub/internal/auth_service/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) port.UserRepository {
	return &userRepo{pool: pool}
}

// -------------------------------------------------------------
// CREATE
// -------------------------------------------------------------
func (r *userRepo) Create(ctx context.Context, u *domain.User) error {
	return r.pool.QueryRow(
		ctx,
		`INSERT INTO users (id, username, email, password_hash, role, 
			email_verified, is_banned, upload_blocked)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING id, created_at, updated_at`,
		u.ID, u.Username, u.Email, u.PasswordHash, u.Role,
		u.EmailVerified, u.IsBanned, u.UploadBlocked,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// -------------------------------------------------------------
// FIND BY EMAIL
// -------------------------------------------------------------
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, role, email_verified,
		        is_banned, created_at, updated_at, upload_blocked
		 FROM users WHERE email=$1`, email,
	)

	u := &domain.User{}
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.Role, &u.EmailVerified, &u.IsBanned,
		&u.CreatedAt, &u.UpdatedAt, &u.UploadBlocked,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// -------------------------------------------------------------
// FIND BY USERNAME
// -------------------------------------------------------------
func (r *userRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, role, email_verified,
		        is_banned, created_at, updated_at, upload_blocked
		 FROM users WHERE username=$1`, username,
	)

	u := &domain.User{}
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.Role, &u.EmailVerified, &u.IsBanned,
		&u.CreatedAt, &u.UpdatedAt, &u.UploadBlocked,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// -------------------------------------------------------------
// FIND BY ID
// -------------------------------------------------------------
func (r *userRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, role, email_verified,
		        is_banned, created_at, updated_at, upload_blocked
		 FROM users WHERE id=$1`, id,
	)

	u := &domain.User{}
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.Role, &u.EmailVerified, &u.IsBanned,
		&u.CreatedAt, &u.UpdatedAt, &u.UploadBlocked,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// -------------------------------------------------------------
// UPDATE
// -------------------------------------------------------------
func (r *userRepo) Update(ctx context.Context, u *domain.User) error {
	log.Println("Updating user:", u.ID)

	_, err := r.pool.Exec(ctx,
		`UPDATE users 
		 SET username=$1, email=$2, password_hash=$3, role=$4, 
		     email_verified=$5, is_banned=$6, upload_blocked=$7,
		     updated_at=$8
		 WHERE id=$9`,
		u.Username, u.Email, u.PasswordHash, u.Role,
		u.EmailVerified, u.IsBanned, u.UploadBlocked,
		u.UpdatedAt, u.ID,
	)
	return err
}

// -------------------------------------------------------------
// UPDATE PASSWORD
// -------------------------------------------------------------
func (r *userRepo) UpdatePassword(ctx context.Context, email, hash string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE users 
		 SET password_hash = $1, updated_at = NOW() 
		 WHERE email = $2`,
		hash, email,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no user found with email: %s", email)
	}

	return nil
}
