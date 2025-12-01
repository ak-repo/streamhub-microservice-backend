package seeder

import (
	"context"

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
