package admingrpc

import (
	"github.com/ak-repo/stream-hub/gen/adminpb"
	"github.com/ak-repo/stream-hub/internal/admin_service/port"
)

type AdminServer struct {
	adminpb.UnimplementedAdminServiceServer
	service port.AdminService
}

func NewAdminServer(service port.AdminService) *AdminServer {
	return &AdminServer{service: service}
}
