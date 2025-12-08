package seeder

import (
	"context"
	"time"

	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AdminSeeder(ctx context.Context, pool *pgxpool.Pool) error {

	id := uuid.New().String()
	hash, err := utils.HashPassword("1234")
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, "INSERT INTO USERS (id,username,email,password_hash,role,email_verified) VALUES($1,$2,$3,$4,$5,$6)", id, "super-admin", "admin@hub.com", hash, "super-admin", true)
	return err
}

func UsersSeeder(ctx context.Context, pool *pgxpool.Pool) error {

	seedUsers := []struct {
		Username string
		Email    string
		Password string
	}{
		{"user1", "user1@example.com", "1234"},
		{"user2", "user2@example.com", "1234"},
		{"user3", "user3@example.com", "1234"},
		{"user4", "user4@example.com", "1234"},
		{"user5", "user5@example.com", "1234"},
		{"user6", "user6@example.com", "1234"},
		{"user7", "user7@example.com", "1234"},
		{"user8", "user8@example.com", "1234"},
		{"user9", "user9@example.com", "1234"},
	}

	avatar := "https://res.cloudinary.com/dersnukrf/image/upload/v1764929207/avatars/avatars/profile.jpg.webp"

	for _, u := range seedUsers {

		hash, err := utils.HashPassword(u.Password)
		if err != nil {
			return err
		}

		user := &domain.User{
			ID:            uuid.New().String(),
			Email:         u.Email,
			Username:      u.Username,
			PasswordHash:  hash,
			Role:          "user",
			AvatarURL:     avatar,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			EmailVerified: true,
			IsBanned:      false,
			UploadBlocked: false,
		}

		_, err = pool.Exec(
			ctx,
			`INSERT INTO users 
			 (id, username, email, password_hash, role, avatar_url, created_at, updated_at, email_verified, is_banned, upload_blocked)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			user.ID, user.Username, user.Email, user.PasswordHash, user.Role,
			user.AvatarURL, user.CreatedAt, user.UpdatedAt,
			user.EmailVerified, user.IsBanned, user.UploadBlocked,
		)

		if err != nil {
			return errors.New(errors.CodeInternal, "failed inserting seed user", err)
		}
	}

	return nil
}
