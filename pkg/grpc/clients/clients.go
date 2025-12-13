package clients

import (
	"fmt"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/adminpb"
	"github.com/ak-repo/stream-hub/gen/authpb"
	"github.com/ak-repo/stream-hub/gen/channelpb"
	"github.com/ak-repo/stream-hub/gen/filespb"
	"github.com/ak-repo/stream-hub/gen/paymentpb"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Admin        adminpb.AdminServiceClient
	Auth         authpb.AuthServiceClient
	AdminAuth    authpb.AdminAuthServiceClient
	File         filespb.FileServiceClient
	AdminFile    filespb.AdminFileServiceClient
	Channel      channelpb.ChannelServiceClient
	AdminChannel channelpb.AdminChannelServiceClient
	Payment      paymentpb.PaymentServiceClient

	conns []*grpc.ClientConn
}

// Generic gRPC client initializer
func initClient[T any](host, port string, factory func(*grpc.ClientConn) T) (T, *grpc.ClientConn) {
	addr := fmt.Sprintf("%s:%s", host, port)

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatal("failed to dial gRPC service",
			zap.String("address", addr),
			zap.Error(err),
		)
	}

	return factory(conn), conn
}

// Create a full client container
func NewClients(cfg *config.Config) *Clients {

	c := &Clients{
		conns: make([]*grpc.ClientConn, 0),
	}

	// Admin Service (Generic Admin)
	adminClient, adminConn := initClient(cfg.Services.Admin.Host, cfg.Services.Admin.Port, func(conn *grpc.ClientConn) adminpb.AdminServiceClient {
		return adminpb.NewAdminServiceClient(conn)
	})
	c.Admin = adminClient
	c.conns = append(c.conns, adminConn)

	// Auth Service (User)
	authClient, authConn := initClient(cfg.Services.Auth.Host, cfg.Services.Auth.Port,
		func(conn *grpc.ClientConn) authpb.AuthServiceClient {
			return authpb.NewAuthServiceClient(conn)
		},
	)
	c.Auth = authClient
	c.conns = append(c.conns, authConn)

	// Auth Service (Admin) - Shares connection with User Auth
	adminAuthClient := initAuthAdminClient(authConn)
	c.AdminAuth = adminAuthClient

	// File Service (User)
	fileClient, fileConn := initClient(cfg.Services.File.Host, cfg.Services.File.Port,
		func(conn *grpc.ClientConn) filespb.FileServiceClient {
			return filespb.NewFileServiceClient(conn)
		},
	)
	c.File = fileClient
	c.conns = append(c.conns, fileConn)

	// File Service (Admin) - Shares connection with User File
	adminFileClient := initFileAdminClient(fileConn)
	c.AdminFile = adminFileClient

	// Chat / Channel Service (User)
	chatClient, chatConn := initClient(cfg.Services.Chat.Host, cfg.Services.Chat.Port,
		func(conn *grpc.ClientConn) channelpb.ChannelServiceClient {
			return channelpb.NewChannelServiceClient(conn)
		},
	)
	c.Channel = chatClient
	c.conns = append(c.conns, chatConn)

	// Chat / Channel Service (Admin) - Shares connection with User Channel
	adminChannelClient := initChannelAdminClient(chatConn)
	c.AdminChannel = adminChannelClient

	// Payment service
	paymentClient, payConn := initClient(cfg.Services.Payment.Host, cfg.Services.Payment.Port, func(conn *grpc.ClientConn) paymentpb.PaymentServiceClient {
		return paymentpb.NewPaymentServiceClient(conn)
	})
	c.Payment = paymentClient
	c.conns = append(c.conns, payConn)

	return c
}

// Helper factory for AdminAuthServiceClient
func initAuthAdminClient(conn *grpc.ClientConn) authpb.AdminAuthServiceClient {
	return authpb.NewAdminAuthServiceClient(conn)
}

// Helper factory for AdminFileServiceClient
func initFileAdminClient(conn *grpc.ClientConn) filespb.AdminFileServiceClient {
	return filespb.NewAdminFileServiceClient(conn)
}

// Helper factory for AdminChannelServiceClient
func initChannelAdminClient(conn *grpc.ClientConn) channelpb.AdminChannelServiceClient {
	return channelpb.NewAdminChannelServiceClient(conn)
}

// Gracefully close all gRPC connections
func (c *Clients) CloseAll() {
	for _, conn := range c.conns {
		_ = conn.Close()
	}
}
