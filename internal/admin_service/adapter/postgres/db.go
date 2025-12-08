package postgres

import (
	"github.com/ak-repo/stream-hub/internal/admin_service/port"
	"github.com/jackc/pgx/v5/pgxpool"
)

type adminRepo struct {
	pool *pgxpool.Pool
}

func NewAdminRepo(pool *pgxpool.Pool) port.AdminRepository {
	return &adminRepo{pool: pool}
}
