package clients

import (
	"fmt"

	"github.com/ak-repo/stream-hub/gen/authpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Auth  authpb.AuthServiceClient
	conns []*grpc.ClientConn
}

func NewContainer() *Clients {
	return &Clients{
		conns: make([]*grpc.ClientConn, 0),
	}
}

func (c *Clients) AddConn(conn *grpc.ClientConn) {
	c.conns = append(c.conns, conn)
}

func (c *Clients) CloseAll() {
	for _, conn := range c.conns {
		conn.Close()
	}
}

func NewClient(host, port string, factory func(*grpc.ClientConn) interface{}) (interface{}, *grpc.ClientConn, error) {
	addr := fmt.Sprintf("%s:%s", host, port)

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial gRPC service at %s: %w", addr, err)
	}

	return factory(conn), conn, nil
}
