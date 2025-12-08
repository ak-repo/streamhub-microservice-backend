package app

import (
	"github.com/ak-repo/stream-hub/internal/admin_service/port"
	"github.com/ak-repo/stream-hub/pkg/grpc/clients"
)

type adminService struct {
	repo    port.AdminRepository
	clients *clients.Clients
}
