package clients

import (
	"fmt"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/authpb"
	"github.com/ak-repo/stream-hub/gen/channelpb"
	"github.com/ak-repo/stream-hub/gen/filespb"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Auth    authpb.AuthServiceClient ``
	File    filespb.FileServiceClient
	Channel channelpb.ChannelServiceClient
	conns   []*grpc.ClientConn
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

	// Auth Service
	authClient, authConn := initClient(cfg.Services.Auth.Host, cfg.Services.Auth.Port,
		func(conn *grpc.ClientConn) authpb.AuthServiceClient {
			return authpb.NewAuthServiceClient(conn)
		},
	)
	c.Auth = authClient
	c.conns = append(c.conns, authConn)

	// File Service
	fileClient, fileConn := initClient(cfg.Services.File.Host, cfg.Services.File.Port,
		func(conn *grpc.ClientConn) filespb.FileServiceClient {
			return filespb.NewFileServiceClient(conn)
		},
	)
	c.File = fileClient
	c.conns = append(c.conns, fileConn)

	// Chat / Channel Service
	chatClient, chatConn := initClient(cfg.Services.Chat.Host, cfg.Services.Chat.Port,
		func(conn *grpc.ClientConn) channelpb.ChannelServiceClient {
			return channelpb.NewChannelServiceClient(conn)
		},
	)
	c.Channel = chatClient
	c.conns = append(c.conns, chatConn)

	return c

}

// Gracefully close all gRPC connections
func (c *Clients) CloseAll() {
	for _, conn := range c.conns {
		_ = conn.Close()
	}
}
